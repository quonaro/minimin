package handlers

import (
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// CrashReportEntry describes a single crash-report file.
type CrashReportEntry struct {
	Name       string    `json:"name"`
	Size       int64     `json:"size"`
	ModifiedAt time.Time `json:"modifiedAt"`
}

// HandleListCrashReports returns all crash-report files for a server.
func (h *Handler) HandleListCrashReports(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	s, ok := h.Instance.Get(id)
	if !ok {
		jsonError(w, "not found", http.StatusNotFound)
		return
	}
	if s.VolumePath == "" {
		jsonResponse(w, map[string]any{"reports": []CrashReportEntry{}})
		return
	}
	reportsDir := filepath.Join(s.VolumePath, "crash-reports")
	entries, err := os.ReadDir(reportsDir)
	if err != nil {
		if os.IsNotExist(err) {
			jsonResponse(w, map[string]any{"reports": []CrashReportEntry{}})
			return
		}
		jsonError(w, "failed to read crash-reports directory", http.StatusInternalServerError)
		return
	}
	items := make([]CrashReportEntry, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		items = append(items, CrashReportEntry{
			Name:       e.Name(),
			Size:       info.Size(),
			ModifiedAt: info.ModTime(),
		})
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].ModifiedAt.After(items[j].ModifiedAt)
	})
	jsonResponse(w, map[string]any{"reports": items})
}

// HandleReadCrashReport returns the plain-text contents of a crash-report file.
func (h *Handler) HandleReadCrashReport(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	filename := r.PathValue("filename")
	if filename == "" {
		jsonError(w, "filename is required", http.StatusBadRequest)
		return
	}
	s, ok := h.Instance.Get(id)
	if !ok {
		jsonError(w, "not found", http.StatusNotFound)
		return
	}
	if s.VolumePath == "" {
		jsonError(w, "server volume not initialized", http.StatusConflict)
		return
	}
	reportsDir := filepath.Join(s.VolumePath, "crash-reports")
	absDir, err := filepath.Abs(reportsDir)
	if err != nil {
		jsonError(w, "failed to resolve crash-reports path", http.StatusInternalServerError)
		return
	}
	absPath, err := filepath.Abs(filepath.Join(absDir, filename))
	if err != nil {
		jsonError(w, "invalid filename", http.StatusBadRequest)
		return
	}
	if !strings.HasPrefix(absPath, absDir+string(os.PathSeparator)) && absPath != absDir {
		jsonError(w, "invalid filename", http.StatusBadRequest)
		return
	}
	data, err := os.ReadFile(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			jsonError(w, "file not found", http.StatusNotFound)
			return
		}
		jsonError(w, "failed to read file", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}

// HandleDeleteCrashReport removes a crash-report file.
func (h *Handler) HandleDeleteCrashReport(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	filename := r.PathValue("filename")
	if filename == "" {
		jsonError(w, "filename is required", http.StatusBadRequest)
		return
	}
	s, ok := h.Instance.Get(id)
	if !ok {
		jsonError(w, "not found", http.StatusNotFound)
		return
	}
	if s.VolumePath == "" {
		jsonError(w, "server volume not initialized", http.StatusConflict)
		return
	}
	reportsDir := filepath.Join(s.VolumePath, "crash-reports")
	absDir, err := filepath.Abs(reportsDir)
	if err != nil {
		jsonError(w, "failed to resolve crash-reports path", http.StatusInternalServerError)
		return
	}
	absPath, err := filepath.Abs(filepath.Join(absDir, filename))
	if err != nil {
		jsonError(w, "invalid filename", http.StatusBadRequest)
		return
	}
	if !strings.HasPrefix(absPath, absDir+string(os.PathSeparator)) && absPath != absDir {
		jsonError(w, "invalid filename", http.StatusBadRequest)
		return
	}
	if err := os.Remove(absPath); err != nil {
		if os.IsNotExist(err) {
			jsonError(w, "file not found", http.StatusNotFound)
			return
		}
		jsonError(w, "failed to delete file", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
