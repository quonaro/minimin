package handlers

import (
	"context"
	"net/http"
)

// HandleStartServer starts an existing server container asynchronously.
func (h *Handler) HandleStartServer(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if _, ok := h.Instance.Get(id); !ok {
		jsonError(w, "not found", http.StatusNotFound)
		return
	}
	s, ok := h.Instance.TrySetDesired(id, "running")
	if !ok {
		jsonError(w, "operation in progress", http.StatusConflict)
		return
	}
	_ = h.Instance.Save()
	var req struct {
		RemoveExisting bool `json:"removeExisting"`
	}
	// Empty body is fine — default to false.
	_ = decodeJSON(r, &req)
	ctx, cancel := context.WithCancel(context.Background())
	h.Actions.RegisterStartCancel(id, cancel)
	go func() {
		defer h.Actions.UnregisterStartCancel(id)
		h.Actions.Start(ctx, id, req.RemoveExisting)
	}()
	jsonResponse(w, s)
}

// HandleStopServer stops a running server container asynchronously.
func (h *Handler) HandleStopServer(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if _, ok := h.Instance.Get(id); !ok {
		jsonError(w, "not found", http.StatusNotFound)
		return
	}
	h.Actions.CancelStart(id)
	s, ok := h.Instance.TrySetDesired(id, "exited")
	if !ok {
		jsonError(w, "operation in progress", http.StatusConflict)
		return
	}
	_ = h.Instance.Save()
	go h.Actions.Stop(context.Background(), id)
	jsonResponse(w, s)
}

// HandleForceStopServer force-stops a running server container asynchronously.
func (h *Handler) HandleForceStopServer(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if _, ok := h.Instance.Get(id); !ok {
		jsonError(w, "not found", http.StatusNotFound)
		return
	}
	h.Actions.CancelStart(id)
	s, ok := h.Instance.TrySetDesired(id, "exited")
	if !ok {
		jsonError(w, "operation in progress", http.StatusConflict)
		return
	}
	_ = h.Instance.Save()
	go h.Actions.ForceStop(context.Background(), id)
	jsonResponse(w, s)
}

// HandleRestartServer restarts a server container asynchronously.
func (h *Handler) HandleRestartServer(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if _, ok := h.Instance.Get(id); !ok {
		jsonError(w, "not found", http.StatusNotFound)
		return
	}
	s, ok := h.Instance.TrySetDesired(id, "running")
	if !ok {
		jsonError(w, "operation in progress", http.StatusConflict)
		return
	}
	_ = h.Instance.Save()
	go h.Actions.Restart(context.Background(), id)
	jsonResponse(w, s)
}

// HandleDeleteServer removes a server container and optionally wipes its data.
func (h *Handler) HandleDeleteServer(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	wipe := r.URL.Query().Get("wipe") == "true"
	if err := h.Actions.Delete(r.Context(), id, wipe); err != nil {
		jsonError(w, err.Error(), http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// HandleRecreateWorld stops the server, deletes world data, and restarts it.
func (h *Handler) HandleRecreateWorld(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if _, ok := h.Instance.Get(id); !ok {
		jsonError(w, "not found", http.StatusNotFound)
		return
	}
	_, ok := h.Instance.TrySetDesired(id, "running")
	if !ok {
		jsonError(w, "operation in progress", http.StatusConflict)
		return
	}
	_ = h.Instance.Save()
	go h.Actions.RecreateWorld(context.Background(), id)
	w.WriteHeader(http.StatusNoContent)
}

// HandleGetPullProgress returns the current image pull progress for a server.
func (h *Handler) HandleGetPullProgress(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	prog := h.Actions.GetPullProgress(id)
	if prog == nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	jsonResponse(w, prog)
}
