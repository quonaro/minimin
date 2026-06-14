package handlers

import (
	"fmt"
	"log/slog"
	"net/http"
	"regexp"
	"time"

	"orchestrator/internal/runner"
	"orchestrator/internal/state"
)

var serverIDRe = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

func isValidServerID(id string) bool {
	return serverIDRe.MatchString(id) && len(id) <= 64
}

// handleCreateServer spawns a new Minecraft server container.
func (h *Handler) HandleCreateServer(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ServerID         string   `json:"serverId,omitempty"`
		RamBytes         int64    `json:"ramBytes,omitempty"`
		GamePort         uint16   `json:"gamePort,omitempty"`
		EngineType       string   `json:"engineType,omitempty"`
		GameVersion      string   `json:"gameVersion,omitempty"`
		LoaderVersion    string   `json:"loaderVersion,omitempty"`
		RconPort         uint16   `json:"rconPort,omitempty"`
		PublicRcon       bool     `json:"publicRcon,omitempty"`
		RestartPolicy    string   `json:"restartPolicy,omitempty"`
		LevelName        string   `json:"levelName,omitempty"`
		LevelSeed        string   `json:"levelSeed,omitempty"`
		LevelType        string   `json:"levelType,omitempty"`
		ExternalJavaArgs []string `json:"externalJavaArgs,omitempty"`
	}
	if err := decodeJSON(r, &req); err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.ServerID == "" {
		id, err := runner.GenerateVolumeID()
		if err != nil {
			jsonError(w, "failed to generate server id", http.StatusInternalServerError)
			return
		}
		req.ServerID = id
	} else if !isValidServerID(req.ServerID) {
		jsonError(w, "invalid server id: must be 1-64 chars of a-z, A-Z, 0-9, _ or -", http.StatusBadRequest)
		return
	}
	if req.RamBytes == 0 {
		req.RamBytes = 2 * 1024 * 1024 * 1024
	}
	if req.GamePort == 0 {
		req.GamePort = 25565
	}
	if req.EngineType == "" {
		req.EngineType = "VANILLA"
	}
	if req.GameVersion == "" {
		req.GameVersion = "LATEST"
	}
	if req.RconPort == 0 {
		req.RconPort = req.GamePort + 10
	}

	portUsed := func(p uint16) bool {
		return h.Instance.IsPortUsed(p, "")
	}
	gamePort, err := runner.FindFreePortExcluding("", req.GamePort, portUsed)
	if err != nil {
		jsonError(w, fmt.Sprintf("no free game port: %v", err), http.StatusInternalServerError)
		return
	}
	rconHost := "127.0.0.1"
	if req.PublicRcon {
		rconHost = ""
	}
	rconPort, err := runner.FindFreePortExcluding(rconHost, req.RconPort, portUsed)
	if err != nil {
		jsonError(w, fmt.Sprintf("no free rcon port: %v", err), http.StatusInternalServerError)
		return
	}
	if rconPort == gamePort {
		rconPort, err = runner.FindFreePortExcluding(rconHost, 0, portUsed)
		if err != nil {
			jsonError(w, fmt.Sprintf("no free rcon port: %v", err), http.StatusInternalServerError)
			return
		}
	}
	req.GamePort = gamePort
	req.RconPort = rconPort

	rconPassword, err := runner.GenerateRconPassword()
	if err != nil {
		jsonError(w, "failed to generate rcon password", http.StatusInternalServerError)
		return
	}

	worldGenEnv := make(map[string]string)
	if req.LevelName != "" {
		worldGenEnv["level-name"] = req.LevelName
	}
	if req.LevelSeed != "" {
		worldGenEnv["level-seed"] = req.LevelSeed
	}
	if req.LevelType != "" {
		worldGenEnv["level-type"] = req.LevelType
	}

	containerID, volumeID, volumePath, err := runner.StartServerContainer(
		r.Context(), h.Cli, req.ServerID,
		req.RamBytes, req.GamePort,
		req.EngineType, req.GameVersion, req.LoaderVersion,
		h.ServersDir, h.ServersHostDir,
		req.RconPort, rconPassword, req.PublicRcon,
		"",
		worldGenEnv,
		req.RestartPolicy,
		h.NetworkName,
		req.ExternalJavaArgs,
	)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s := state.ServerState{
		ServerID:           req.ServerID,
		VolumeID:           volumeID,
		VolumePath:         volumePath,
		HostPath:           runner.HostPathForDocker(volumePath, h.ServersDir, h.ServersHostDir),
		ContainerPath:      "/data",
		ContainerID:        containerID,
		RamBytes:           req.RamBytes,
		GamePort:           req.GamePort,
		EngineType:         req.EngineType,
		GameVersion:        req.GameVersion,
		LoaderVersion:      req.LoaderVersion,
		RconPassword:       rconPassword,
		RconPort:           req.RconPort,
		PublicRcon:         req.PublicRcon,
		RestartPolicy:      req.RestartPolicy,
		ExternalJavaArgs:   req.ExternalJavaArgs,
		ContainerStatus:    "running",
		ContainerStartedAt: time.Now().UTC(),
		ServerStatus:       "starting",
		ModCount:           state.CountMods(state.ServerState{VolumePath: volumePath}),
	}
	h.Instance.Set(s)
	_ = h.Instance.Save()

	slog.Info("server created",
		"server_id", req.ServerID,
		"container_id", containerID[:12],
		"volume_id", volumeID,
		"game_port", req.GamePort,
		"engine", req.EngineType,
		"restart_policy", req.RestartPolicy,
	)
	w.WriteHeader(http.StatusCreated)
	jsonResponse(w, s)
}

// HandleReassignPorts picks new free ports for a stopped server and clears its container ID.
func (h *Handler) HandleReassignPorts(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	s, ok := h.Instance.Get(id)
	if !ok {
		jsonError(w, "not found", http.StatusNotFound)
		return
	}
	if s.ContainerStatus == "running" {
		jsonError(w, "cannot reassign ports while server is running", http.StatusConflict)
		return
	}

	portUsed := func(p uint16) bool {
		return h.Instance.IsPortUsed(p, id)
	}

	gamePort, err := runner.FindFreePortExcluding("", s.GamePort, portUsed)
	if err != nil {
		jsonError(w, fmt.Sprintf("no free game port: %v", err), http.StatusInternalServerError)
		return
	}

	rconHost := "127.0.0.1"
	if s.PublicRcon {
		rconHost = ""
	}
	rconPort, err := runner.FindFreePortExcluding(rconHost, s.RconPort, portUsed)
	if err != nil {
		jsonError(w, fmt.Sprintf("no free rcon port: %v", err), http.StatusInternalServerError)
		return
	}
	if rconPort == gamePort {
		rconPort, err = runner.FindFreePortExcluding(rconHost, 0, portUsed)
		if err != nil {
			jsonError(w, fmt.Sprintf("no free rcon port: %v", err), http.StatusInternalServerError)
			return
		}
	}

	updated := h.Instance.UpdateMeta(id, func(st *state.ServerState) {
		st.GamePort = gamePort
		st.RconPort = rconPort
		st.ContainerID = ""
	})
	if !updated {
		jsonError(w, "not found", http.StatusNotFound)
		return
	}
	_ = h.Instance.Save()
	s, _ = h.Instance.Get(id)
	jsonResponse(w, s)
}

// handleListServers returns all registered server states.
func (h *Handler) HandleListServers(w http.ResponseWriter, r *http.Request) {
	servers := h.Instance.All()
	for i := range servers {
		servers[i] = h.resolveLegacyFields(servers[i])
	}
	jsonResponse(w, servers)
}

// handleGetServer returns a single server state by ID.
func (h *Handler) HandleGetServer(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	s, ok := h.Instance.Get(id)
	if !ok {
		jsonError(w, "not found", http.StatusNotFound)
		return
	}
	jsonResponse(w, h.resolveLegacyFields(s))
}

// handleUpdateServer patches mutable server metadata fields.
func (h *Handler) HandleUpdateServer(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	s, ok := h.Instance.Get(id)
	if !ok {
		jsonError(w, "not found", http.StatusNotFound)
		return
	}

	var req struct {
		RamBytes         int64    `json:"ramBytes,omitempty"`
		GamePort         uint16   `json:"gamePort,omitempty"`
		RconPort         uint16   `json:"rconPort,omitempty"`
		PublicRcon       bool     `json:"publicRcon,omitempty"`
		RestartPolicy    string   `json:"restartPolicy,omitempty"`
		EngineType       string   `json:"engineType,omitempty"`
		ExternalJavaArgs []string `json:"externalJavaArgs,omitempty"`
	}
	if err := decodeJSON(r, &req); err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if s.ContainerStatus == "running" && ((req.GamePort > 0 && req.GamePort != s.GamePort) || (req.RconPort > 0 && req.RconPort != s.RconPort)) {
		jsonError(w, "cannot change port while server is running", http.StatusConflict)
		return
	}
	if s.ContainerStatus == "running" && ((req.RamBytes > 0 && req.RamBytes != s.RamBytes) || (len(req.ExternalJavaArgs) > 0 && !slicesEqual(req.ExternalJavaArgs, s.ExternalJavaArgs))) {
		jsonError(w, "cannot change resources while server is running", http.StatusConflict)
		return
	}

	if req.GamePort > 0 && req.GamePort != s.GamePort {
		if h.Instance.IsPortUsed(req.GamePort, id) || !runner.IsPortFree("", req.GamePort) {
			jsonError(w, "game port unavailable", http.StatusConflict)
			return
		}
	}
	if req.RconPort > 0 && req.RconPort != s.RconPort {
		rconHost := "127.0.0.1"
		if req.PublicRcon {
			rconHost = ""
		}
		if h.Instance.IsPortUsed(req.RconPort, id) || !runner.IsPortFree(rconHost, req.RconPort) {
			jsonError(w, "rcon port unavailable", http.StatusConflict)
			return
		}
	}

	updated := h.Instance.UpdateMeta(id, func(st *state.ServerState) {
		if req.RamBytes > 0 && req.RamBytes != st.RamBytes {
			st.RamBytes = req.RamBytes
			st.ContainerID = ""
		}
		if req.GamePort > 0 && req.GamePort != st.GamePort {
			st.GamePort = req.GamePort
			st.ContainerID = ""
		}
		if req.RconPort > 0 && req.RconPort != st.RconPort {
			st.RconPort = req.RconPort
			st.ContainerID = ""
		}
		if req.PublicRcon != st.PublicRcon {
			st.PublicRcon = req.PublicRcon
			st.ContainerID = ""
		}
		if req.RestartPolicy != "" && req.RestartPolicy != st.RestartPolicy {
			st.RestartPolicy = req.RestartPolicy
			st.ContainerID = ""
		}
		if req.EngineType != "" {
			st.EngineType = req.EngineType
		}
		if len(req.ExternalJavaArgs) > 0 || st.ExternalJavaArgs != nil {
			if !slicesEqual(req.ExternalJavaArgs, st.ExternalJavaArgs) {
				st.ExternalJavaArgs = req.ExternalJavaArgs
				st.ContainerID = ""
			}
		}
	})
	if !updated {
		jsonError(w, "not found", http.StatusNotFound)
		return
	}
	_ = h.Instance.Save()
	s, _ = h.Instance.Get(id)
	jsonResponse(w, h.resolveLegacyFields(s))
}
