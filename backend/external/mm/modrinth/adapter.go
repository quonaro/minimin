package modrinth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"orchestrator/external/mm"
)

const defaultBaseURL = "https://api.modrinth.com/v2"

// Adapter implements mm.ContentSource for the Modrinth API.
type Adapter struct {
	httpClient *http.Client
	cache      *Cache
	baseURL    string
}

// NewAdapter creates a new Modrinth adapter.
func NewAdapter(baseURL string) *Adapter {
	if baseURL == "" {
		baseURL = defaultBaseURL
	}
	return &Adapter{
		httpClient: &http.Client{Timeout: 15 * time.Second},
		cache:      NewCache(),
		baseURL:    baseURL,
	}
}

// Name returns the source name.
func (a *Adapter) Name() string { return "modrinth" }

// Capabilities lists supported content types.
func (a *Adapter) Capabilities() []mm.ContentType {
	return []mm.ContentType{mm.ContentMod, mm.ContentResourcepack, mm.ContentShaderpack}
}

func (a *Adapter) doWithRetry(req *http.Request) (*http.Response, error) {
	const maxRetries = 3
	delay := 500 * time.Millisecond

	for i := 0; i < maxRetries; i++ {
		resp, err := a.httpClient.Do(req)
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
func (a *Adapter) Search(ctx context.Context, req mm.SearchRequest) ([]mm.ContentSummary, int, error) {
	facets := a.buildFacets(req)
	key := fmt.Sprintf("%s:%s:%d:%d", req.Query, facets, req.Offset, req.Limit)
	if cached, ok := a.cache.GetSearch(key); ok {
		return a.mapSearchResponse(cached), cached.TotalHits, nil
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, a.baseURL+"/search", nil)
	if err != nil {
		return nil, 0, err
	}
	q := httpReq.URL.Query()
	q.Set("query", req.Query)
	q.Set("offset", fmt.Sprintf("%d", req.Offset))
	q.Set("limit", fmt.Sprintf("%d", req.Limit))
	if facets != "" {
		q.Set("facets", facets)
	}
	httpReq.URL.RawQuery = q.Encode()

	resp, err := a.doWithRetry(httpReq)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	var result SearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, 0, err
	}
	a.cache.SetSearch(key, result)
	return a.mapSearchResponse(result), result.TotalHits, nil
}

func (a *Adapter) mapSearchResponse(r SearchResponse) []mm.ContentSummary {
	items := make([]mm.ContentSummary, len(r.Hits))
	for i, h := range r.Hits {
		items[i] = mapHitToSummary(h)
	}
	return items
}

func (a *Adapter) buildFacets(req mm.SearchRequest) string {
	var parts []string
	switch req.ContentType {
	case mm.ContentMod:
		parts = append(parts, `["project_type:mod"]`)
	case mm.ContentResourcepack:
		parts = append(parts, `["project_type:resourcepack"]`)
	case mm.ContentShaderpack:
		parts = append(parts, `["project_type:shader"]`)
	}
	if req.Loader != "" {
		parts = append(parts, fmt.Sprintf(`["categories:%s"]`, req.Loader))
	}
	if req.GameVersion != "" {
		parts = append(parts, fmt.Sprintf(`["versions:%s"]`, req.GameVersion))
	}
	if len(parts) == 0 {
		return ""
	}
	return "[" + strings.Join(parts, ",") + "]"
}

// GetContent fetches detailed project info.
func (a *Adapter) GetContent(ctx context.Context, id string) (*mm.ContentDetail, error) {
	if cached, ok := a.cache.GetProject(id); ok {
		return mapProjectToDetail(cached), nil
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, a.baseURL+"/project/"+id, nil)
	if err != nil {
		return nil, err
	}
	resp, err := a.doWithRetry(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result Project
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	a.cache.SetProject(id, result)
	return mapProjectToDetail(result), nil
}

// GetVersions fetches versions for a project.
func (a *Adapter) GetVersions(ctx context.Context, projectID string, filter mm.VersionFilter) ([]mm.ContentVersion, error) {
	key := fmt.Sprintf("%s:%s:%s", projectID, strings.Join(filter.Loaders, ","), strings.Join(filter.GameVersions, ","))
	if cached, ok := a.cache.GetVersions(key); ok {
		return a.mapVersions(cached), nil
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, a.baseURL+"/project/"+projectID+"/version", nil)
	if err != nil {
		return nil, err
	}
	q := req.URL.Query()
	if len(filter.Loaders) > 0 {
		q.Set("loaders", `["`+strings.Join(filter.Loaders, `","`)+`"]`)
	}
	if len(filter.GameVersions) > 0 {
		q.Set("game_versions", `["`+strings.Join(filter.GameVersions, `","`)+`"]`)
	}
	req.URL.RawQuery = q.Encode()

	resp, err := a.doWithRetry(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result []Version
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	a.cache.SetVersions(key, result)
	return a.mapVersions(result), nil
}

func (a *Adapter) mapVersions(vs []Version) []mm.ContentVersion {
	out := make([]mm.ContentVersion, len(vs))
	for i, v := range vs {
		out[i] = mapVersion(v)
	}
	return out
}

// GetVersion fetches a single version by ID.
func (a *Adapter) GetVersion(ctx context.Context, id string) (*mm.ContentVersion, error) {
	if cached, ok := a.cache.GetVersion(id); ok {
		v := mapVersion(cached)
		return &v, nil
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, a.baseURL+"/version/"+id, nil)
	if err != nil {
		return nil, err
	}
	resp, err := a.doWithRetry(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result Version
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	a.cache.SetVersion(id, result)
	v := mapVersion(result)
	return &v, nil
}

// GetDownloadURL returns a direct download URL for a version.
func (a *Adapter) GetDownloadURL(ctx context.Context, versionID string) (string, error) {
	v, err := a.GetVersion(ctx, versionID)
	if err != nil {
		return "", err
	}
	for _, f := range v.Files {
		if f.Primary {
			return f.URL, nil
		}
	}
	if len(v.Files) > 0 {
		return v.Files[0].URL, nil
	}
	return "", fmt.Errorf("no files for version %s", versionID)
}
