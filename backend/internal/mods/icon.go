package mods

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// ZipReadCloser wraps a zip file entry reader and the parent zip.ReadCloser so
// the underlying jar stays open until the icon has been fully streamed.
type ZipReadCloser struct {
	RC io.ReadCloser
	ZR *zip.ReadCloser
}

func (z *ZipReadCloser) Read(p []byte) (int, error) {
	return z.RC.Read(p)
}

func (z *ZipReadCloser) Close() error {
	_ = z.RC.Close()
	return z.ZR.Close()
}

// GetServerModIcon opens a mod jar and returns the icon file reader and content type.
func (s *Service) GetServerModIcon(serverID, filename string) (io.ReadCloser, string, error) {
	st, ok := s.instance.Get(serverID)
	if !ok {
		return nil, "", ErrNotFound
	}
	if st.VolumePath == "" {
		return nil, "", ErrVolumeNotInitialized
	}

	modsDir := filepath.Join(st.VolumePath, "mods")
	modPath := filepath.Join(modsDir, filename)
	if !strings.HasPrefix(modPath, modsDir+string(filepath.Separator)) && modPath != modsDir {
		return nil, "", ErrInvalidFilename
	}

	var info *ModInfo
	if fi, statErr := os.Stat(modPath); statErr == nil {
		info, _ = ParseModInfoCached(modPath, fi.Size(), fi.ModTime())
	}

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
		return nil, "", ErrNotFound
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
			return &ZipReadCloser{RC: rc, ZR: zr}, contentType, nil
		}
	}
	return nil, "", ErrNotFound
}

// GetServerIcon returns the server-icon.png reader.
func (s *Service) GetServerIcon(serverID string) (io.ReadCloser, error) {
	st, ok := s.instance.Get(serverID)
	if !ok {
		return nil, ErrNotFound
	}
	iconPath := filepath.Join(st.VolumePath, "server-icon.png")
	f, err := os.Open(iconPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return f, nil
}

// UploadServerIcon writes an image as server-icon.png into the server's volume.
func (s *Service) UploadServerIcon(serverID string, file io.Reader) error {
	st, ok := s.instance.Get(serverID)
	if !ok {
		return ErrNotFound
	}
	if st.VolumePath == "" {
		return ErrVolumeNotInitialized
	}

	iconPath := filepath.Join(st.VolumePath, "server-icon.png")
	if err := os.MkdirAll(st.VolumePath, 0o755); err != nil {
		return err
	}
	f, err := os.Create(iconPath)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()
	if _, err := io.Copy(f, file); err != nil {
		return err
	}
	return nil
}
