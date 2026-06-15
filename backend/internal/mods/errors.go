package mods

import "errors"

var (
	ErrNotFound             = errors.New("not found")
	ErrVolumeNotInitialized = errors.New("server volume not initialized")
	ErrInvalidFilename      = errors.New("invalid filename")
	ErrOperationInProgress  = errors.New("operation in progress")
)
