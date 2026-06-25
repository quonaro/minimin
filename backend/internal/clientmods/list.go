package clientmods

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"orchestrator/internal/mods"
)

// ListClientMods returns metadata for every .jar in the server's mods-client directory.
func (s *Service) ListClientMods(serverID string) ([]mods.ModInfo, error) {
	modsDir, err := s.getClientModsDir(serverID)
	if err != nil {
		return nil, err
	}
	if modsDir == "" {
		return nil, mods.ErrNotFound
	}

	entries, err := os.ReadDir(modsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []mods.ModInfo{}, nil
		}
		return nil, err
	}

	var result []mods.ModInfo
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
		modInfo, _ := mods.ParseModInfoCached(modPath, info.Size(), info.ModTime())
		if modInfo == nil {
			modInfo = &mods.ModInfo{
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
		result = append(result, *modInfo)
	}
	return result, nil
}

// GetClientModIcon opens a client mod jar and returns the icon reader and content type.
func (s *Service) GetClientModIcon(serverID, filename string) (io.ReadCloser, string, error) {
	st, ok := s.instance.Get(serverID)
	if !ok {
		return nil, "", mods.ErrNotFound
	}
	if st.VolumePath == "" {
		return nil, "", mods.ErrNotFound
	}

	modsDir := filepath.Join(st.VolumePath, "mods-client")
	modPath := filepath.Join(modsDir, filename)
	if !strings.HasPrefix(modPath, modsDir+string(filepath.Separator)) && modPath != modsDir {
		return nil, "", mods.ErrInvalidFilename
	}

	data, contentType, err := mods.ExtractModIcon(modPath)
	if err != nil {
		return nil, "", err
	}
	return io.NopCloser(bytes.NewReader(data)), contentType, nil
}

// GetClientModIconsBatch returns a map of filename -> data URL for multiple client mod icons.
func (s *Service) GetClientModIconsBatch(serverID string, filenames []string) (map[string]string, error) {
	st, ok := s.instance.Get(serverID)
	if !ok {
		return nil, mods.ErrNotFound
	}
	if st.VolumePath == "" {
		return nil, mods.ErrNotFound
	}

	modsDir := filepath.Join(st.VolumePath, "mods-client")
	result := make(map[string]string, len(filenames))
	for _, filename := range filenames {
		modPath := filepath.Join(modsDir, filename)
		if !strings.HasPrefix(modPath, modsDir+string(filepath.Separator)) && modPath != modsDir {
			continue
		}
		data, contentType, err := mods.ExtractModIcon(modPath)
		if err != nil {
			continue
		}
		b64 := base64.StdEncoding.EncodeToString(data)
		result[filename] = fmt.Sprintf("data:%s;base64,%s", contentType, b64)
	}
	return result, nil
}
