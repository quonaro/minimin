package handlers

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"orchestrator/internal/instances"
	"orchestrator/internal/runner"
	"orchestrator/internal/state"
)

func (h *Handler) maxInstanceUploadMB() int {
	if h.ModUploadMaxMB > 0 {
		return h.ModUploadMaxMB
	}
	return 1024
}

func parseInt64(s string, fallback int64) int64 {
	if s == "" {
		return fallback
	}
	v, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return fallback
	}
	return v
}

func parseUint16(s string, fallback uint16) uint16 {
	if s == "" {
		return fallback
	}
	v, err := strconv.ParseUint(s, 10, 16)
	if err != nil {
		return fallback
	}
	return uint16(v)
}

// HandlePrepareInstance accepts an uploaded archive, stores it temporarily and
// returns detected metadata so the UI can pre-fill engine/version fields.
func (h *Handler) HandlePrepareInstance(w http.ResponseWriter, r *http.Request) {
	maxMB := h.maxInstanceUploadMB()
	maxSize := int64(maxMB) << 20
	r.Body = http.MaxBytesReader(w, r.Body, maxSize)
	if err := r.ParseMultipartForm(maxSize); err != nil {
		jsonError(w, fmt.Sprintf("invalid multipart form or file too large: %s", err.Error()), http.StatusBadRequest)
		return
	}
	defer func() {
		if r.MultipartForm != nil {
			_ = r.MultipartForm.RemoveAll()
		}
	}()

	file, header, err := r.FormFile("file")
	if err != nil {
		jsonError(w, "missing file field", http.StatusBadRequest)
		return
	}
	defer func() { _ = file.Close() }()

	token, err := h.InstanceService.Save(header.Filename, header.Size, file)
	if err != nil {
		jsonError(w, fmt.Sprintf("failed to store archive: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	meta, err := h.InstanceService.Prepare(token)
	if err != nil {
		_ = h.InstanceService.Remove(token)
		jsonError(w, fmt.Sprintf("failed to parse archive: %s", err.Error()), http.StatusBadRequest)
		return
	}

	jsonResponse(w, map[string]any{
		"token":         token,
		"format":        meta.Format,
		"instanceName":  meta.InstanceName,
		"gameVersion":   meta.GameVersion,
		"engineType":    meta.EngineType,
		"loaderVersion": meta.LoaderVersion,
		"detectedPaths": meta.DetectedPaths,
		"worlds":        meta.Worlds,
	})
}

// HandleCreateServerFromInstance creates a server and pre-populates its volume
// from a previously uploaded archive.
func (h *Handler) HandleCreateServerFromInstance(w http.ResponseWriter, r *http.Request) {
	maxMB := h.maxInstanceUploadMB()
	maxSize := int64(maxMB) << 20
	r.Body = http.MaxBytesReader(w, r.Body, maxSize)
	if err := r.ParseMultipartForm(maxSize); err != nil {
		jsonError(w, fmt.Sprintf("invalid multipart form or file too large: %s", err.Error()), http.StatusBadRequest)
		return
	}
	defer func() {
		if r.MultipartForm != nil {
			_ = r.MultipartForm.RemoveAll()
		}
	}()

	token := r.FormValue("token")
	if token == "" {
		jsonError(w, "missing token", http.StatusBadRequest)
		return
	}

	serverID := r.FormValue("serverId")
	ramBytes := parseInt64(r.FormValue("ramBytes"), 2*1024*1024*1024)
	gamePort := parseUint16(r.FormValue("gamePort"), 25565)
	engineType := r.FormValue("engineType")
	gameVersion := r.FormValue("gameVersion")
	loaderVersion := r.FormValue("loaderVersion")
	rconPort := parseUint16(r.FormValue("rconPort"), 0)
	publicRcon := r.FormValue("publicRcon") == "true"
	levelName := r.FormValue("levelName")
	levelSeed := r.FormValue("levelSeed")
	levelType := r.FormValue("levelType")
	worldPath := r.FormValue("world")

	if serverID == "" {
		id, err := runner.GenerateVolumeID()
		if err != nil {
			jsonError(w, "failed to generate server id", http.StatusInternalServerError)
			return
		}
		serverID = id
	} else if !isValidServerID(serverID) {
		jsonError(w, "invalid server id: must be 1-64 chars of a-z, A-Z, 0-9, _ or -", http.StatusBadRequest)
		return
	}
	if ramBytes == 0 {
		ramBytes = 2 * 1024 * 1024 * 1024
	}
	if gamePort == 0 {
		gamePort = 25565
	}
	if engineType == "" {
		engineType = "VANILLA"
	}
	if gameVersion == "" {
		gameVersion = "LATEST"
	}
	if rconPort == 0 {
		rconPort = gamePort + 10
	}

	portUsed := func(p uint16) bool {
		return h.Instance.IsPortUsed(p, "")
	}
	freeGamePort, err := runner.FindFreePortExcluding("", gamePort, portUsed)
	if err != nil {
		jsonError(w, fmt.Sprintf("no free game port: %v", err), http.StatusInternalServerError)
		return
	}
	rconHost := "127.0.0.1"
	if publicRcon {
		rconHost = ""
	}
	freeRconPort, err := runner.FindFreePortExcluding(rconHost, rconPort, portUsed)
	if err != nil {
		jsonError(w, fmt.Sprintf("no free rcon port: %v", err), http.StatusInternalServerError)
		return
	}
	if freeRconPort == freeGamePort {
		freeRconPort, err = runner.FindFreePortExcluding(rconHost, 0, portUsed)
		if err != nil {
			jsonError(w, fmt.Sprintf("no free rcon port: %v", err), http.StatusInternalServerError)
			return
		}
	}
	gamePort = freeGamePort
	rconPort = freeRconPort

	rconPassword, err := runner.GenerateRconPassword()
	if err != nil {
		jsonError(w, "failed to generate rcon password", http.StatusInternalServerError)
		return
	}

	absServersDir, err := filepath.Abs(h.ServersDir)
	if err != nil {
		jsonError(w, "failed to resolve servers directory", http.StatusInternalServerError)
		return
	}
	localPath := filepath.Join(absServersDir, serverID)
	if mkdirErr := os.MkdirAll(localPath, 0o755); mkdirErr != nil {
		jsonError(w, fmt.Sprintf("failed to create server directory: %v", mkdirErr), http.StatusInternalServerError)
		return
	}

	worldGenEnv := make(map[string]string)
	if levelName != "" {
		worldGenEnv["level-name"] = levelName
	}
	if levelSeed != "" {
		worldGenEnv["level-seed"] = levelSeed
	}
	if levelType != "" {
		worldGenEnv["level-type"] = levelType
	}

	_, err = h.InstanceService.Extract(token, instances.ExtractOptions{
		TargetDir: localPath,
		World:     worldPath,
		LevelName: levelName,
	})
	if err != nil {
		_ = os.RemoveAll(localPath)
		jsonError(w, fmt.Sprintf("failed to extract instance: %v", err), http.StatusBadRequest)
		return
	}

	uid, gid := runner.ContainerUIDGID()
	if err := os.Chown(localPath, uid, gid); err != nil {
		slog.Warn("failed to chown server data directory", "path", localPath, "uid", uid, "gid", gid, "error", err)
	}
	if err := os.Chmod(localPath, 0o775); err != nil {
		slog.Warn("failed to chmod server data directory", "path", localPath, "error", err)
	}

	containerID, volumeID, volumePath, err := runner.StartServerContainer(
		r.Context(), h.Cli, serverID,
		ramBytes, gamePort,
		engineType, gameVersion, loaderVersion,
		h.ServersDir, h.ServersHostDir,
		rconPort, rconPassword, publicRcon,
		localPath,
		worldGenEnv,
		"",
		h.NetworkName,
		nil,
	)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s := state.ServerState{
		ServerID:           serverID,
		VolumeID:           volumeID,
		VolumePath:         volumePath,
		HostPath:           runner.HostPathForDocker(volumePath, h.ServersDir, h.ServersHostDir),
		ContainerPath:      "/data",
		ContainerID:        containerID,
		RamBytes:           ramBytes,
		GamePort:           gamePort,
		EngineType:         engineType,
		GameVersion:        gameVersion,
		LoaderVersion:      loaderVersion,
		RconPassword:       rconPassword,
		RconPort:           rconPort,
		PublicRcon:         publicRcon,
		ContainerStatus:    "running",
		ContainerStartedAt: time.Now().UTC(),
		ServerStatus:       "starting",
		ModCount:           state.CountMods(state.ServerState{VolumePath: volumePath}),
		ImageName:          runner.ImageName,
	}
	h.Instance.Set(s)
	_ = h.Instance.Save()

	if err := runner.FixVolumeOwnership(r.Context(), h.Cli, runner.HostPathForDocker(volumePath, h.ServersDir, h.ServersHostDir), uid, gid); err != nil {
		slog.Warn("failed to fix volume ownership", "server_id", serverID, "error", err)
	}

	slog.Info("server created from instance",
		"server_id", serverID,
		"container_id", containerID[:12],
		"volume_id", volumeID,
		"game_port", gamePort,
		"engine", engineType,
	)
	w.WriteHeader(http.StatusCreated)
	jsonResponse(w, s)
}
