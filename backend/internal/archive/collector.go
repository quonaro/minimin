package archive

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"orchestrator/internal/state"
)

func collectClientFiles(instance *state.InstanceFile, serverID string, include []string) (map[string][]string, error) {
	s, ok := instance.Get(serverID)
	if !ok || s.VolumePath == "" {
		return nil, fmt.Errorf("server volume not initialized")
	}

	typeDirMap := map[string]string{
		"mods":          filepath.Join(s.VolumePath, "mods-client"),
		"resourcepacks": filepath.Join(s.VolumePath, "resourcepacks"),
		"shaderpacks":   filepath.Join(s.VolumePath, "shaderpacks"),
	}
	extMap := map[string]string{
		"mods":          ".jar",
		"resourcepacks": ".zip",
		"shaderpacks":   ".zip",
	}

	filesByType := make(map[string][]string)
	for _, t := range include {
		dir, ok := typeDirMap[t]
		if !ok {
			continue
		}
		ext := extMap[t]
		entries, err := os.ReadDir(dir)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, fmt.Errorf("failed to read %s directory: %w", t, err)
		}
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			name := strings.ToLower(e.Name())
			if strings.HasSuffix(name, ext) || strings.HasSuffix(name, ext+".disabled") {
				filesByType[t] = append(filesByType[t], filepath.Join(dir, e.Name()))
			}
		}
	}

	return filesByType, nil
}
