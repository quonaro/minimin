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
		parts := strings.SplitN(rel, "/", 2)
		if len(parts) > 0 {
			switch parts[0] {
			case "world":
				worldBytes += size
			case "world_nether":
				netherBytes += size
			case "world_the_end":
				endBytes += size
			}
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
