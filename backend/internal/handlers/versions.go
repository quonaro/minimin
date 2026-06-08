package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/danielgtaylor/huma/v2"
)

// VanillaVersion represents a Mojang version entry.
type VanillaVersion struct {
	ID          string `json:"id"`
	ReleaseDate string `json:"releaseDate,omitempty"`
	Stable      bool   `json:"stable,omitempty"`
}

// VanillaVersionsOutput is the output for GET /versions/vanilla.
type VanillaVersionsOutput struct {
	Body []VanillaVersion
}

// GetVanillaVersions fetches the Minecraft version manifest from Mojang.
func (h *Handler) GetVanillaVersions(ctx context.Context, input *struct{}) (*VanillaVersionsOutput, error) {
	client := &http.Client{Timeout: 30 * time.Second}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://launchermeta.mojang.com/mc/game/version_manifest.json", nil)
	if err != nil {
		return nil, huma.Error500InternalServerError("failed to create request", err)
	}
	resp, err := client.Do(req)
	if err != nil {
		slog.Error("failed to fetch vanilla versions", "error", err)
		return nil, huma.Error500InternalServerError("failed to fetch vanilla versions", err)
	}
	defer func() { _ = resp.Body.Close() }()

	var body struct {
		Versions []struct {
			ID          string `json:"id"`
			ReleaseTime string `json:"releaseTime"`
			Type        string `json:"type"`
		} `json:"versions"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, huma.Error500InternalServerError("failed to decode vanilla versions", err)
	}

	versions := make([]VanillaVersion, 0, len(body.Versions))
	for i := len(body.Versions) - 1; i >= 0; i-- {
		v := body.Versions[i]
		versions = append(versions, VanillaVersion{
			ID:          v.ID,
			ReleaseDate: v.ReleaseTime,
			Stable:      v.Type == "release",
		})
	}
	return &VanillaVersionsOutput{Body: versions}, nil
}

// PaperVersion represents a Paper version entry.
type PaperVersion struct {
	ID     string `json:"id"`
	Stable bool   `json:"stable,omitempty"`
}

// PaperVersionsOutput is the output for GET /versions/paper.
type PaperVersionsOutput struct {
	Body []PaperVersion
}

// GetPaperVersions fetches available Paper versions.
func (h *Handler) GetPaperVersions(ctx context.Context, input *struct{}) (*PaperVersionsOutput, error) {
	client := &http.Client{Timeout: 30 * time.Second}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.papermc.io/v2/projects/paper", nil)
	if err != nil {
		return nil, huma.Error500InternalServerError("failed to create request", err)
	}
	resp, err := client.Do(req)
	if err != nil {
		slog.Error("failed to fetch paper versions", "error", err)
		return nil, huma.Error500InternalServerError("failed to fetch paper versions", err)
	}
	defer func() { _ = resp.Body.Close() }()

	var body struct {
		Versions []string `json:"versions"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, huma.Error500InternalServerError("failed to decode paper versions", err)
	}

	versions := make([]PaperVersion, 0, len(body.Versions))
	for i := len(body.Versions) - 1; i >= 0; i-- {
		versions = append(versions, PaperVersion{ID: body.Versions[i], Stable: true})
	}
	return &PaperVersionsOutput{Body: versions}, nil
}

// FabricGameVersion represents a Fabric game version.
type FabricGameVersion struct {
	ID     string `json:"id"`
	Stable bool   `json:"stable,omitempty"`
}

// FabricLoaderVersion represents a Fabric loader version.
type FabricLoaderVersion struct {
	ID     string `json:"id"`
	Stable bool   `json:"stable,omitempty"`
}

// FabricVersionsOutput is the output for GET /versions/fabric.
type FabricVersionsOutput struct {
	Body struct {
		GameVersions   []FabricGameVersion   `json:"gameVersions"`
		LoaderVersions []FabricLoaderVersion `json:"loaderVersions"`
	}
}

// GetFabricVersions fetches Fabric game and loader versions in parallel.
func (h *Handler) GetFabricVersions(ctx context.Context, input *struct{}) (*FabricVersionsOutput, error) {
	client := &http.Client{Timeout: 30 * time.Second}

	type gameResult struct {
		versions []struct {
			Version string `json:"version"`
			Stable  bool   `json:"stable"`
		}
		err error
	}
	type loaderResult struct {
		versions []struct {
			Version string `json:"version"`
			Stable  bool   `json:"stable"`
		}
		err error
	}

	gameCh := make(chan gameResult, 1)
	loaderCh := make(chan loaderResult, 1)

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://meta.fabricmc.net/v2/versions/game", nil)
		if err != nil {
			gameCh <- gameResult{err: err}
			return
		}
		resp, err := client.Do(req)
		if err != nil {
			gameCh <- gameResult{err: err}
			return
		}
		defer func() { _ = resp.Body.Close() }()
		var versions []struct {
			Version string `json:"version"`
			Stable  bool   `json:"stable"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&versions); err != nil {
			gameCh <- gameResult{err: err}
			return
		}
		gameCh <- gameResult{versions: versions}
	}()

	go func() {
		defer wg.Done()
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://meta.fabricmc.net/v2/versions/loader", nil)
		if err != nil {
			loaderCh <- loaderResult{err: err}
			return
		}
		resp, err := client.Do(req)
		if err != nil {
			loaderCh <- loaderResult{err: err}
			return
		}
		defer func() { _ = resp.Body.Close() }()
		var versions []struct {
			Version string `json:"version"`
			Stable  bool   `json:"stable"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&versions); err != nil {
			loaderCh <- loaderResult{err: err}
			return
		}
		loaderCh <- loaderResult{versions: versions}
	}()

	wg.Wait()
	close(gameCh)
	close(loaderCh)

	gameRes := <-gameCh
	loaderRes := <-loaderCh

	if gameRes.err != nil {
		return nil, huma.Error500InternalServerError("failed to fetch fabric game versions", gameRes.err)
	}
	if loaderRes.err != nil {
		return nil, huma.Error500InternalServerError("failed to fetch fabric loader versions", loaderRes.err)
	}

	out := &FabricVersionsOutput{}
	out.Body.GameVersions = make([]FabricGameVersion, 0, len(gameRes.versions))
	for _, v := range gameRes.versions {
		out.Body.GameVersions = append(out.Body.GameVersions, FabricGameVersion{ID: v.Version, Stable: v.Stable})
	}
	out.Body.LoaderVersions = make([]FabricLoaderVersion, 0, len(loaderRes.versions))
	for _, v := range loaderRes.versions {
		out.Body.LoaderVersions = append(out.Body.LoaderVersions, FabricLoaderVersion{ID: v.Version, Stable: v.Stable})
	}
	return out, nil
}

// ForgeVersionInfo holds parsed Forge data for a MC version.
type ForgeVersionInfo struct {
	ID      string   `json:"id"`
	Stable  bool     `json:"stable,omitempty"`
	Loaders []string `json:"loaders,omitempty"`
}

// ForgeVersionsOutput is the output for GET /versions/forge.
type ForgeVersionsOutput struct {
	Body []ForgeVersionInfo
}

// CachedVersions holds all version data with a timestamp.
type CachedVersions struct {
	Vanilla       []VanillaVersion
	Paper         []PaperVersion
	FabricGames   []FabricGameVersion
	FabricLoaders []FabricLoaderVersion
	Forge         []ForgeVersionInfo
	CachedAt      time.Time
	Version       int
}

const cacheVersion = 2

var (
	versionsCache    *CachedVersions
	versionsCacheMu  sync.RWMutex
	versionsCacheTTL = 6 * time.Hour
)

// semverGreater reports whether a is semantically greater than b (descending order).
// Both strings are expected to be dot-separated numeric segments.
func semverGreater(a, b string) bool {
	partsA := strings.Split(a, ".")
	partsB := strings.Split(b, ".")
	for i := 0; i < len(partsA) && i < len(partsB); i++ {
		na, errA := strconv.Atoi(partsA[i])
		nb, errB := strconv.Atoi(partsB[i])
		if errA != nil || errB != nil {
			// Fallback to string comparison if non-numeric
			if partsA[i] != partsB[i] {
				return partsA[i] > partsB[i]
			}
			continue
		}
		if na != nb {
			return na > nb
		}
	}
	return len(partsA) > len(partsB)
}

// AllVersionsOutput is the output for GET /versions/all.
type AllVersionsOutput struct {
	Body struct {
		Vanilla       []VanillaVersion      `json:"vanilla"`
		Paper         []PaperVersion        `json:"paper"`
		FabricGames   []FabricGameVersion   `json:"fabricGames"`
		FabricLoaders []FabricLoaderVersion `json:"fabricLoaders"`
		Forge         []ForgeVersionInfo    `json:"forge"`
	}
}

// GetAllVersions fetches all Minecraft versions in parallel with caching.
func (h *Handler) GetAllVersions(ctx context.Context, input *struct{}) (*AllVersionsOutput, error) {
	versionsCacheMu.RLock()
	cached := versionsCache
	versionsCacheMu.RUnlock()

	if cached != nil && cached.Version == cacheVersion && time.Since(cached.CachedAt) < versionsCacheTTL {
		out := &AllVersionsOutput{}
		out.Body.Vanilla = cached.Vanilla
		out.Body.Paper = cached.Paper
		out.Body.FabricGames = cached.FabricGames
		out.Body.FabricLoaders = cached.FabricLoaders
		out.Body.Forge = cached.Forge
		return out, nil
	}

	client := &http.Client{Timeout: 30 * time.Second}

	type vanillaRes struct {
		versions []VanillaVersion
		err      error
	}
	type paperRes struct {
		versions []PaperVersion
		err      error
	}
	type fabricGameRes struct {
		versions []FabricGameVersion
		err      error
	}
	type fabricLoaderRes struct {
		versions []FabricLoaderVersion
		err      error
	}
	type forgeRes struct {
		versions []ForgeVersionInfo
		err      error
	}

	vanillaCh := make(chan vanillaRes, 1)
	paperCh := make(chan paperRes, 1)
	fabricGameCh := make(chan fabricGameRes, 1)
	fabricLoaderCh := make(chan fabricLoaderRes, 1)
	forgeCh := make(chan forgeRes, 1)

	var wg sync.WaitGroup
	wg.Add(5)

	go func() {
		defer wg.Done()
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://launchermeta.mojang.com/mc/game/version_manifest.json", nil)
		if err != nil {
			vanillaCh <- vanillaRes{err: err}
			return
		}
		resp, err := client.Do(req)
		if err != nil {
			vanillaCh <- vanillaRes{err: err}
			return
		}
		defer func() { _ = resp.Body.Close() }()
		var body struct {
			Versions []struct {
				ID          string `json:"id"`
				ReleaseTime string `json:"releaseTime"`
				Type        string `json:"type"`
			} `json:"versions"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
			vanillaCh <- vanillaRes{err: err}
			return
		}
		versions := make([]VanillaVersion, 0, len(body.Versions))
		for _, v := range body.Versions {
			versions = append(versions, VanillaVersion{
				ID:          v.ID,
				ReleaseDate: v.ReleaseTime,
				Stable:      v.Type == "release",
			})
		}
		vanillaCh <- vanillaRes{versions: versions}
	}()

	go func() {
		defer wg.Done()
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.papermc.io/v2/projects/paper", nil)
		if err != nil {
			paperCh <- paperRes{err: err}
			return
		}
		resp, err := client.Do(req)
		if err != nil {
			paperCh <- paperRes{err: err}
			return
		}
		defer func() { _ = resp.Body.Close() }()
		var body struct {
			Versions []string `json:"versions"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
			paperCh <- paperRes{err: err}
			return
		}
		versions := make([]PaperVersion, 0, len(body.Versions))
		for i := len(body.Versions) - 1; i >= 0; i-- {
			versions = append(versions, PaperVersion{ID: body.Versions[i], Stable: true})
		}
		paperCh <- paperRes{versions: versions}
	}()

	go func() {
		defer wg.Done()
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://meta.fabricmc.net/v2/versions/game", nil)
		if err != nil {
			fabricGameCh <- fabricGameRes{err: err}
			return
		}
		resp, err := client.Do(req)
		if err != nil {
			fabricGameCh <- fabricGameRes{err: err}
			return
		}
		defer func() { _ = resp.Body.Close() }()
		var versions []struct {
			Version string `json:"version"`
			Stable  bool   `json:"stable"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&versions); err != nil {
			fabricGameCh <- fabricGameRes{err: err}
			return
		}
		result := make([]FabricGameVersion, 0, len(versions))
		for _, v := range versions {
			result = append(result, FabricGameVersion{ID: v.Version, Stable: v.Stable})
		}
		fabricGameCh <- fabricGameRes{versions: result}
	}()

	go func() {
		defer wg.Done()
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://meta.fabricmc.net/v2/versions/loader", nil)
		if err != nil {
			fabricLoaderCh <- fabricLoaderRes{err: err}
			return
		}
		resp, err := client.Do(req)
		if err != nil {
			fabricLoaderCh <- fabricLoaderRes{err: err}
			return
		}
		defer func() { _ = resp.Body.Close() }()
		var versions []struct {
			Version string `json:"version"`
			Stable  bool   `json:"stable"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&versions); err != nil {
			fabricLoaderCh <- fabricLoaderRes{err: err}
			return
		}
		result := make([]FabricLoaderVersion, 0, len(versions))
		for _, v := range versions {
			result = append(result, FabricLoaderVersion{ID: v.Version, Stable: v.Stable})
		}
		fabricLoaderCh <- fabricLoaderRes{versions: result}
	}()

	go func() {
		defer wg.Done()
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://files.minecraftforge.net/net/minecraftforge/forge/promotions_slim.json", nil)
		if err != nil {
			forgeCh <- forgeRes{err: err}
			return
		}
		resp, err := client.Do(req)
		if err != nil {
			forgeCh <- forgeRes{err: err}
			return
		}
		defer func() { _ = resp.Body.Close() }()
		if resp.StatusCode != http.StatusOK {
			forgeCh <- forgeRes{err: fmt.Errorf("forge API returned status %d", resp.StatusCode)}
			return
		}
		ct := resp.Header.Get("Content-Type")
		if ct == "" || !strings.Contains(ct, "application/json") {
			forgeCh <- forgeRes{err: fmt.Errorf("forge API returned non-JSON content type: %s", ct)}
			return
		}
		var body struct {
			Promos map[string]string `json:"promos"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
			forgeCh <- forgeRes{err: err}
			return
		}
		mcVersions := make(map[string]*ForgeVersionInfo)
		for key, forgeVer := range body.Promos {
			isRecommended := len(key) > 12 && key[len(key)-12:] == "-recommended"
			isLatest := len(key) > 7 && key[len(key)-7:] == "-latest"
			if !isRecommended && !isLatest {
				continue
			}
			var mcVer string
			if isRecommended {
				mcVer = key[:len(key)-12]
			} else {
				mcVer = key[:len(key)-7]
			}
			entry, ok := mcVersions[mcVer]
			if !ok {
				entry = &ForgeVersionInfo{ID: mcVer, Stable: true}
				mcVersions[mcVer] = entry
			}
			found := false
			for _, existing := range entry.Loaders {
				if existing == forgeVer {
					found = true
					break
				}
			}
			if !found {
				entry.Loaders = append(entry.Loaders, forgeVer)
			}
		}
		versions := make([]ForgeVersionInfo, 0, len(mcVersions))
		for _, v := range mcVersions {
			versions = append(versions, *v)
		}
		sort.Slice(versions, func(i, j int) bool {
			return semverGreater(versions[i].ID, versions[j].ID)
		})
		forgeCh <- forgeRes{versions: versions}
	}()

	wg.Wait()
	close(vanillaCh)
	close(paperCh)
	close(fabricGameCh)
	close(fabricLoaderCh)
	close(forgeCh)

	vRes := <-vanillaCh
	pRes := <-paperCh
	fgRes := <-fabricGameCh
	flRes := <-fabricLoaderCh
	fRes := <-forgeCh

	if vRes.err != nil {
		return nil, huma.Error500InternalServerError("failed to fetch vanilla versions", vRes.err)
	}
	if pRes.err != nil {
		return nil, huma.Error500InternalServerError("failed to fetch paper versions", pRes.err)
	}
	if fgRes.err != nil {
		return nil, huma.Error500InternalServerError("failed to fetch fabric game versions", fgRes.err)
	}
	if flRes.err != nil {
		return nil, huma.Error500InternalServerError("failed to fetch fabric loader versions", flRes.err)
	}
	if fRes.err != nil {
		return nil, huma.Error503ServiceUnavailable("failed to fetch forge versions", fRes.err)
	}

	cached = &CachedVersions{
		Vanilla:       vRes.versions,
		Paper:         pRes.versions,
		FabricGames:   fgRes.versions,
		FabricLoaders: flRes.versions,
		Forge:         fRes.versions,
		CachedAt:      time.Now(),
		Version:       cacheVersion,
	}
	versionsCacheMu.Lock()
	versionsCache = cached
	versionsCacheMu.Unlock()

	out := &AllVersionsOutput{}
	out.Body.Vanilla = cached.Vanilla
	out.Body.Paper = cached.Paper
	out.Body.FabricGames = cached.FabricGames
	out.Body.FabricLoaders = cached.FabricLoaders
	out.Body.Forge = cached.Forge
	return out, nil
}

// GetForgeVersions fetches Forge promotions and maps them to MC versions.
func (h *Handler) GetForgeVersions(ctx context.Context, input *struct{}) (*ForgeVersionsOutput, error) {
	client := &http.Client{Timeout: 60 * time.Second}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://files.minecraftforge.net/net/minecraftforge/forge/promotions_slim.json", nil)
	if err != nil {
		return nil, huma.Error500InternalServerError("failed to create request", err)
	}
	resp, err := client.Do(req)
	if err != nil {
		slog.Error("failed to fetch forge versions", "error", err)
		return nil, huma.Error500InternalServerError("failed to fetch forge versions", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		slog.Error("forge API returned non-200 status", "status", resp.StatusCode)
		return nil, huma.Error503ServiceUnavailable("forge API unavailable", nil)
	}

	ct := resp.Header.Get("Content-Type")
	if ct == "" || !strings.Contains(ct, "application/json") {
		slog.Error("forge API returned non-JSON content", "content-type", ct)
		return nil, huma.Error503ServiceUnavailable("forge API returned invalid content type", nil)
	}

	var body struct {
		Promos map[string]string `json:"promos"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		slog.Error("failed to decode forge versions JSON", "error", err)
		return nil, huma.Error503ServiceUnavailable("failed to decode forge versions", err)
	}

	mcVersions := make(map[string]*ForgeVersionInfo)
	for key, forgeVer := range body.Promos {
		isRecommended := len(key) > 12 && key[len(key)-12:] == "-recommended"
		isLatest := len(key) > 7 && key[len(key)-7:] == "-latest"
		if !isRecommended && !isLatest {
			continue
		}
		var mcVer string
		if isRecommended {
			mcVer = key[:len(key)-12]
		} else {
			mcVer = key[:len(key)-7]
		}
		entry, ok := mcVersions[mcVer]
		if !ok {
			entry = &ForgeVersionInfo{ID: mcVer, Stable: true}
			mcVersions[mcVer] = entry
		}
		found := false
		for _, existing := range entry.Loaders {
			if existing == forgeVer {
				found = true
				break
			}
		}
		if !found {
			entry.Loaders = append(entry.Loaders, forgeVer)
		}
	}

	versions := make([]ForgeVersionInfo, 0, len(mcVersions))
	for _, v := range mcVersions {
		versions = append(versions, *v)
	}
	sort.Slice(versions, func(i, j int) bool {
		return semverGreater(versions[i].ID, versions[j].ID)
	})
	return &ForgeVersionsOutput{Body: versions}, nil
}
