package runner

import (
	"fmt"
	"net/url"
	"os"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/go-connections/nat"
)

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
