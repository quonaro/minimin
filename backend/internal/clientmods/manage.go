package clientmods

import (
	"os"
	"path/filepath"
	"strings"

	"orchestrator/internal/mods"
)

// DeleteClientMod removes a single mod .jar from the server's mods-client directory.
func (s *Service) DeleteClientMod(serverID, filename string) error {
	modsDir, err := s.getClientModsDir(serverID)
	if err != nil {
		return err
	}
	if modsDir == "" {
		return mods.ErrNotFound
	}

	modPath := filepath.Join(modsDir, filename)
	if !strings.HasPrefix(modPath, modsDir+string(filepath.Separator)) && modPath != modsDir {
		return mods.ErrInvalidFilename
	}

	if err := os.Remove(modPath); err != nil {
		if os.IsNotExist(err) {
			return mods.ErrNotFound
		}
		return err
	}
	return nil
}

// ToggleClientMod renames a mod file between .jar and .jar.deactivated in mods-client.
func (s *Service) ToggleClientMod(serverID, filename string) (string, bool, error) {
	modsDir, err := s.getClientModsDir(serverID)
	if err != nil {
		return "", false, err
	}
	if modsDir == "" {
		return "", false, mods.ErrNotFound
	}

	modPath := filepath.Join(modsDir, filename)
	if !strings.HasPrefix(modPath, modsDir+string(filepath.Separator)) && modPath != modsDir {
		return "", false, mods.ErrInvalidFilename
	}

	lowerPath := strings.ToLower(modPath)
	var newPath string
	enabled := false
	if strings.HasSuffix(lowerPath, ".deactivated") {
		newPath = strings.TrimSuffix(modPath, ".deactivated")
		enabled = true
	} else {
		newPath = modPath + ".deactivated"
	}

	if err := os.Rename(modPath, newPath); err != nil {
		if os.IsNotExist(err) {
			return "", false, mods.ErrNotFound
		}
		return "", false, err
	}

	return filepath.Base(newPath), enabled, nil
}
