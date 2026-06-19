package instances

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
)

// Service provides a high-level API for importing instance archives.
type Service struct {
	store *Store
}

// NewService creates an instance import service backed by the given store.
func NewService(store *Store) *Service {
	return &Service{store: store}
}

// Prepare reads a previously stored archive by token, detects its format and
// returns parsed metadata without extracting anything.
func (s *Service) Prepare(token string) (*Metadata, error) {
	entry, err := s.store.Get(token)
	if err != nil {
		return nil, err
	}
	zr, size, err := openZip(entry.Path)
	if err != nil {
		return nil, err
	}
	defer zr.Close()

	format, err := DetectFormat(zr.Reader)
	if err != nil {
		return nil, fmt.Errorf("detect format: %w", err)
	}

	meta, err := ParseMetadata(zr.Reader, format)
	if err != nil {
		return nil, err
	}
	meta.DetectedPaths = DetectPaths(zr.Reader, format)
	meta.Worlds = DetectWorlds(zr.Reader, format)
	_ = size
	return meta, nil
}

// Save stores an uploaded archive in the temporary store and returns its token.
func (s *Service) Save(name string, size int64, src io.Reader) (string, error) {
	return s.store.Save(name, size, func(dst *os.File) error {
		_, err := io.Copy(dst, src)
		return err
	})
}

// Extract token expands a previously stored archive into the target directory.
// opts controls which directories and which world are imported.
func (s *Service) Extract(token string, opts ExtractOptions) (*Metadata, error) {
	entry, err := s.store.Get(token)
	if err != nil {
		return nil, err
	}
	zr, _, err := openZip(entry.Path)
	if err != nil {
		return nil, err
	}
	defer zr.Close()

	format, err := DetectFormat(zr.Reader)
	if err != nil {
		return nil, fmt.Errorf("detect format: %w", err)
	}

	meta, err := ParseMetadata(zr.Reader, format)
	if err != nil {
		return nil, err
	}

	if err := Extract(zr.Reader, format, opts); err != nil {
		return nil, err
	}

	_ = s.store.Remove(token)
	return meta, nil
}

// Remove deletes a prepared archive.
func (s *Service) Remove(token string) error {
	return s.store.Remove(token)
}

// Cleanup removes expired entries from the store.
func (s *Service) Cleanup() {
	s.store.Cleanup()
}

// closeableReader wraps a zip.Reader with a closable file handle.
type closeableReader struct {
	*zip.Reader
	f *os.File
}

// Close closes the underlying file handle.
func (c *closeableReader) Close() error {
	if c.f == nil {
		return nil
	}
	return c.f.Close()
}

// openZip opens a zip file and returns a reader with a closeable handle.
func openZip(path string) (*closeableReader, int64, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, 0, fmt.Errorf("open archive: %w", err)
	}
	info, err := f.Stat()
	if err != nil {
		_ = f.Close()
		return nil, 0, fmt.Errorf("stat archive: %w", err)
	}
	zr, err := zip.NewReader(f, info.Size())
	if err != nil {
		_ = f.Close()
		return nil, 0, fmt.Errorf("read archive: %w", err)
	}
	return &closeableReader{Reader: zr, f: f}, info.Size(), nil
}
