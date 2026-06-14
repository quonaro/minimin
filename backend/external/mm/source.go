package mm

import "context"

// ContentSource is the unified contract for every external mod/content provider.
type ContentSource interface {
	// Name returns the canonical name of the source (e.g. "modrinth").
	Name() string

	// Capabilities lists which content types this source supports.
	Capabilities() []ContentType

	// Search queries the source for content matching the request.
	Search(ctx context.Context, req SearchRequest) (items []ContentSummary, total int, err error)

	// GetContent fetches detailed metadata for a single content item.
	GetContent(ctx context.Context, id string) (*ContentDetail, error)

	// GetVersions returns all versions for a content item, optionally filtered.
	GetVersions(ctx context.Context, contentID string, filter VersionFilter) ([]ContentVersion, error)

	// GetVersion fetches a single version by its ID.
	GetVersion(ctx context.Context, versionID string) (*ContentVersion, error)

	// GetDownloadURL returns a direct download URL for a version.
	GetDownloadURL(ctx context.Context, versionID string) (string, error)
}
