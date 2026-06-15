package mods

import (
	"github.com/docker/docker/client"

	"orchestrator/internal/state"
)

// Service provides server-side mod operations.
type Service struct {
	instance       *state.InstanceFile
	cli            *client.Client
	modUploadMaxMB int
}

// NewService creates a new mods service.
func NewService(instance *state.InstanceFile, cli *client.Client, modUploadMaxMB int) *Service {
	if modUploadMaxMB <= 0 {
		modUploadMaxMB = 1024
	}
	return &Service{
		instance:       instance,
		cli:            cli,
		modUploadMaxMB: modUploadMaxMB,
	}
}
