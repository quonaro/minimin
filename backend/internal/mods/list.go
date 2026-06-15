package mods

import (
	"os"
	"path/filepath"
	"strings"

	"orchestrator/internal/state"
)

// ListServerMods returns metadata for every .jar in the server's mods directory.
func (s *Service) ListServerMods(serverID string) ([]ModInfo, error) {
	st, ok := s.instance.Get(serverID)
	if !ok {
		return nil, ErrNotFound
	}
	if st.VolumePath == "" {
		return nil, ErrVolumeNotInitialized
	}

	modsDir := filepath.Join(st.VolumePath, "mods")
	entries, err := os.ReadDir(modsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []ModInfo{}, nil
		}
		return nil, err
	}

	var mods []ModInfo
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		lowerName := strings.ToLower(e.Name())
		isDeactivated := strings.HasSuffix(lowerName, ".deactivated")
		if !strings.HasSuffix(lowerName, ".jar") && !isDeactivated {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		modPath := filepath.Join(modsDir, e.Name())
		modInfo, _ := ParseModInfoCached(modPath, info.Size(), info.ModTime())
		if modInfo == nil {
			modInfo = &ModInfo{
				Filename:    e.Name(),
				Name:        strings.TrimSuffix(e.Name(), ".deactivated"),
				Size:        info.Size(),
				Enabled:     !isDeactivated,
				Corrupted:   true,
				InstalledAt: info.ModTime().Unix(),
			}
		} else {
			modInfo.Filename = e.Name()
			modInfo.Enabled = !isDeactivated
			modInfo.InstalledAt = info.ModTime().Unix()
			if isDeactivated && strings.HasSuffix(modInfo.Name, ".deactivated") {
				modInfo.Name = strings.TrimSuffix(modInfo.Name, ".deactivated")
			}
		}
		mods = append(mods, *modInfo)
	}
	return mods, nil
}

// CountServerMods returns the number of mods for a server (wrapper around state.CountMods).
func CountServerMods(s state.ServerState) int {
	return state.CountMods(s)
}
