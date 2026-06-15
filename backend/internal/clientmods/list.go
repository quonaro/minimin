package clientmods

import (
	"archive/zip"
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

	info, _ := mods.ParseModInfo(modPath, 0)

	zr, err := zip.OpenReader(modPath)
	if err != nil {
		return nil, "", fmt.Errorf("failed to open jar: %w", err)
	}

	iconPath := ""
	if info != nil {
		iconPath = info.Icon
	}
	if iconPath == "" {
		for _, f := range zr.File {
			name := strings.ToLower(f.Name)
			if name == "icon.png" || name == "icon.jpg" || name == "icon.jpeg" {
				iconPath = f.Name
				break
			}
		}
	}
	if iconPath == "" {
		return nil, "", mods.ErrNotFound
	}

	for _, f := range zr.File {
		if f.Name == iconPath {
			rc, err := f.Open()
			if err != nil {
				return nil, "", fmt.Errorf("failed to read icon: %w", err)
			}
			contentType := "image/png"
			lower := strings.ToLower(iconPath)
			if strings.HasSuffix(lower, ".jpg") || strings.HasSuffix(lower, ".jpeg") {
				contentType = "image/jpeg"
			}
			return &mods.ZipReadCloser{RC: rc, ZR: zr}, contentType, nil
		}
	}
	return nil, "", mods.ErrNotFound
}
