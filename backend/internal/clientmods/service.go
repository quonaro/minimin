package clientmods

import (
	"github.com/docker/docker/client"

	"orchestrator/internal/state"
)

// Service provides client-side mod operations.
type Service struct {
	instance *state.InstanceFile
	cli      *client.Client
}

// NewService creates a new clientmods service.
func NewService(instance *state.InstanceFile, cli *client.Client) *Service {
	return &Service{instance: instance, cli: cli}
}

// getClientModsDir returns the mods-client directory for a server.
func (s *Service) getClientModsDir(serverID string) (string, error) {
	st, ok := s.instance.Get(serverID)
	if !ok {
		return "", nil
	}
	if st.VolumePath == "" {
		return "", nil
	}
	return st.VolumePath + "/mods-client", nil
}
