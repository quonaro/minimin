package modrinth

// SearchResponse represents the Modrinth search API response.
type SearchResponse struct {
	Hits      []Hit `json:"hits"`
	Offset    int   `json:"offset"`
	Limit     int   `json:"limit"`
	TotalHits int   `json:"total_hits"`
}

// Hit represents a single search result from Modrinth.
type Hit struct {
	ProjectID   string `json:"project_id"`
	Slug        string `json:"slug"`
	Title       string `json:"title"`
	Description string `json:"description"`
	IconURL     string `json:"icon_url"`
	Author      string `json:"author"`
	Downloads   int    `json:"downloads"`
	ServerSide  string `json:"server_side"`
	ClientSide  string `json:"client_side"`
}

// Project represents detailed project info from Modrinth.
type Project struct {
	ID          string `json:"id"`
	Slug        string `json:"slug"`
	Title       string `json:"title"`
	Description string `json:"description"`
	IconURL     string `json:"icon_url"`
	Author      string `json:"author"`
	Downloads   int    `json:"downloads"`
	ServerSide  string `json:"server_side"`
	ClientSide  string `json:"client_side"`
}

// Version represents a project version from Modrinth.
type Version struct {
	ID            string        `json:"id"`
	ProjectID     string        `json:"project_id"`
	VersionNumber string        `json:"version_number"`
	GameVersions  []string      `json:"game_versions"`
	Loaders       []string      `json:"loaders"`
	Files         []VersionFile `json:"files"`
	Dependencies  []Dependency  `json:"dependencies"`
}

// VersionFile represents a single file within a version.
type VersionFile struct {
	URL      string `json:"url"`
	Filename string `json:"filename"`
	Primary  bool   `json:"primary"`
}

// Dependency represents a mod dependency.
type Dependency struct {
	VersionID      string `json:"version_id"`
	ProjectID      string `json:"project_id"`
	DependencyType string `json:"dependency_type"`
}

// SearchParams holds query parameters for the search endpoint.
type SearchParams struct {
	Query  string
	Facets string
	Offset int
	Limit  int
}

// VersionParams holds query parameters for the versions endpoint.
type VersionParams struct {
	Loaders      []string
	GameVersions []string
}
