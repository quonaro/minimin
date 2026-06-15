package archive

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"orchestrator/internal/state"
)

// Service provides archive generation and management.
type Service struct {
	instance *state.InstanceFile
}

// NewService creates a new archive service.
func NewService(instance *state.InstanceFile) *Service {
	return &Service{instance: instance}
}

func generateToken() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func computeSHA256(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer func() { _ = f.Close() }()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// Create generates a new archive token for the given server.
func (s *Service) Create(serverID string, req CreateRequest) (*Token, error) {
	st, ok := s.instance.Get(serverID)
	if !ok {
		return nil, fmt.Errorf("server not found")
	}
	if st.VolumePath == "" {
		return nil, fmt.Errorf("server volume not initialized")
	}

	if req.TTL <= 0 {
		req.TTL = 24
	}
	const maxTTL = 8760 // 12 months in hours
	if req.TTL > maxTTL {
		req.TTL = maxTTL
	}
	if len(req.Include) == 0 {
		req.Include = []string{"mods", "resourcepacks", "shaderpacks"}
	}

	token := generateToken()
	expiresAt := time.Now().Add(time.Duration(req.TTL) * time.Hour)

	archive := &Token{
		Token:          token,
		ServerID:       serverID,
		ServerName:     st.ServerID,
		Include:        req.Include,
		ExpiresAt:      expiresAt,
		CreatedAt:      time.Now(),
		DownloadCounts: make(map[string]int),
	}

	tokensMu.Lock()
	tokens[token] = archive
	tokensMu.Unlock()

	saveArchiveMeta(s.instance, serverID, archive)
	return archive, nil
}

// Get retrieves an archive token by its token string.
func (s *Service) Get(token string) (*Token, error) {
	tokensMu.RLock()
	archive, ok := tokens[token]
	tokensMu.RUnlock()

	if !ok || time.Now().After(archive.ExpiresAt) {
		return nil, fmt.Errorf("archive not found or expired")
	}
	return archive, nil
}

// Download streams the archive in the requested format to w.
func (s *Service) Download(tokenStr, format, baseURL string, w io.Writer) error {
	archive, err := s.Get(tokenStr)
	if err != nil {
		return err
	}

	filesByType, err := collectClientFiles(s.instance, archive.ServerID, archive.Include)
	if err != nil {
		return err
	}

	srv, ok := s.instance.Get(archive.ServerID)
	if !ok || srv.VolumePath == "" {
		return fmt.Errorf("server not found or volume not initialized")
	}

	var genErr error
	switch format {
	case "zip":
		genErr = createZipArchive(w, filesByType)
	case "mrpack":
		genErr = createMrpackArchive(w, archive.ServerName, filesByType)
	case "curseforge":
		genErr = createCurseForgeArchive(w, archive.ServerName, srv.GameVersion, srv.EngineType, srv.LoaderVersion, filesByType)
	case "prism":
		genErr = createPrismArchive(w, archive.ServerName, srv.GameVersion, srv.EngineType, srv.LoaderVersion, archive.Token, baseURL, filesByType)
	default:
		return fmt.Errorf("invalid format")
	}

	if genErr != nil {
		return genErr
	}

	tokensMu.Lock()
	archive.TotalDownloads++
	archive.DownloadCounts[format]++
	tokensMu.Unlock()

	saveArchiveMeta(s.instance, archive.ServerID, archive)
	return nil
}

// Manifest returns a manifest of all files in the archive with SHA256 hashes.
func (s *Service) Manifest(tokenStr string) (map[string]any, error) {
	archive, err := s.Get(tokenStr)
	if err != nil {
		return nil, err
	}

	filesByType, err := collectClientFiles(s.instance, archive.ServerID, archive.Include)
	if err != nil {
		return nil, err
	}

	type manifestFile struct {
		Path   string `json:"path"`
		SHA256 string `json:"sha256"`
		Size   int64  `json:"size"`
	}

	dirMap := map[string]string{
		"mods":          ".minecraft/mods/",
		"resourcepacks": ".minecraft/resourcepacks/",
		"shaderpacks":   ".minecraft/shaderpacks/",
	}

	var files []manifestFile
	for t, paths := range filesByType {
		prefix := dirMap[t]
		for _, p := range paths {
			info, err := os.Stat(p)
			if err != nil {
				continue
			}
			hash, err := computeSHA256(p)
			if err != nil {
				continue
			}
			files = append(files, manifestFile{
				Path:   prefix + filepath.Base(p),
				SHA256: hash,
				Size:   info.Size(),
			})
		}
	}

	return map[string]any{
		"serverName": archive.ServerName,
		"files":      files,
	}, nil
}

// DownloadFile streams a single file from an archive.
func (s *Service) DownloadFile(tokenStr, filePath string, w http.ResponseWriter) error {
	archive, err := s.Get(tokenStr)
	if err != nil {
		return err
	}

	srv, ok := s.instance.Get(archive.ServerID)
	if !ok || srv.VolumePath == "" {
		return fmt.Errorf("server volume not initialized")
	}

	var diskPath string
	parts := strings.SplitN(filePath, "/", 3)
	if len(parts) >= 3 && parts[0] == ".minecraft" {
		switch parts[1] {
		case "mods":
			diskPath = filepath.Join(srv.VolumePath, "mods-client", parts[2])
		case "resourcepacks":
			diskPath = filepath.Join(srv.VolumePath, "resourcepacks", parts[2])
		case "shaderpacks":
			diskPath = filepath.Join(srv.VolumePath, "shaderpacks", parts[2])
		}
	}

	if diskPath == "" {
		return fmt.Errorf("invalid file path")
	}

	f, err := os.Open(diskPath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("file not found")
		}
		return err
	}
	defer func() { _ = f.Close() }()

	info, err := f.Stat()
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filepath.Base(diskPath)))
	w.Header().Set("Content-Length", fmt.Sprintf("%d", info.Size()))
	w.WriteHeader(http.StatusOK)
	_, _ = io.Copy(w, f)
	return nil
}

// List returns all active archive tokens for a server.
func (s *Service) List(serverID string) []Summary {
	tokensMu.RLock()
	defer tokensMu.RUnlock()

	now := time.Now()
	var results []Summary
	for _, a := range tokens {
		if a.ServerID != serverID {
			continue
		}
		if now.After(a.ExpiresAt) {
			continue
		}
		results = append(results, Summary{
			Token:          a.Token,
			ServerName:     a.ServerName,
			ExpiresAt:      a.ExpiresAt,
			CreatedAt:      a.CreatedAt,
			Formats:        []string{"zip", "mrpack", "curseforge", "prism"},
			DownloadCounts: a.DownloadCounts,
			TotalDownloads: a.TotalDownloads,
		})
	}
	return results
}

// Delete removes an archive token.
func (s *Service) Delete(serverID, tokenStr string) error {
	tokensMu.Lock()
	archive, ok := tokens[tokenStr]
	if !ok || archive.ServerID != serverID {
		tokensMu.Unlock()
		return fmt.Errorf("not found")
	}
	delete(tokens, tokenStr)
	tokensMu.Unlock()

	removeArchiveMeta(s.instance, serverID, tokenStr)
	return nil
}
