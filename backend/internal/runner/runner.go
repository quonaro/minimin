package runner

import (
	"context"
	"crypto/rand"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/containerd/errdefs"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/moby/moby/api/pkg/stdcopy"
)

// Clean, neutral constants for the system namespace
const (
	ImageName   = "itzg/minecraft-server:latest"
	NetworkName = "mc-agent-mesh" // Shared isolated network for the agent and containers
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
	// Docker's ContainerInspect accepts names, short IDs and long IDs.
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
	// Quick heuristic for older Docker runtimes
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

// ContainerConfigBuilder assembles Docker container configuration via a fluent API.
type ContainerConfigBuilder struct {
	image         string
	env           []string
	labels        map[string]string
	portBindings  nat.PortMap
	exposedPorts  nat.PortSet
	bindMounts    []string
	networkName   string
	restartPolicy string
}

// NewContainerBuilder starts building a Docker container config for the Minecraft server image.
func NewContainerBuilder(imageName string) *ContainerConfigBuilder {
	return &ContainerConfigBuilder{
		image:        imageName,
		labels:       make(map[string]string),
		portBindings: make(nat.PortMap),
		exposedPorts: make(nat.PortSet),
		env:          []string{"EULA=TRUE"},
	}
}

// WithPort exposes an internal container port and binds it to an external host port.
func (b *ContainerConfigBuilder) WithPort(internalPort, externalPort uint16, hostIP string) *ContainerConfigBuilder {
	containerPort := nat.Port(fmt.Sprintf("%d/tcp", internalPort))
	binding := nat.PortBinding{HostPort: fmt.Sprintf("%d", externalPort)}
	if hostIP != "" {
		binding.HostIP = hostIP
	}
	b.exposedPorts[containerPort] = struct{}{}
	b.portBindings[containerPort] = append(b.portBindings[containerPort], binding)
	return b
}

// WithVolume mounts a host directory into the container at the given path.
func (b *ContainerConfigBuilder) WithVolume(hostPath, containerPath string) *ContainerConfigBuilder {
	b.bindMounts = append(b.bindMounts, fmt.Sprintf("%s:%s", hostPath, containerPath))
	return b
}

// WithNetwork attaches the container to the specified Docker network.
func (b *ContainerConfigBuilder) WithNetwork(netName string) *ContainerConfigBuilder {
	b.networkName = netName
	return b
}

// WithLabel adds a Docker label to the container.
func (b *ContainerConfigBuilder) WithLabel(key, value string) *ContainerConfigBuilder {
	b.labels[key] = value
	return b
}

// WithEnv adds an environment variable to the container config.
func (b *ContainerConfigBuilder) WithEnv(key, value string) *ContainerConfigBuilder {
	b.env = append(b.env, fmt.Sprintf("%s=%s", key, value))
	return b
}

// WithRestartPolicy sets the Docker restart policy for the container.
func (b *ContainerConfigBuilder) WithRestartPolicy(name string) *ContainerConfigBuilder {
	b.restartPolicy = name
	return b
}

// WithRcon enables RCON on port 25575 inside the container and binds it to hostPort.
// If public is true the binding uses 0.0.0.0; otherwise it binds to 127.0.0.1.
func (b *ContainerConfigBuilder) WithRcon(password string, hostPort uint16, public bool) *ContainerConfigBuilder {
	b.WithEnv("ENABLE_RCON", "true")
	b.WithEnv("RCON_PORT", "25575")
	b.WithEnv("RCON_PASSWORD", password)
	hostIP := "127.0.0.1"
	if public {
		hostIP = "0.0.0.0"
	}
	b.WithPort(25575, hostPort, hostIP)
	return b
}

// WithProxyFromEnv automatically injects proxy variables if they are present on the host.
// If MC_AGENT_PROXY is set, it is parsed and used to populate HTTP_PROXY, HTTPS_PROXY
// and JAVA_TOOL_OPTIONS automatically.
func (b *ContainerConfigBuilder) WithProxyFromEnv() *ContainerConfigBuilder {
	proxyVars := []string{
		"HTTP_PROXY", "HTTPS_PROXY", "NO_PROXY",
		"http_proxy", "https_proxy", "no_proxy",
	}
	for _, v := range proxyVars {
		if val := os.Getenv(v); val != "" {
			b.WithEnv(v, val)
		}
	}

	if agentProxy := os.Getenv("MC_AGENT_PROXY"); agentProxy != "" {
		b.WithEnv("HTTP_PROXY", agentProxy)
		b.WithEnv("HTTPS_PROXY", agentProxy)

		if u, err := url.Parse(agentProxy); err == nil {
			host := u.Hostname()
			port := u.Port()
			if port == "" {
				switch u.Scheme {
				case "https":
					port = "443"
				default:
					port = "80"
				}
			}
			b.WithEnv("JAVA_TOOL_OPTIONS", fmt.Sprintf(
				"-Dhttp.proxyHost=%s -Dhttp.proxyPort=%s -Dhttps.proxyHost=%s -Dhttps.proxyPort=%s",
				host, port, host, port,
			))
		}
	}

	return b
}

// Build finalizes the builder and returns the Docker container, host and network configs.
func (b *ContainerConfigBuilder) Build() (*container.Config, *container.HostConfig, *network.NetworkingConfig) {
	config := &container.Config{
		Image:        b.image,
		Env:          b.env,
		ExposedPorts: b.exposedPorts,
		Labels:       b.labels,
	}

	hostConfig := &container.HostConfig{
		Binds:        b.bindMounts,
		PortBindings: b.portBindings,
		LogConfig: container.LogConfig{
			Type: "json-file",
			Config: map[string]string{
				"max-size": "10m",
				"max-file": "3",
			},
		},
	}
	if b.restartPolicy != "" {
		hostConfig.RestartPolicy = container.RestartPolicy{Name: container.RestartPolicyMode(b.restartPolicy)}
	}

	networkingConfig := &network.NetworkingConfig{}
	if b.networkName != "" {
		networkingConfig.EndpointsConfig = map[string]*network.EndpointSettings{
			b.networkName: {},
		}
	}

	return config, hostConfig, networkingConfig
}

// GetClient creates a new Docker client from environment settings.
func GetClient() (*client.Client, error) {
	cli, err := client.NewClientWithOpts(
		client.FromEnv,
		client.WithAPIVersionNegotiation(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize docker client: %w", err)
	}
	return cli, nil
}

// ImageExists checks whether the managed Minecraft server image is already present locally.
func ImageExists(ctx context.Context, cli *client.Client) (bool, error) {
	images, err := cli.ImageList(ctx, image.ListOptions{})
	if err != nil {
		return false, fmt.Errorf("failed to list docker images: %w", err)
	}

	for _, img := range images {
		for _, tag := range img.RepoTags {
			if tag == ImageName {
				return true, nil
			}
		}
	}
	return false, nil
}

// PullImageIfNeeded pulls the managed image if it is not already available locally.
func PullImageIfNeeded(ctx context.Context, cli *client.Client) error {
	exists, err := ImageExists(ctx, cli)
	if err != nil {
		return err
	}

	if !exists {
		slog.Info("pulling docker image", "image", ImageName)
		reader, err := cli.ImagePull(ctx, ImageName, image.PullOptions{})
		if err != nil {
			return fmt.Errorf("failed to pull image: %w", err)
		}
		defer func() { _ = reader.Close() }()
		_, _ = io.Copy(io.Discard, reader)
		slog.Info("docker image pulled", "image", ImageName)
	}
	return nil
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

// GenerateRconPassword creates a 16-character random alphanumeric string.
func GenerateRconPassword() (string, error) {
	const alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	for i := range b {
		b[i] = alphabet[int(b[i])%len(alphabet)]
	}
	return string(b), nil
}

// GenerateVolumeID creates an 8-character random alphanumeric string.
func GenerateVolumeID() (string, error) {
	const alphabet = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	for i := range b {
		b[i] = alphabet[int(b[i])%len(alphabet)]
	}
	return string(b), nil
}

// IsPortFree reports whether the given TCP port is available on the host.
func IsPortFree(host string, port uint16) bool {
	addr := fmt.Sprintf("%s:%d", host, port)
	if host == "" {
		addr = fmt.Sprintf(":%d", port)
	}
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return false
	}
	_ = ln.Close()
	return true
}

// FindFreePort returns preferred if it is free on the given host,
// otherwise it searches for the next free port starting from preferred+1,
// wrapping around at 65535 to 1024.
func FindFreePort(host string, preferred uint16) (uint16, error) {
	return FindFreePortExcluding(host, preferred, func(uint16) bool { return false })
}

// FindFreePortExcluding returns preferred if it is free on the given host
// and not excluded by the filter; otherwise it searches for the next free port.
func FindFreePortExcluding(host string, preferred uint16, exclude func(uint16) bool) (uint16, error) {
	if preferred != 0 && IsPortFree(host, preferred) && !exclude(preferred) {
		return preferred, nil
	}
	start := uint16(1024)
	if preferred != 0 {
		start = preferred + 1
	}
	for p := start; p < 65535; p++ {
		if IsPortFree(host, p) && !exclude(p) {
			return p, nil
		}
	}
	for p := uint16(1024); p < start; p++ {
		if IsPortFree(host, p) && !exclude(p) {
			return p, nil
		}
	}
	return 0, fmt.Errorf("no free port found on host %q", host)
}

// StartServerContainer prepares the host environment, cleans up residual dead containers,
// creates a unique servers/<volume_id> directory, builds configuration and runs the container.
// It returns the Docker container ID, the generated volume ID, and the absolute host path of the volume.
// If existingVolumePath is non-empty and the directory exists, it is reused instead of creating a new one.
// HostPathForDocker translates a local path inside the backend container to the
// corresponding host path so that Docker daemon binds the correct directory.
func HostPathForDocker(localPath, serversDir, serversHostDir string) string {
	rel, err := filepath.Rel(serversDir, localPath)
	if err != nil || strings.HasPrefix(rel, "..") {
		return localPath
	}
	return filepath.Join(serversHostDir, rel)
}

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
	if err := PullImageIfNeeded(ctx, cli); err != nil {
		return "", "", "", err
	}
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

// SplitEnv splits a KEY=VALUE string into its components.
func SplitEnv(s string) (key, value string, ok bool) {
	for i := 0; i < len(s); i++ {
		if s[i] == '=' {
			return s[:i], s[i+1:], true
		}
	}
	return "", "", false
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
