package mm

// ContentSummary is a lightweight result returned from Search.
type ContentSummary struct {
	ID            string   `json:"id"`
	Slug          string   `json:"slug"`
	Title         string   `json:"title"`
	Description   string   `json:"description"`
	IconURL       string   `json:"icon_url"`
	Author        string   `json:"author"`
	Downloads     int      `json:"downloads"`
	ServerSide    string   `json:"server_side,omitempty"`
	ClientSide    string   `json:"client_side,omitempty"`
	Versions      []string `json:"versions,omitempty"`
	LatestVersion string   `json:"latest_version,omitempty"`
}

// ContentDetail is the full metadata for a single piece of content.
type ContentDetail struct {
	ID          string `json:"id"`
	Slug        string `json:"slug"`
	Title       string `json:"title"`
	Description string `json:"description"`
	IconURL     string `json:"icon_url"`
	Author      string `json:"author"`
	Downloads   int    `json:"downloads"`
	ServerSide  string `json:"server_side,omitempty"`
	ClientSide  string `json:"client_side,omitempty"`
}

// ContentVersion represents a downloadable version.
type ContentVersion struct {
	ID            string              `json:"id"`
	ProjectID     string              `json:"project_id"`
	VersionNumber string              `json:"version_number"`
	GameVersions  []string            `json:"game_versions"`
	Loaders       []string            `json:"loaders"`
	Files         []ContentFile       `json:"files"`
	Dependencies  []ContentDependency `json:"dependencies"`
}

// ContentFile is a single binary inside a version.
type ContentFile struct {
	URL      string `json:"url"`
	Filename string `json:"filename"`
	Primary  bool   `json:"primary"`
}

// ContentDependency describes a required or optional dependency.
type ContentDependency struct {
	VersionID      string `json:"version_id,omitempty"`
	ProjectID      string `json:"project_id,omitempty"`
	DependencyType string `json:"dependency_type"`
}

// SearchRequest carries parameters for a source search.
type SearchRequest struct {
	ContentType ContentType
	Query       string
	GameVersion string
	Loader      string
	Offset      int
	Limit       int
}

// VersionFilter narrows the version list for a given content item.
type VersionFilter struct {
	GameVersions []string
	Loaders      []string
}
