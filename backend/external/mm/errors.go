package mm

import "errors"

// Sentinel errors returned by ContentSource implementations.
var (
	ErrNotFound              = errors.New("not found")
	ErrRateLimited           = errors.New("rate limited")
	ErrUnsupportedContentType = errors.New("unsupported content type")
)
