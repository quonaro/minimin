package mods

import (
	"os"
	"path/filepath"
	"strings"
)

// DeleteServerMod removes a single mod .jar from the server's mods directory.
func (s *Service) DeleteServerMod(serverID, filename string) error {
	st, ok := s.instance.Get(serverID)
	if !ok {
		return ErrNotFound
	}
	if st.VolumePath == "" {
		return ErrVolumeNotInitialized
	}

	modsDir := filepath.Join(st.VolumePath, "mods")
	modPath := filepath.Join(modsDir, filename)
	if !strings.HasPrefix(modPath, modsDir+string(filepath.Separator)) && modPath != modsDir {
		return ErrInvalidFilename
	}

	if err := os.Remove(modPath); err != nil {
		if os.IsNotExist(err) {
			return ErrNotFound
		}
		return err
	}
	return nil
}

// ToggleServerMod renames a mod file between .jar and .jar.deactivated.
// It returns the new filename and the enabled state.
func (s *Service) ToggleServerMod(serverID, filename string) (string, bool, error) {
	st, ok := s.instance.Get(serverID)
	if !ok {
		return "", false, ErrNotFound
	}
	if st.VolumePath == "" {
		return "", false, ErrVolumeNotInitialized
	}

	modsDir := filepath.Join(st.VolumePath, "mods")
	modPath := filepath.Join(modsDir, filename)
	if !strings.HasPrefix(modPath, modsDir+string(filepath.Separator)) && modPath != modsDir {
		return "", false, ErrInvalidFilename
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
			return "", false, ErrNotFound
		}
		return "", false, err
	}
	return filepath.Base(newPath), enabled, nil
}
