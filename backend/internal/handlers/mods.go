package handlers

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"orchestrator/internal/modrinth"

	"github.com/danielgtaylor/huma/v2"
)

// ModSearchInput is the input for GET /api/mods/search.
type ModSearchInput struct {
	Query       string `query:"q" doc:"Search query"`
	Loader      string `query:"loader" doc:"Mod loader (fabric or forge)"`
	GameVersion string `query:"game_version" doc:"Minecraft game version"`
	Offset      int    `query:"offset" doc:"Pagination offset"`
	Limit       int    `query:"limit" doc:"Pagination limit"`
}

// ModSearchOutput is the output for GET /api/mods/search.
type ModSearchOutput struct {
	Body struct {
		Results []modrinth.SearchResult `json:"results"`
	}
}

// ModSearch proxies a search to Modrinth.
func (h *Handler) ModSearch(ctx context.Context, input *ModSearchInput) (*ModSearchOutput, error) {
	client := modrinth.NewClient()
	results, err := client.Search(input.Query, input.Loader, input.GameVersion, input.Offset, input.Limit)
	if err != nil {
		return nil, huma.Error500InternalServerError("modrinth search failed", err)
	}
	return &ModSearchOutput{Body: struct {
		Results []modrinth.SearchResult `json:"results"`
	}{Results: results}}, nil
}

// ModVersionsInput is the input for GET /api/mods/versions/{project_id}.
type ModVersionsInput struct {
	ProjectID   string `path:"project_id" doc:"Modrinth project ID"`
	Loader      string `query:"loader" doc:"Mod loader"`
	GameVersion string `query:"game_version" doc:"Minecraft game version"`
}

// ModVersionsOutput is the output for GET /api/mods/versions/{project_id}.
type ModVersionsOutput struct {
	Body struct {
		Versions []modrinth.Version `json:"versions"`
	}
}

// ModVersions returns available versions for a Modrinth project.
func (h *Handler) ModVersions(ctx context.Context, input *ModVersionsInput) (*ModVersionsOutput, error) {
	client := modrinth.NewClient()
	versions, err := client.GetProjectVersions(input.ProjectID, input.Loader, input.GameVersion)
	if err != nil {
		return nil, huma.Error500InternalServerError("modrinth versions failed", err)
	}
	return &ModVersionsOutput{Body: struct {
		Versions []modrinth.Version `json:"versions"`
	}{Versions: versions}}, nil
}

// ModInstallInput is the input for POST /api/mods/install.
type ModInstallInput struct {
	Body struct {
		AgentID   string `json:"agentId" doc:"Agent ID"`
		ServerID  string `json:"serverId" doc:"Server ID"`
		ProjectID string `json:"projectId" doc:"Modrinth project ID"`
		VersionID string `json:"versionId" doc:"Modrinth version ID"`
	}
}

// ModInstallOutput is the output for POST /api/mods/install.
type ModInstallOutput struct {
	Body struct {
		Success  bool   `json:"success"`
		Filename string `json:"filename,omitempty"`
	}
}

// ModInstall downloads a mod from Modrinth and uploads it to the agent.
func (h *Handler) ModInstall(ctx context.Context, input *ModInstallInput) (*ModInstallOutput, error) {
	agent, ok := h.DB.GetAgent(input.Body.AgentID)
	if !ok {
		return nil, huma.Error404NotFound("agent not found", nil)
	}

	client := modrinth.NewClient()
	ver, err := client.GetVersion(input.Body.VersionID)
	if err != nil {
		return nil, huma.Error500InternalServerError("failed to resolve mod version", err)
	}

	var fileURL, filename string
	for _, f := range ver.Files {
		if f.Primary {
			fileURL = f.URL
			filename = f.Filename
			break
		}
	}
	if fileURL == "" && len(ver.Files) > 0 {
		fileURL = ver.Files[0].URL
		filename = ver.Files[0].Filename
	}
	if fileURL == "" {
		return nil, huma.Error404NotFound("no download file found for version", nil)
	}

	// Download from Modrinth
	body, dlErr := client.DownloadFile(fileURL)
	if dlErr != nil {
		return nil, huma.Error500InternalServerError("failed to download mod", dlErr)
	}
	defer func() { _ = body.Close() }()

	// Stream upload to agent
	if err := h.uploadModToAgent(agent.Host, agent.APIKey, input.Body.ServerID, filename, body); err != nil {
		return nil, huma.Error500InternalServerError("failed to upload mod to agent", err)
	}

	return &ModInstallOutput{Body: struct {
		Success  bool   `json:"success"`
		Filename string `json:"filename,omitempty"`
	}{Success: true, Filename: filename}}, nil
}

// ModDownloadInput is the input for POST /api/mods/download.
type ModDownloadInput struct {
	Body struct {
		AgentID  string `json:"agentId" doc:"Agent ID"`
		ServerID string `json:"serverId" doc:"Server ID"`
		URL      string `json:"url" doc:"Direct download URL"`
		Filename string `json:"filename,omitempty" doc:"Optional target filename"`
	}
}

// ModDownloadOutput is the output for POST /api/mods/download.
type ModDownloadOutput struct {
	Body struct {
		Success  bool   `json:"success"`
		Filename string `json:"filename,omitempty"`
	}
}

// ModDownload fetches a file from a URL and uploads it to the agent.
func (h *Handler) ModDownload(ctx context.Context, input *ModDownloadInput) (*ModDownloadOutput, error) {
	agent, ok := h.DB.GetAgent(input.Body.AgentID)
	if !ok {
		return nil, huma.Error404NotFound("agent not found", nil)
	}

	parsedURL, err := url.Parse(input.Body.URL)
	if err != nil || (parsedURL.Scheme != "http" && parsedURL.Scheme != "https") {
		return nil, huma.Error400BadRequest("invalid URL", err)
	}

	client := &http.Client{Timeout: 120 * time.Second}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, input.Body.URL, nil)
	if err != nil {
		return nil, huma.Error500InternalServerError("failed to create request", err)
	}
	req.Header.Set("User-Agent", "Minecraft-Server-Manager/1.0")

	resp, err := client.Do(req)
	if err != nil {
		return nil, huma.Error500InternalServerError("download failed", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= 300 {
		return nil, huma.Error500InternalServerError("download returned non-2xx", fmt.Errorf("status %d", resp.StatusCode))
	}

	filename := input.Body.Filename
	if filename == "" {
		filename = filenameFromResponse(resp, parsedURL)
	}
	if filename == "" {
		filename = "mod.jar"
	}

	if err := h.uploadModToAgent(agent.Host, agent.APIKey, input.Body.ServerID, filename, resp.Body); err != nil {
		return nil, huma.Error500InternalServerError("failed to upload mod to agent", err)
	}

	return &ModDownloadOutput{Body: struct {
		Success  bool   `json:"success"`
		Filename string `json:"filename,omitempty"`
	}{Success: true, Filename: filename}}, nil
}

func filenameFromResponse(resp *http.Response, parsedURL *url.URL) string {
	cd := resp.Header.Get("Content-Disposition")
	if cd != "" {
		if idx := strings.Index(cd, "filename="); idx != -1 {
			fn := cd[idx+len("filename="):]
			fn = strings.Trim(fn, `"`)
			if fn != "" {
				return path.Base(fn)
			}
		}
	}
	return path.Base(parsedURL.Path)
}

func (h *Handler) uploadModToAgent(agentHost, apiKey, serverID, filename string, reader io.Reader) error {
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	fw, err := mw.CreateFormFile("file", filename)
	if err != nil {
		return err
	}
	if _, err := io.Copy(fw, reader); err != nil {
		return err
	}
	if err := mw.Close(); err != nil {
		return err
	}

	targetURL := strings.TrimSuffix(agentHost, "/") + "/api/v1/servers/" + serverID + "/mods/upload"
	req, err := http.NewRequest(http.MethodPost, targetURL, &buf)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", mw.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("agent returned %d: %s", resp.StatusCode, string(body))
	}
	return nil
}
