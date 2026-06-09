package modrinth

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const baseURL = "https://api.modrinth.com/v2"

// Client is a thin HTTP client for the Modrinth API.
type Client struct {
	httpClient *http.Client
}

// NewClient creates a Modrinth API client.
func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{Timeout: 15 * time.Second},
	}
}

// SearchResult is a single project from Modrinth search.
type SearchResult struct {
	ProjectID   string `json:"project_id"`
	Slug        string `json:"slug"`
	Title       string `json:"title"`
	Description string `json:"description"`
	IconURL     string `json:"icon_url"`
	Author      string `json:"author"`
	Downloads   int    `json:"downloads"`
}

// SearchHits is the response from /search.
type SearchHits struct {
	Hits []SearchResult `json:"hits"`
}

// Search queries Modrinth for mods.
func (c *Client) Search(query, loader, gameVersion string, offset, limit int) ([]SearchResult, error) {
	if limit == 0 {
		limit = 20
	}
	u, _ := url.Parse(baseURL + "/search")
	q := u.Query()
	q.Set("query", query)
	q.Set("offset", fmt.Sprintf("%d", offset))
	q.Set("limit", fmt.Sprintf("%d", limit))
	var facets [][]string
	if loader != "" {
		facets = append(facets, []string{fmt.Sprintf("categories:%s", loader)})
	}
	if gameVersion != "" {
		facets = append(facets, []string{fmt.Sprintf("versions:%s", gameVersion)})
	}
	if len(facets) > 0 {
		fb, _ := json.Marshal(facets)
		q.Set("facets", string(fb))
	}
	u.RawQuery = q.Encode()

	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("modrinth search request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("modrinth search returned %d: %s", resp.StatusCode, string(body))
	}
	var result SearchHits
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode modrinth search: %w", err)
	}
	return result.Hits, nil
}

// VersionFile is a downloadable file within a Modrinth version.
type VersionFile struct {
	URL      string `json:"url"`
	Filename string `json:"filename"`
	Primary  bool   `json:"primary"`
	SHA1     string `json:"hashes,omitempty"`
}

// Version is a specific version of a Modrinth project.
type Version struct {
	ID            string        `json:"id"`
	ProjectID     string        `json:"project_id"`
	VersionNumber string        `json:"version_number"`
	GameVersions  []string      `json:"game_versions"`
	Loaders       []string      `json:"loaders"`
	Files         []VersionFile `json:"files"`
}

// GetVersion fetches a specific version by its ID.
func (c *Client) GetVersion(versionID string) (*Version, error) {
	u := baseURL + "/version/" + versionID
	req, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("modrinth version request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("modrinth version returned %d: %s", resp.StatusCode, string(body))
	}
	var v Version
	if err := json.NewDecoder(resp.Body).Decode(&v); err != nil {
		return nil, fmt.Errorf("failed to decode modrinth version: %w", err)
	}
	return &v, nil
}

// GetProjectVersions fetches all versions for a project, optionally filtering by loader and game version.
func (c *Client) GetProjectVersions(projectID, loader, gameVersion string) ([]Version, error) {
	u, _ := url.Parse(baseURL + "/project/" + projectID + "/version")
	q := u.Query()
	if loader != "" {
		q.Set("loaders", fmt.Sprintf("[\"%s\"]", loader))
	}
	if gameVersion != "" {
		q.Set("game_versions", fmt.Sprintf("[\"%s\"]", gameVersion))
	}
	u.RawQuery = q.Encode()

	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("modrinth project versions request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("modrinth project versions returned %d: %s", resp.StatusCode, string(body))
	}
	var versions []Version
	if err := json.NewDecoder(resp.Body).Decode(&versions); err != nil {
		return nil, fmt.Errorf("failed to decode modrinth versions: %w", err)
	}
	return versions, nil
}

// DownloadFile streams a file from the given URL into the provided writer.
func (c *Client) DownloadFile(fileURL string) (io.ReadCloser, error) {
	resp, err := c.httpClient.Get(fileURL)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		_ = resp.Body.Close()
		return nil, fmt.Errorf("download returned %d", resp.StatusCode)
	}
	return resp.Body, nil
}
