package runner

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"github.com/moby/moby/api/pkg/stdcopy"
)

// Clean, neutral constants for the system namespace
const (
	ImageName   = "itzg/minecraft-server:java21"
	NetworkName = "mc-agent-mesh" // Shared isolated network for the agent and containers
)

// StartServerContainer prepares the host environment, cleans up residual dead containers,
// creates a unique servers/<volume_id> directory, builds configuration and runs the container.
// It returns the Docker container ID, the generated volume ID, and the absolute host path of the volume.
// If existingVolumePath is non-empty and the directory exists, it is reused instead of creating a new one.
func StartServerContainer(
	ctx context.Context,
	cli *client.Client,
	serverID string,
	ramBytes int64,
	gamePort uint16,
	engineType string,
	gameVersion string,
	loaderVersion string,
	serversDir string,
	serversHostDir string,
	rconHostPort uint16,
	rconPassword string,
	publicRcon bool,
	existingVolumePath string,
	worldGenEnv map[string]string,
	restartPolicy string,
	networkName string,
	externalJavaArgs []string,
) (string, string, string, error) {
	if err := EnsureNetwork(ctx, cli, networkName); err != nil {
		return "", "", "", err
	}

	containerName := fmt.Sprintf("mc-srv-%s", serverID)

	absServersDir, err := filepath.Abs(serversDir)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to resolve servers directory %s: %w", serversDir, err)
	}

	var volumeID string
	var localPath string
	if existingVolumePath != "" {
		if info, err := os.Stat(existingVolumePath); err == nil && info.IsDir() {
			localPath = existingVolumePath
			volumeID = filepath.Base(localPath)
		}
	}
	if localPath == "" {
		if serverID == "" {
			return "", "", "", fmt.Errorf("server id is required to create volume directory")
		}
		volumeID = serverID
		localPath = filepath.Join(absServersDir, volumeID)
		if mkdirErr := os.MkdirAll(localPath, 0o755); mkdirErr != nil {
			return "", "", "", fmt.Errorf("failed to create server data directory %s: %w", localPath, mkdirErr)
		}
	}

	uid, gid := ContainerUIDGID()
	if err := os.Chown(localPath, uid, gid); err != nil {
		slog.Warn("failed to chown server data directory", "path", localPath, "uid", uid, "gid", gid, "error", err)
	}
	if err := os.Chmod(localPath, 0o775); err != nil {
		slog.Warn("failed to chmod server data directory", "path", localPath, "error", err)
	}

	dockerHostPath := HostPathForDocker(localPath, serversDir, serversHostDir)

	// Leave ~20 % headroom for JVM off-heap / native / container overhead.
	heapBytes := int64(float64(ramBytes) * 0.8)
	memoryVal := fmt.Sprintf("%dM", heapBytes/1024/1024)

	b := NewContainerBuilder(ImageName).
		WithEnv("MEMORY", memoryVal).
		WithEnv("INIT_MEMORY", memoryVal).
		WithEnv("JVM_XX_OPTS", "-XX:+UnlockExperimentalVMOptions").
		WithPort(25565, gamePort, "0.0.0.0").
		WithNetwork(networkName).
		WithEnv("TYPE", engineType).
		WithEnv("VERSION", gameVersion).
		WithEnv("UID", strconv.Itoa(uid)).
		WithEnv("GID", strconv.Itoa(gid)).
		WithEnv("OVERRIDE_SERVER_PROPERTIES", "false").
		WithVolume(dockerHostPath, "/data").
		WithRcon(rconPassword, rconHostPort, publicRcon).
		WithProxyFromEnv().
		WithRestartPolicy(restartPolicy)

	if loaderVersion != "" {
		switch strings.ToUpper(engineType) {
		case "FABRIC":
			b.WithEnv("FABRIC_LOADER_VERSION", loaderVersion)
		case "FORGE":
			b.WithEnv("FORGE_VERSION", loaderVersion)
		}
	}

	if len(externalJavaArgs) > 0 {
		b.WithEnv("JVM_OPTS", strings.Join(externalJavaArgs, " "))
	}

	if worldGenEnv != nil {
		if v := worldGenEnv["level-name"]; v != "" {
			b.WithEnv("LEVEL", v)
		}
		if v := worldGenEnv["level-seed"]; v != "" {
			b.WithEnv("SEED", v)
		}
		if v := worldGenEnv["level-type"]; v != "" {
			b.WithEnv("LEVEL_TYPE", v)
		}
	}

	config, hostConfig, netConfig := b.Build()

	// Check if a container with this name already exists (orphaned or exited).
	existing, err := cli.ContainerInspect(ctx, containerName)
	if err == nil {
		cid := existing.ID
		if existing.State != nil && existing.State.Running {
			slog.Info("container already running", "server_id", serverID, "container_id", cid[:12])
			return cid, volumeID, localPath, nil
		}
		slog.Info("container exists but not running, starting", "server_id", serverID, "container_id", cid[:12])
		if err := cli.ContainerStart(ctx, cid, container.StartOptions{}); err != nil {
			return "", "", "", fmt.Errorf("failed to start existing container: %w", err)
		}
		slog.Info("container started", "server_id", serverID, "container_id", cid[:12])
		return cid, volumeID, localPath, nil
	}
	if !client.IsErrNotFound(err) {
		return "", "", "", fmt.Errorf("failed to inspect container: %w", err)
	}

	resp, err := cli.ContainerCreate(ctx, config, hostConfig, netConfig, nil, containerName)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to create container: %w", err)
	}
	slog.Info("container created", "server_id", serverID, "container_id", resp.ID[:12])

	if err := cli.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		return "", "", "", fmt.Errorf("failed to start container: %w", err)
	}
	slog.Info("container started", "server_id", serverID, "container_id", resp.ID[:12])

	return resp.ID, volumeID, localPath, nil
}

// FixVolumeOwnership runs an ephemeral alpine container to recursively chown a host path
// to the given uid/gid and chmod it to 0o775. This works around permission issues when
// the backend process cannot chown the directory directly.
func FixVolumeOwnership(ctx context.Context, cli *client.Client, hostPath string, uid, gid int) error {
	img := "alpine:latest"
	if _, _, err := cli.ImageInspectWithRaw(ctx, img); err != nil {
		if reader, err := cli.ImagePull(ctx, img, image.PullOptions{}); err != nil {
			slog.Warn("failed to pull ownership-fix image", "image", img, "error", err)
		} else {
			_, _ = io.Copy(io.Discard, reader)
			_ = reader.Close()
		}
	}
	config := &container.Config{
		Image: img,
		Cmd:   []string{"sh", "-c", fmt.Sprintf("chown -R %d:%d /data && chmod 775 /data", uid, gid)},
	}
	hostConfig := &container.HostConfig{
		Binds: []string{fmt.Sprintf("%s:/data", hostPath)},
	}
	resp, err := cli.ContainerCreate(ctx, config, hostConfig, nil, nil, "")
	if err != nil {
		return fmt.Errorf("failed to create ownership-fix container: %w", err)
	}
	if err := cli.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		_ = cli.ContainerRemove(ctx, resp.ID, container.RemoveOptions{Force: true})
		return fmt.Errorf("failed to start ownership-fix container: %w", err)
	}
	statusCh, errCh := cli.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		if err != nil {
			slog.Warn("ownership-fix container wait error", "error", err)
		}
	case <-statusCh:
	}
	if err := cli.ContainerRemove(ctx, resp.ID, container.RemoveOptions{Force: true}); err != nil {
		slog.Warn("failed to remove ownership-fix container", "error", err)
	}
	return nil
}

// StreamContainerLogs streams stdout and stderr from a container to the provided writers.
func StreamContainerLogs(ctx context.Context, cli *client.Client, containerID string, stdout, stderr io.Writer, tailLines int, follow bool) error {
	opts := container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     follow,
	}
	if tailLines >= 0 {
		opts.Tail = strconv.Itoa(tailLines)
	}
	out, err := cli.ContainerLogs(ctx, containerID, opts)
	if err != nil {
		return fmt.Errorf("failed to get container logs stream: %w", err)
	}
	defer func() { _ = out.Close() }()

	_, err = stdcopy.StdCopy(stdout, stderr, out)
	if err != nil && err != io.EOF {
		return fmt.Errorf("error during logs streaming: %w", err)
	}
	return nil
}
