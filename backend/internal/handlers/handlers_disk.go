package handlers

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// HandleGetServerDisk returns disk usage for a server volume and its worlds.
func (h *Handler) HandleGetServerDisk(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	s, ok := h.Instance.Get(id)
	if !ok {
		jsonError(w, "not found", http.StatusNotFound)
		return
	}
	if s.VolumePath == "" {
		jsonResponse(w, map[string]int64{
			"totalBytes":       0,
			"worldBytes":       0,
			"worldNetherBytes": 0,
			"worldEndBytes":    0,
		})
		return
	}

	var totalBytes, worldBytes, netherBytes, endBytes int64

	_ = filepath.WalkDir(s.VolumePath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return nil
		}
		size := info.Size()
		totalBytes += size

		rel, err := filepath.Rel(s.VolumePath, path)
		if err != nil {
			return nil
		}
		rel = filepath.ToSlash(rel)
		switch {
		case strings.HasPrefix(rel, "world_nether/"):
			netherBytes += size
		case strings.HasPrefix(rel, "world_the_end/"):
			endBytes += size
		case strings.HasPrefix(rel, "world/DIM-1/"):
			netherBytes += size
		case strings.HasPrefix(rel, "world/DIM1/"):
			endBytes += size
		case strings.HasPrefix(rel, "world/"):
			worldBytes += size
		}
		return nil
	})

	jsonResponse(w, map[string]int64{
		"totalBytes":       totalBytes,
		"worldBytes":       worldBytes,
		"worldNetherBytes": netherBytes,
		"worldEndBytes":    endBytes,
	})
}
