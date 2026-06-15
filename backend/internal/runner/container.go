package runner

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/containerd/errdefs"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
)

// ContainerUIDGID returns the host uid/gid that the Minecraft container should run as.
// When the backend is root we fall back to 1000:1000 so the container does not run as root.
func ContainerUIDGID() (int, int) {
	uid := os.Getuid()
	gid := os.Getgid()
	if uid == 0 {
		slog.Warn("backend running as root; falling back to uid/gid 1000 for minecraft container")
		uid = 1000
		gid = 1000
	}
	return uid, gid
}

// OwnContainerID tries to determine the current container ID from cgroup or mountinfo.
func OwnContainerID() (string, error) {
	// Try cgroup v1 / v2 via /proc/self/cgroup
	data, err := os.ReadFile("/proc/self/cgroup")
	if err == nil {
		for _, line := range strings.Split(string(data), "\n") {
			parts := strings.SplitN(line, ":", 3)
			if len(parts) < 3 {
				continue
			}
			path := strings.TrimSpace(parts[2])
			// cgroup v1: /docker/<id>
			if idx := strings.LastIndex(path, "/docker/"); idx != -1 {
				id := strings.TrimSpace(path[idx+len("/docker/"):])
				if id != "" {
					return id, nil
				}
			}
			// containerd: /containerd/<id>
			if idx := strings.LastIndex(path, "/containerd/"); idx != -1 {
				id := strings.TrimSpace(path[idx+len("/containerd/"):])
				if id != "" {
					return id, nil
				}
			}
			// cgroup v2 with systemd: 0::/system.slice/docker-<id>.scope
			if idx := strings.LastIndex(path, "docker-"); idx != -1 {
				rest := path[idx+len("docker-"):]
				if end := strings.IndexAny(rest, ". \t\n"); end != -1 {
					rest = rest[:end]
				}
				if rest != "" {
					return rest, nil
				}
			}
			// cgroup v2: 0::/docker/<id>
			if idx := strings.LastIndex(path, "/docker/"); idx != -1 {
				id := strings.TrimSpace(path[idx+len("/docker/"):])
				if id != "" {
					return id, nil
				}
			}
		}
	}

	// Fallback: /proc/self/mountinfo overlay path
	data, err = os.ReadFile("/proc/self/mountinfo")
	if err == nil {
		for _, line := range strings.Split(string(data), "\n") {
			if strings.Contains(line, "overlay") && strings.Contains(line, "/var/lib/docker/") {
				parts := strings.Split(line, "/")
				for i, p := range parts {
					if p == "overlay2" || p == "overlay" {
						if i+1 < len(parts) {
							id := strings.TrimSpace(parts[i+1])
							if id != "" && id != "merged" {
								return id, nil
							}
						}
					}
				}
			}
		}
	}

	// Fallback: Docker sets hostname to the container ID (or name) by default.
	if h, err := os.Hostname(); err == nil && h != "" {
		return h, nil
	}

	return "", fmt.Errorf("could not determine container ID; is this running inside Docker?")
}

// OwnNetworkName inspects the current container and returns the name of the first attached network.
func OwnNetworkName(ctx context.Context, cli *client.Client) (string, error) {
	id, err := OwnContainerID()
	if err != nil {
		return "", err
	}

	info, err := cli.ContainerInspect(ctx, id)
	if err != nil {
		return "", fmt.Errorf("failed to inspect own container %s: %w", id, err)
	}

	for name := range info.NetworkSettings.Networks {
		if name != "" {
			return name, nil
		}
	}

	return "", fmt.Errorf("own container %s has no attached networks", id)
}

// ValidateDockerEnvironment checks that the backend is running inside a Docker container
// and returns the name of the network it should use for spawned server containers.
func ValidateDockerEnvironment(ctx context.Context, cli *client.Client) (string, error) {
	if _, err := os.Stat("/.dockerenv"); os.IsNotExist(err) {
		// Missing .dockerenv does not prove we are outside Docker, but OwnNetworkName
		// will fail anyway if we cannot determine the container ID.
	}

	netName, err := OwnNetworkName(ctx, cli)
	if err != nil {
		return "", fmt.Errorf("backend must run inside a Docker container: %w", err)
	}

	return netName, nil
}

// EnsureNetwork creates the Docker bridge network if it does not already exist.
func EnsureNetwork(ctx context.Context, cli *client.Client, netName string) error {
	_, err := cli.NetworkInspect(ctx, netName, network.InspectOptions{})
	if err == nil {
		return nil
	}

	if errdefs.IsNotFound(err) {
		slog.Info("creating docker network", "name", netName)
		_, err = cli.NetworkCreate(ctx, netName, network.CreateOptions{
			Driver: "bridge",
		})
		if err != nil {
			return fmt.Errorf("failed to create network: %w", err)
		}
		slog.Info("docker network created", "name", netName)
		return nil
	}
	return fmt.Errorf("failed to inspect network: %w", err)
}
