package state

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"gopkg.in/yaml.v3"
)

// ServerState holds all persisted metadata for a single managed Minecraft server.
type ServerState struct {
	ServerID           string            `json:"serverId"     yaml:"server_id"`
	VolumeID           string            `json:"volumeId"     yaml:"volume_id"`
	VolumePath         string            `json:"volumePath"   yaml:"volume_path"`
	HostPath           string            `json:"hostPath"     yaml:"host_path"`
	ContainerPath      string            `json:"containerPath" yaml:"container_path"`
	ContainerID        string            `json:"containerId"  yaml:"container_id"`
	RamBytes           int64             `json:"ramBytes"     yaml:"ram_bytes"`
	GamePort           uint16            `json:"gamePort"     yaml:"game_port"`
	EngineType         string            `json:"engineType"   yaml:"engine_type"`
	GameVersion        string            `json:"gameVersion"  yaml:"game_version"`
	LoaderVersion      string            `json:"loaderVersion" yaml:"loader_version,omitempty"`
	RconPassword       string            `json:"rconPassword" yaml:"rcon_password"`
	RconPort           uint16            `json:"rconPort"     yaml:"rcon_port"`
	PublicRcon         bool              `json:"publicRcon"   yaml:"public_rcon"`
	RestartPolicy      string            `json:"restartPolicy" yaml:"restart_policy,omitempty"`
	Status             string            `json:"-"            yaml:"status,omitempty"`     // legacy
	StartedAt          time.Time         `json:"-"            yaml:"started_at,omitempty"` // legacy
	ContainerStatus    string            `json:"containerStatus" yaml:"container_status"`
	ContainerStartedAt time.Time         `json:"containerStartedAt" yaml:"container_started_at"`
	ServerStatus       string            `json:"serverStatus" yaml:"server_status"`
	ServerStartedAt    time.Time         `json:"serverStartedAt" yaml:"server_started_at"`
	CreatedAt          time.Time         `json:"createdAt"      yaml:"created_at"`
	UpdatedAt          time.Time         `json:"updatedAt"      yaml:"updated_at"`
	DesiredStatus      string            `json:"desiredStatus"      yaml:"desired_status,omitempty"`
	ModCount           int               `json:"modCount"       yaml:"mod_count"`
	ExternalJavaArgs   []string          `json:"externalJavaArgs"   yaml:"external_java_args,omitempty"`
	PendingProperties  map[string]string `json:"pendingProperties"  yaml:"pending_properties,omitempty"`
}

// InstanceFile is the in-memory representation of instance.yml.
type InstanceFile struct {
	APIKey  string                 `yaml:"api_key"`
	Servers map[string]ServerState `yaml:"servers"`
	mu      sync.RWMutex
	path    string
	// Broadcast is called after every state mutation. It may be nil.
	Broadcast func(serverID string, state ServerState) `yaml:"-"`
}

// Load reads instance.yml from disk or creates an empty structure if the file does not exist.
func Load(path string) (*InstanceFile, error) {
	i := &InstanceFile{
		Servers: make(map[string]ServerState),
		path:    path,
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return i, nil
		}
		return nil, fmt.Errorf("failed to read instance.yml: %w", err)
	}

	if err := yaml.Unmarshal(data, i); err != nil {
		return nil, fmt.Errorf("failed to parse instance.yml: %w", err)
	}

	if i.Servers == nil {
		i.Servers = make(map[string]ServerState)
	}

	// Migrate legacy fields
	for k, s := range i.Servers {
		migrated := false
		if s.ContainerStatus == "" && s.Status != "" {
			s.ContainerStatus = s.Status
			migrated = true
		}
		if s.ContainerStartedAt.IsZero() && !s.StartedAt.IsZero() {
			s.ContainerStartedAt = s.StartedAt
			migrated = true
		}
		if s.ServerStatus == "" {
			if s.ContainerStatus == "running" {
				s.ServerStatus = "starting"
			} else {
				s.ServerStatus = "stopped"
			}
			migrated = true
		}
		if migrated {
			s.Status = ""
			s.StartedAt = time.Time{}
			i.Servers[k] = s
		}
	}

	return i, nil
}

// Save atomically writes the current state back to instance.yml.
func (i *InstanceFile) Save() error {
	i.mu.RLock()
	defer i.mu.RUnlock()

	data, err := yaml.Marshal(i)
	if err != nil {
		return fmt.Errorf("failed to marshal instance.yml: %w", err)
	}

	if err := os.WriteFile(i.path, data, 0o644); err != nil {
		return fmt.Errorf("failed to write instance.yml: %w", err)
	}

	return nil
}

// Get retrieves a single server state by its logical server ID.
func (i *InstanceFile) Get(serverID string) (ServerState, bool) {
	i.mu.RLock()
	defer i.mu.RUnlock()
	s, ok := i.Servers[serverID]
	return s, ok
}

// Set inserts or updates a server state and stamps UpdatedAt.
func (i *InstanceFile) Set(s ServerState) {
	i.mu.Lock()
	defer i.mu.Unlock()
	if s.CreatedAt.IsZero() {
		s.CreatedAt = time.Now().UTC()
	}
	s.UpdatedAt = time.Now().UTC()
	i.Servers[s.ServerID] = s
	if i.Broadcast != nil {
		i.Broadcast(s.ServerID, s)
	}
}

// UpdateMeta patches mutable metadata fields and stamps UpdatedAt.
func (i *InstanceFile) UpdateMeta(serverID string, f func(*ServerState)) bool {
	i.mu.Lock()
	defer i.mu.Unlock()
	s, ok := i.Servers[serverID]
	if !ok {
		return false
	}
	f(&s)
	s.UpdatedAt = time.Now().UTC()
	i.Servers[serverID] = s
	if i.Broadcast != nil {
		i.Broadcast(s.ServerID, s)
	}
	return true
}

// Delete removes a server state from the file.
func (i *InstanceFile) Delete(serverID string) {
	i.mu.Lock()
	defer i.mu.Unlock()
	delete(i.Servers, serverID)
	if i.Broadcast != nil {
		i.Broadcast(serverID, ServerState{ServerID: serverID})
	}
}

// TrySetDesired sets the desired status if no operation is already pending.
// Returns the current state and true on success, false if an operation is in progress.
func (i *InstanceFile) TrySetDesired(serverID, desired string) (ServerState, bool) {
	i.mu.Lock()
	defer i.mu.Unlock()
	s, ok := i.Servers[serverID]
	if !ok {
		return ServerState{}, false
	}
	if s.DesiredStatus != "" && s.DesiredStatus != s.ContainerStatus {
		return s, false
	}
	s.DesiredStatus = desired
	s.UpdatedAt = time.Now().UTC()
	i.Servers[serverID] = s
	return s, true
}

// ClearDesired atomically sets the final status and clears DesiredStatus.
func (i *InstanceFile) ClearDesired(serverID, finalStatus string) {
	i.mu.Lock()
	defer i.mu.Unlock()
	s, ok := i.Servers[serverID]
	if !ok {
		return
	}
	s.ContainerStatus = finalStatus
	s.DesiredStatus = ""
	if finalStatus == "exited" {
		s.ServerStatus = "stopped"
		s.ServerStartedAt = time.Time{}
	} else if finalStatus == "running" && s.ServerStatus == "" {
		s.ServerStatus = "starting"
	}
	s.UpdatedAt = time.Now().UTC()
	i.Servers[serverID] = s
	if i.Broadcast != nil {
		i.Broadcast(s.ServerID, s)
	}
}

// AddPendingProperties merges new pending properties into a server's state.
func (i *InstanceFile) AddPendingProperties(serverID string, props map[string]string) bool {
	i.mu.Lock()
	defer i.mu.Unlock()
	s, ok := i.Servers[serverID]
	if !ok {
		return false
	}
	if s.PendingProperties == nil {
		s.PendingProperties = make(map[string]string)
	}
	for k, v := range props {
		s.PendingProperties[k] = v
	}
	s.UpdatedAt = time.Now().UTC()
	i.Servers[serverID] = s
	return true
}

// ClearPendingProperties removes all pending properties from a server's state.
func (i *InstanceFile) ClearPendingProperties(serverID string) bool {
	i.mu.Lock()
	defer i.mu.Unlock()
	s, ok := i.Servers[serverID]
	if !ok {
		return false
	}
	s.PendingProperties = nil
	s.UpdatedAt = time.Now().UTC()
	i.Servers[serverID] = s
	return true
}

// CountMods returns the number of .jar files in the server's mods directory.
func CountMods(s ServerState) int {
	if s.VolumePath == "" {
		return 0
	}
	modsDir := filepath.Join(s.VolumePath, "mods")
	entries, err := os.ReadDir(modsDir)
	if err != nil {
		return 0
	}
	count := 0
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(strings.ToLower(e.Name()), ".jar") {
			count++
		}
	}
	return count
}

// ReadServerJSON reads a JSON file from the server's volume path.
func ReadServerJSON(s ServerState, filename string) ([]map[string]any, error) {
	if s.VolumePath == "" {
		return []map[string]any{}, nil
	}
	p := filepath.Join(s.VolumePath, filename)
	data, err := os.ReadFile(p)
	if err != nil {
		if os.IsNotExist(err) {
			return []map[string]any{}, nil
		}
		return nil, err
	}
	var out []map[string]any
	if err := json.Unmarshal(data, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// IsPortUsed reports whether the given port is already assigned to any stored server.
// If excludeServerID is non-empty, that server is ignored.
func (i *InstanceFile) IsPortUsed(port uint16, excludeServerID string) bool {
	i.mu.RLock()
	defer i.mu.RUnlock()
	for id, s := range i.Servers {
		if id == excludeServerID {
			continue
		}
		if s.GamePort == port || s.RconPort == port {
			return true
		}
	}
	return false
}

// All returns a snapshot of every stored server state.
func (i *InstanceFile) All() []ServerState {
	i.mu.RLock()
	defer i.mu.RUnlock()
	out := make([]ServerState, 0, len(i.Servers))
	for _, v := range i.Servers {
		out = append(out, v)
	}
	sort.Slice(out, func(a, b int) bool {
		if !out[a].CreatedAt.Equal(out[b].CreatedAt) {
			return out[a].CreatedAt.Before(out[b].CreatedAt)
		}
		return out[a].ServerID < out[b].ServerID
	})
	return out
}
