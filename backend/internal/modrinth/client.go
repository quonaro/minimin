package modrinth

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const defaultBaseURL = "https://api.modrinth.com/v2"

// Client wraps HTTP calls to the Modrinth API with caching.
type Client struct {
	httpClient *http.Client
	cache      *Cache
	baseURL    string
}

// NewClient creates a new Modrinth client.
// If baseURL is empty, the official Modrinth API is used.
func NewClient(baseURL string) *Client {
	if baseURL == "" {
		baseURL = defaultBaseURL
	}
	return &Client{
		httpClient: &http.Client{Timeout: 15 * time.Second},
		cache:      NewCache(),
		baseURL:    baseURL,
	}
}

func (c *Client) doWithRetry(req *http.Request) (*http.Response, error) {
	const maxRetries = 3
	delay := 500 * time.Millisecond

	for i := 0; i < maxRetries; i++ {
		resp, err := c.httpClient.Do(req)
		if err != nil {
			return nil, err
		}

		if resp.StatusCode == http.StatusOK {
			return resp, nil
		}

		_, _ = io.Copy(io.Discard, resp.Body)
		resp.Body.Close()

		if resp.StatusCode != http.StatusTooManyRequests && resp.StatusCode < 500 {
			return nil, fmt.Errorf("modrinth returned %d", resp.StatusCode)
		}

		if i == maxRetries-1 {
			break
		}

		time.Sleep(delay)
		if delay < 5*time.Second {
			delay *= 2
		}
	}

	return nil, fmt.Errorf("modrinth request failed after %d retries", maxRetries)
}

// Search queries Modrinth for projects.
func (c *Client) Search(params SearchParams) (SearchResponse, error) {
	key := fmt.Sprintf("%s:%s:%d:%d", params.Query, params.Facets, params.Offset, params.Limit)
	if cached, ok := c.cache.GetSearch(key); ok {
		return cached, nil
	}

	req, err := http.NewRequest(http.MethodGet, c.baseURL+"/search", nil)
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

	resp, err := c.doWithRetry(req)
	if err != nil {
		return SearchResponse{}, err
	}
	defer resp.Body.Close()

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

	req, err := http.NewRequest(http.MethodGet, c.baseURL+"/project/"+id, nil)
	if err != nil {
		return Project{}, err
	}
	resp, err := c.doWithRetry(req)
	if err != nil {
		return Project{}, err
	}
	defer resp.Body.Close()

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

	req, err := http.NewRequest(http.MethodGet, c.baseURL+"/project/"+projectID+"/version", nil)
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

	resp, err := c.doWithRetry(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

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

	req, err := http.NewRequest(http.MethodGet, c.baseURL+"/version/"+id, nil)
	if err != nil {
		return Version{}, err
	}
	resp, err := c.doWithRetry(req)
	if err != nil {
		return Version{}, err
	}
	defer resp.Body.Close()

	var result Version
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return Version{}, err
	}
	c.cache.SetVersion(id, result)
	return result, nil
}
