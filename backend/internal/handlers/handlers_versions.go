package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"
	"time"
)

var (
	versionsCache     []byte
	versionsCacheTime time.Time
	versionsCacheMu   sync.RWMutex
)

// HandleGetVersions aggregates Minecraft version metadata from upstream APIs
// and returns a unified JSON payload. Results are cached for 5 minutes.
func (h *Handler) HandleGetVersions(w http.ResponseWriter, r *http.Request) {
	versionsCacheMu.RLock()
	if versionsCache != nil && time.Since(versionsCacheTime) < 5*time.Minute {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "public, max-age=300")
		w.Write(versionsCache)
		versionsCacheMu.RUnlock()
		return
	}
	versionsCacheMu.RUnlock()

	data, err := h.fetchVersions()
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	versionsCacheMu.Lock()
	versionsCache = data
	versionsCacheTime = time.Now()
	versionsCacheMu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "public, max-age=300")
	w.Write(data)
}

type versionResp struct {
	ID     string `json:"id"`
	Stable bool   `json:"stable"`
}

type loaderResp struct {
	ID     string `json:"id"`
	Stable bool   `json:"stable"`
}

type forgeResp struct {
	ID      string   `json:"id"`
	Stable  bool     `json:"stable"`
	Loaders []string `json:"loaders,omitempty"`
}

type allVersionsResp struct {
	Vanilla       []versionResp `json:"vanilla"`
	Paper         []versionResp `json:"paper"`
	FabricGames   []versionResp `json:"fabricGames"`
	FabricLoaders []loaderResp  `json:"fabricLoaders"`
	Forge         []forgeResp   `json:"forge"`
}

func (h *Handler) fetchVersions() ([]byte, error) {
	result := allVersionsResp{}
	client := &http.Client{Timeout: 10 * time.Second}

	// Mojang
	if resp, err := client.Get("https://launchermeta.mojang.com/mc/game/version_manifest.json"); err == nil {
		var d struct {
			Versions []struct {
				ID   string `json:"id"`
				Type string `json:"type"`
			} `json:"versions"`
		}
		if json.NewDecoder(resp.Body).Decode(&d) == nil {
			for _, v := range d.Versions {
				result.Vanilla = append(result.Vanilla, versionResp{ID: v.ID, Stable: v.Type == "release"})
			}
		}
		resp.Body.Close()
	}

	// Paper
	if resp, err := client.Get("https://api.papermc.io/v2/projects/paper"); err == nil {
		var d struct {
			Versions []string `json:"versions"`
		}
		if json.NewDecoder(resp.Body).Decode(&d) == nil {
			for i := len(d.Versions) - 1; i >= 0; i-- {
				result.Paper = append(result.Paper, versionResp{ID: d.Versions[i], Stable: true})
			}
		}
		resp.Body.Close()
	}

	// Fabric games
	if resp, err := client.Get("https://meta.fabricmc.net/v2/versions/game"); err == nil {
		var d []struct {
			Version string `json:"version"`
			Stable  bool   `json:"stable"`
		}
		if json.NewDecoder(resp.Body).Decode(&d) == nil {
			for _, v := range d {
				result.FabricGames = append(result.FabricGames, versionResp{ID: v.Version, Stable: v.Stable})
			}
		}
		resp.Body.Close()
	}

	// Fabric loaders
	if resp, err := client.Get("https://meta.fabricmc.net/v2/versions/loader"); err == nil {
		var d []struct {
			Version string `json:"version"`
		}
		if json.NewDecoder(resp.Body).Decode(&d) == nil {
			for _, v := range d {
				result.FabricLoaders = append(result.FabricLoaders, loaderResp{ID: v.Version, Stable: true})
			}
		}
		resp.Body.Close()
	}

	// Forge
	if resp, err := client.Get("https://files.minecraftforge.net/net/minecraftforge/forge/promotions_slim.json"); err == nil {
		var d struct {
			Promos map[string]string `json:"promos"`
		}
		if json.NewDecoder(resp.Body).Decode(&d) == nil {
			forgeMap := make(map[string][]string)
			for key, val := range d.Promos {
				parts := strings.Split(key, "-")
				if len(parts) == 2 {
					mc := parts[0]
					forgeMap[mc] = append(forgeMap[mc], val)
				}
			}
			for k, loaders := range forgeMap {
				result.Forge = append(result.Forge, forgeResp{ID: k, Stable: true, Loaders: loaders})
			}
		}
		resp.Body.Close()
	}

	return json.Marshal(result)
}
