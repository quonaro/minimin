package modrinth

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const baseURL = "https://api.modrinth.com/v2"

// Client wraps HTTP calls to the Modrinth API with caching.
type Client struct {
	httpClient *http.Client
	cache      *Cache
}

// NewClient creates a new Modrinth client.
func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{Timeout: 15 * time.Second},
		cache:      NewCache(),
	}
}

// Search queries Modrinth for projects.
func (c *Client) Search(params SearchParams) (SearchResponse, error) {
	key := fmt.Sprintf("%s:%s:%d:%d", params.Query, params.Facets, params.Offset, params.Limit)
	if cached, ok := c.cache.GetSearch(key); ok {
		return cached, nil
	}

	req, err := http.NewRequest(http.MethodGet, baseURL+"/search", nil)
	if err != nil {
		return SearchResponse{}, err
	}
	q := req.URL.Query()
	q.Set("query", params.Query)
	q.Set("offset", fmt.Sprintf("%d", params.Offset))
	q.Set("limit", fmt.Sprintf("%d", params.Limit))
	if params.Facets != "" {
		q.Set("facets", params.Facets)
	}
	req.URL.RawQuery = q.Encode()

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return SearchResponse{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return SearchResponse{}, fmt.Errorf("modrinth search returned %d", resp.StatusCode)
	}

	var result SearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return SearchResponse{}, err
	}
	c.cache.SetSearch(key, result)
	return result, nil
}

// GetProject fetches detailed project info.
func (c *Client) GetProject(id string) (Project, error) {
	if cached, ok := c.cache.GetProject(id); ok {
		return cached, nil
	}

	resp, err := c.httpClient.Get(baseURL + "/project/" + id)
	if err != nil {
		return Project{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return Project{}, fmt.Errorf("modrinth project returned %d", resp.StatusCode)
	}

	var result Project
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return Project{}, err
	}
	c.cache.SetProject(id, result)
	return result, nil
}

// GetVersions fetches versions for a project.
func (c *Client) GetVersions(projectID string, params VersionParams) ([]Version, error) {
	key := fmt.Sprintf("%s:%s:%s", projectID, strings.Join(params.Loaders, ","), strings.Join(params.GameVersions, ","))
	if cached, ok := c.cache.GetVersions(key); ok {
		return cached, nil
	}

	req, err := http.NewRequest(http.MethodGet, baseURL+"/project/"+projectID+"/version", nil)
	if err != nil {
		return nil, err
	}
	q := req.URL.Query()
	if len(params.Loaders) > 0 {
		q.Set("loaders", "[\""+strings.Join(params.Loaders, "\",\"")+"\"]")
	}
	if len(params.GameVersions) > 0 {
		q.Set("game_versions", "[\""+strings.Join(params.GameVersions, "\",\"")+"\"]")
	}
	req.URL.RawQuery = q.Encode()

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("modrinth versions returned %d", resp.StatusCode)
	}

	var result []Version
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	c.cache.SetVersions(key, result)
	return result, nil
}

// GetVersion fetches a single version by ID.
func (c *Client) GetVersion(id string) (Version, error) {
	if cached, ok := c.cache.GetVersion(id); ok {
		return cached, nil
	}

	resp, err := c.httpClient.Get(baseURL + "/version/" + id)
	if err != nil {
		return Version{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return Version{}, fmt.Errorf("modrinth version returned %d", resp.StatusCode)
	}

	var result Version
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return Version{}, err
	}
	c.cache.SetVersion(id, result)
	return result, nil
}
