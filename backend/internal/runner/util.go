package runner

import (
	"crypto/rand"
	"fmt"
	"net"
	"path/filepath"
	"strings"
)

// GenerateRconPassword creates a 16-character random alphanumeric string.
func GenerateRconPassword() (string, error) {
	const alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	for i := range b {
		b[i] = alphabet[int(b[i])%len(alphabet)]
	}
	return string(b), nil
}

// GenerateVolumeID creates an 8-character random alphanumeric string.
func GenerateVolumeID() (string, error) {
	const alphabet = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	for i := range b {
		b[i] = alphabet[int(b[i])%len(alphabet)]
	}
	return string(b), nil
}

// IsPortFree reports whether the given TCP port is available on the host.
func IsPortFree(host string, port uint16) bool {
	addr := fmt.Sprintf("%s:%d", host, port)
	if host == "" {
		addr = fmt.Sprintf(":%d", port)
	}
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return false
	}
	_ = ln.Close()
	return true
}

// FindFreePort returns preferred if it is free on the given host,
// otherwise it searches for the next free port starting from preferred+1,
// wrapping around at 65535 to 1024.
func FindFreePort(host string, preferred uint16) (uint16, error) {
	return FindFreePortExcluding(host, preferred, func(uint16) bool { return false })
}

// FindFreePortExcluding returns preferred if it is free on the given host
// and not excluded by the filter; otherwise it searches for the next free port.
func FindFreePortExcluding(host string, preferred uint16, exclude func(uint16) bool) (uint16, error) {
	if preferred != 0 && IsPortFree(host, preferred) && !exclude(preferred) {
		return preferred, nil
	}
	start := uint16(1024)
	if preferred != 0 {
		start = preferred + 1
	}
	for p := start; p < 65535; p++ {
		if IsPortFree(host, p) && !exclude(p) {
			return p, nil
		}
	}
	for p := uint16(1024); p < start; p++ {
		if IsPortFree(host, p) && !exclude(p) {
			return p, nil
		}
	}
	return 0, fmt.Errorf("no free port found on host %q", host)
}

// HostPathForDocker translates a local path inside the backend container to the
// corresponding host path so that Docker daemon binds the correct directory.
func HostPathForDocker(localPath, serversDir, serversHostDir string) string {
	rel, err := filepath.Rel(serversDir, localPath)
	if err != nil || strings.HasPrefix(rel, "..") {
		return localPath
	}
	return filepath.Join(serversHostDir, rel)
}

// SplitEnv splits a KEY=VALUE string into its components.
func SplitEnv(s string) (key, value string, ok bool) {
	for i := 0; i < len(s); i++ {
		if s[i] == '=' {
			return s[:i], s[i+1:], true
		}
	}
	return "", "", false
}
