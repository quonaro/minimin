package handlers

import (
	"net/http"
	"orchestrator/internal/actionlog"
	"os"
	"strings"
	"time"
)

// HandleListBackups returns all backups for a server.
func (h *Handler) HandleListBackups(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if _, ok := h.Instance.Get(id); !ok {
		jsonError(w, "not found", http.StatusNotFound)
		return
	}
	if h.BackupService == nil {
		jsonError(w, "backup service unavailable", http.StatusServiceUnavailable)
		return
	}
	list, err := h.BackupService.List(id)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, list)
}

// HandleCreateBackup triggers a manual backup.
func (h *Handler) HandleCreateBackup(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if _, ok := h.Instance.Get(id); !ok {
		jsonError(w, "not found", http.StatusNotFound)
		return
	}
	if h.BackupService == nil {
		jsonError(w, "backup service unavailable", http.StatusServiceUnavailable)
		return
	}
	bk, err := h.BackupService.Create(id, 0, 0)
	if err != nil {
		code := http.StatusInternalServerError
		if strings.Contains(err.Error(), "world directory not found") {
			code = http.StatusBadRequest
		}
		if h.ActionLogStore != nil {
			_ = h.ActionLogStore.Append(actionlog.Entry{
				ID:        generateUUID(),
				Timestamp: time.Now().UTC(),
				ServerID:  id,
				Source:    "manual",
				Action:    "backup",
				Detail:    "backup world",
				Status:    "error",
				Message:   err.Error(),
			})
		}
		jsonError(w, err.Error(), code)
		return
	}
	if h.ActionLogStore != nil {
		_ = h.ActionLogStore.Append(actionlog.Entry{
			ID:        generateUUID(),
			Timestamp: time.Now().UTC(),
			ServerID:  id,
			Source:    "manual",
			Action:    "backup",
			Detail:    "backup world",
			Status:    "success",
			Message:   bk.Name,
		})
	}
	jsonResponse(w, bk)
}

// HandleDownloadBackup serves a backup archive.
func (h *Handler) HandleDownloadBackup(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	name := r.PathValue("name")
	if _, ok := h.Instance.Get(id); !ok {
		jsonError(w, "not found", http.StatusNotFound)
		return
	}
	if h.BackupService == nil {
		jsonError(w, "backup service unavailable", http.StatusServiceUnavailable)
		return
	}
	path := h.BackupService.DownloadPath(id, name)
	if _, err := os.Stat(path); err != nil {
		jsonError(w, "not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/gzip")
	w.Header().Set("Content-Disposition", "attachment; filename=\""+name+"\"")
	http.ServeFile(w, r, path)
}

// HandleRestoreBackup restores a backup.
func (h *Handler) HandleRestoreBackup(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	name := r.PathValue("name")
	if _, ok := h.Instance.Get(id); !ok {
		jsonError(w, "not found", http.StatusNotFound)
		return
	}
	if h.BackupService == nil {
		jsonError(w, "backup service unavailable", http.StatusServiceUnavailable)
		return
	}
	if err := h.BackupService.Restore(id, name); err != nil {
		if h.ActionLogStore != nil {
			_ = h.ActionLogStore.Append(actionlog.Entry{
				ID:        generateUUID(),
				Timestamp: time.Now().UTC(),
				ServerID:  id,
				Source:    "manual",
				Action:    "restore",
				Detail:    name,
				Status:    "error",
				Message:   err.Error(),
			})
		}
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if h.ActionLogStore != nil {
		_ = h.ActionLogStore.Append(actionlog.Entry{
			ID:        generateUUID(),
			Timestamp: time.Now().UTC(),
			ServerID:  id,
			Source:    "manual",
			Action:    "restore",
			Detail:    name,
			Status:    "success",
		})
	}
	w.WriteHeader(http.StatusNoContent)
}

// HandleDeleteBackup removes a backup archive.
func (h *Handler) HandleDeleteBackup(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	name := r.PathValue("name")
	if _, ok := h.Instance.Get(id); !ok {
		jsonError(w, "not found", http.StatusNotFound)
		return
	}
	if h.BackupService == nil {
		jsonError(w, "backup service unavailable", http.StatusServiceUnavailable)
		return
	}
	if err := h.BackupService.Delete(id, name); err != nil {
		if h.ActionLogStore != nil {
			_ = h.ActionLogStore.Append(actionlog.Entry{
				ID:        generateUUID(),
				Timestamp: time.Now().UTC(),
				ServerID:  id,
				Source:    "manual",
				Action:    "delete_backup",
				Detail:    name,
				Status:    "error",
				Message:   err.Error(),
			})
		}
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if h.ActionLogStore != nil {
		_ = h.ActionLogStore.Append(actionlog.Entry{
			ID:        generateUUID(),
			Timestamp: time.Now().UTC(),
			ServerID:  id,
			Source:    "manual",
			Action:    "delete_backup",
			Detail:    name,
			Status:    "success",
		})
	}
	w.WriteHeader(http.StatusNoContent)
}
