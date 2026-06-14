package handlers

import (
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"
)

const maxEditableFileBytes int64 = 1 << 20 // 1MB

// ServerFileEntry describes one filesystem object inside server volume.
type ServerFileEntry struct {
	Name       string    `json:"name"`
	Path       string    `json:"path"`
	IsDir      bool      `json:"isDir"`
	Size       int64     `json:"size"`
	ModifiedAt time.Time `json:"modifiedAt"`
}

func (h *Handler) resolveServerPath(serverID, relPath string, allowEmpty bool) (string, string, error) {
	s, ok := h.Instance.Get(serverID)
	if !ok {
		return "", "", errors.New("not found")
	}
	if s.VolumePath == "" {
		return "", "", errors.New("server volume not initialized")
	}
	rel := strings.TrimSpace(relPath)
	if !allowEmpty && rel == "" {
		return "", "", errors.New("path is required")
	}
	if strings.ContainsRune(rel, '\x00') {
		return "", "", errors.New("invalid path")
	}
	cleanRel := strings.TrimPrefix(filepath.Clean("/"+rel), "/")
	if cleanRel == "." {
		cleanRel = ""
	}
	absRoot, err := filepath.Abs(s.VolumePath)
	if err != nil {
		return "", "", errors.New("failed to resolve root path")
	}
	absPath, err := filepath.Abs(filepath.Join(absRoot, cleanRel))
	if err != nil {
		return "", "", errors.New("invalid path")
	}
	if absPath != absRoot && !strings.HasPrefix(absPath, absRoot+string(os.PathSeparator)) {
		return "", "", errors.New("path escapes server volume")
	}
	return absRoot, absPath, nil
}

// handleListServerFiles lists files in a server volume directory.
func (h *Handler) HandleListServerFiles(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	_, absPath, err := h.resolveServerPath(id, r.URL.Query().Get("path"), true)
	if err != nil {
		jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}
	entries, err := os.ReadDir(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			jsonError(w, "path not found", http.StatusNotFound)
			return
		}
		jsonError(w, "failed to read directory", http.StatusInternalServerError)
		return
	}
	items := make([]ServerFileEntry, 0, len(entries))
	for _, e := range entries {
		info, infoErr := e.Info()
		if infoErr != nil {
			continue
		}
		itemRelPath := strings.Trim(strings.TrimPrefix(filepath.ToSlash(filepath.Join(r.URL.Query().Get("path"), e.Name())), "/"), " ")
		items = append(items, ServerFileEntry{Name: e.Name(), Path: itemRelPath, IsDir: e.IsDir(), Size: info.Size(), ModifiedAt: info.ModTime()})
	}
	slices.SortFunc(items, func(a, b ServerFileEntry) int {
		if a.IsDir != b.IsDir {
			if a.IsDir {
				return -1
			}
			return 1
		}
		return strings.Compare(strings.ToLower(a.Name), strings.ToLower(b.Name))
	})
	jsonResponse(w, map[string]any{"path": strings.Trim(strings.TrimPrefix(filepath.ToSlash(r.URL.Query().Get("path")), "/"), " "), "entries": items})
}

// handleReadServerFile reads file content for inline editor when safe.
func (h *Handler) HandleReadServerFile(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	_, absPath, err := h.resolveServerPath(id, r.URL.Query().Get("path"), false)
	if err != nil {
		jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}
	info, err := os.Stat(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			jsonError(w, "file not found", http.StatusNotFound)
			return
		}
		jsonError(w, "failed to stat file", http.StatusInternalServerError)
		return
	}
	if info.IsDir() {
		jsonError(w, "path is a directory", http.StatusBadRequest)
		return
	}
	out := map[string]any{
		"path":             strings.Trim(strings.TrimPrefix(filepath.ToSlash(r.URL.Query().Get("path")), "/"), " "),
		"size":             info.Size(),
		"maxEditableBytes": maxEditableFileBytes,
		"tooLarge":         false,
		"isBinary":         false,
		"content":          "",
	}
	if info.Size() > maxEditableFileBytes {
		out["tooLarge"] = true
		jsonResponse(w, out)
		return
	}
	data, err := os.ReadFile(absPath)
	if err != nil {
		jsonError(w, "failed to read file", http.StatusInternalServerError)
		return
	}
	if isLikelyBinary(data) {
		out["isBinary"] = true
		jsonResponse(w, out)
		return
	}
	out["content"] = string(data)
	jsonResponse(w, out)
}

// handleWriteServerFile writes text content to a file.
func (h *Handler) HandleWriteServerFile(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var req struct {
		Path    string `json:"path"`
		Content string `json:"content"`
	}
	if err := decodeJSON(r, &req); err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	_, absPath, err := h.resolveServerPath(id, req.Path, false)
	if err != nil {
		jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := os.MkdirAll(filepath.Dir(absPath), 0o755); err != nil {
		jsonError(w, "failed to create directory", http.StatusInternalServerError)
		return
	}
	if err := os.WriteFile(absPath, []byte(req.Content), 0o644); err != nil {
		jsonError(w, "failed to write file", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// handleCreateServerFile creates a new file with optional initial content.
func (h *Handler) HandleCreateServerFile(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var req struct {
		Path    string `json:"path"`
		Content string `json:"content,omitempty"`
	}
	if err := decodeJSON(r, &req); err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	_, absPath, err := h.resolveServerPath(id, req.Path, false)
	if err != nil {
		jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}
	if _, statErr := os.Stat(absPath); statErr == nil {
		jsonError(w, "path already exists", http.StatusConflict)
		return
	}
	if err := os.MkdirAll(filepath.Dir(absPath), 0o755); err != nil {
		jsonError(w, "failed to create directory", http.StatusInternalServerError)
		return
	}
	if err := os.WriteFile(absPath, []byte(req.Content), 0o644); err != nil {
		jsonError(w, "failed to create file", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// handleMkdirServerFile creates a new directory.
func (h *Handler) HandleMkdirServerFile(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var req struct {
		Path string `json:"path"`
	}
	if err := decodeJSON(r, &req); err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	_, absPath, err := h.resolveServerPath(id, req.Path, false)
	if err != nil {
		jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := os.MkdirAll(absPath, 0o755); err != nil {
		jsonError(w, "failed to create directory", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// handleMoveServerFile renames or moves a file or directory.
func (h *Handler) HandleMoveServerFile(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var req struct {
		FromPath string `json:"fromPath"`
		ToPath   string `json:"toPath"`
	}
	if err := decodeJSON(r, &req); err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	_, fromPath, err := h.resolveServerPath(id, req.FromPath, false)
	if err != nil {
		jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}
	_, toPath, err := h.resolveServerPath(id, req.ToPath, false)
	if err != nil {
		jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := os.MkdirAll(filepath.Dir(toPath), 0o755); err != nil {
		jsonError(w, "failed to create destination directory", http.StatusInternalServerError)
		return
	}
	if err := os.Rename(fromPath, toPath); err != nil {
		if os.IsNotExist(err) {
			jsonError(w, "source path not found", http.StatusNotFound)
			return
		}
		jsonError(w, "failed to move path", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// handleDeleteServerFile deletes a file or directory recursively.
func (h *Handler) HandleDeleteServerFile(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	_, absPath, err := h.resolveServerPath(id, r.URL.Query().Get("path"), false)
	if err != nil {
		jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := os.RemoveAll(absPath); err != nil {
		jsonError(w, "failed to delete path", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// DownloadServerFile streams a server file to the client.
func (h *Handler) DownloadServerFile(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	_, absPath, err := h.resolveServerPath(id, r.URL.Query().Get("path"), false)
	if err != nil {
		jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}
	info, err := os.Stat(absPath)
	if err != nil || info.IsDir() {
		jsonError(w, "file not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Disposition", "attachment; filename=\""+filepath.Base(absPath)+"\"")
	http.ServeFile(w, r, absPath)
}

// UploadServerFile uploads a file into a server directory.
func (h *Handler) UploadServerFile(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	_, absPath, err := h.resolveServerPath(id, r.URL.Query().Get("path"), true)
	if err != nil {
		jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := r.ParseMultipartForm(128 << 20); err != nil {
		jsonError(w, "invalid multipart form", http.StatusBadRequest)
		return
	}
	defer func() {
		if r.MultipartForm != nil {
			_ = r.MultipartForm.RemoveAll()
		}
	}()
	file, header, err := r.FormFile("file")
	if err != nil {
		jsonError(w, "missing file field", http.StatusBadRequest)
		return
	}
	defer func() { _ = file.Close() }()
	if err := os.MkdirAll(absPath, 0o755); err != nil {
		jsonError(w, "failed to create directory", http.StatusInternalServerError)
		return
	}
	targetPath := filepath.Join(absPath, filepath.Base(header.Filename))
	out, err := os.Create(targetPath)
	if err != nil {
		jsonError(w, "failed to create file", http.StatusInternalServerError)
		return
	}
	defer func() { _ = out.Close() }()
	if _, err := io.Copy(out, file); err != nil {
		jsonError(w, "failed to save file", http.StatusInternalServerError)
		return
	}
	jsonResponse(w, map[string]bool{"success": true})
}

func isLikelyBinary(data []byte) bool {
	if len(data) == 0 {
		return false
	}
	if len(data) > 8192 {
		data = data[:8192]
	}
	if strings.IndexByte(string(data), 0) >= 0 {
		return true
	}
	return false
}
