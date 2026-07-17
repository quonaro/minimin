package handlers

import (
	"net/http"
	"time"

	"orchestrator/internal/actionlog"
	"orchestrator/internal/scheduler"
)

// HandleListTasks returns all tasks for a server.
func (h *Handler) HandleListTasks(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if _, ok := h.Instance.Get(id); !ok {
		jsonError(w, "not found", http.StatusNotFound)
		return
	}
	tasks := h.Scheduler.TasksForServer(id)
	jsonResponse(w, tasks)
}

// HandleCreateTask adds a new task.
func (h *Handler) HandleCreateTask(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if _, ok := h.Instance.Get(id); !ok {
		jsonError(w, "not found", http.StatusNotFound)
		return
	}
	var req scheduler.Task
	if err := decodeJSON(r, &req); err != nil {
		jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}
	req.ServerID = id
	if req.ID == "" {
		req.ID = generateUUID()
	}
	req.LastRun = time.Time{}
	req.NextRun = time.Time{}

	h.Scheduler.Store.Put(req)
	if err := h.Scheduler.Store.Save(); err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	h.Scheduler.Reload()
	jsonResponse(w, req)
}

// HandleUpdateTask modifies an existing task.
func (h *Handler) HandleUpdateTask(w http.ResponseWriter, r *http.Request) {
	serverID := r.PathValue("id")
	taskID := r.PathValue("taskId")
	if _, ok := h.Instance.Get(serverID); !ok {
		jsonError(w, "not found", http.StatusNotFound)
		return
	}
	var req scheduler.Task
	if err := decodeJSON(r, &req); err != nil {
		jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}
	existing, ok := h.Scheduler.Store.Get(taskID)
	if !ok || existing.ServerID != serverID {
		jsonError(w, "not found", http.StatusNotFound)
		return
	}
	req.ID = taskID
	req.ServerID = serverID
	req.LastRun = existing.LastRun
	req.NextRun = existing.NextRun

	h.Scheduler.Store.Put(req)
	if err := h.Scheduler.Store.Save(); err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	h.Scheduler.Reload()
	jsonResponse(w, req)
}

// HandleDeleteTask removes a task.
func (h *Handler) HandleDeleteTask(w http.ResponseWriter, r *http.Request) {
	serverID := r.PathValue("id")
	taskID := r.PathValue("taskId")
	if _, ok := h.Instance.Get(serverID); !ok {
		jsonError(w, "not found", http.StatusNotFound)
		return
	}
	existing, ok := h.Scheduler.Store.Get(taskID)
	if !ok || existing.ServerID != serverID {
		jsonError(w, "not found", http.StatusNotFound)
		return
	}
	h.Scheduler.Store.Delete(taskID)
	if err := h.Scheduler.Store.Save(); err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	h.Scheduler.Reload()
	w.WriteHeader(http.StatusNoContent)
}

// HandleRunTask triggers a task immediately.
func (h *Handler) HandleRunTask(w http.ResponseWriter, r *http.Request) {
	serverID := r.PathValue("id")
	taskID := r.PathValue("taskId")
	if _, ok := h.Instance.Get(serverID); !ok {
		jsonError(w, "not found", http.StatusNotFound)
		return
	}
	existing, ok := h.Scheduler.Store.Get(taskID)
	if !ok || existing.ServerID != serverID {
		jsonError(w, "not found", http.StatusNotFound)
		return
	}
	if err := h.Scheduler.RunNow(taskID); err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// HandleListActionLog returns the action log for a server.
func (h *Handler) HandleListActionLog(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if _, ok := h.Instance.Get(id); !ok {
		jsonError(w, "not found", http.StatusNotFound)
		return
	}
	if h.ActionLogStore == nil {
		jsonResponse(w, []actionlog.Entry{})
		return
	}
	entries, err := h.ActionLogStore.List(id)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, entries)
}
