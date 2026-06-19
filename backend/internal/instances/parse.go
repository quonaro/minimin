package instances

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
)

// ParseMetadata extracts game version, engine, loader version and instance name
// from the archive based on its format.
func ParseMetadata(r *zip.Reader, format Format) (*Metadata, error) {
	meta := &Metadata{Format: format, DetectedPaths: DetectPaths(r, format)}
	var err error

	switch format {
	case FormatPrism:
		err = parsePrism(r, meta)
	case FormatMrpack:
		err = parseMrpack(r, meta)
	case FormatCurseForge:
		err = parseCurseForge(r, meta)
	}

	if err != nil {
		return nil, fmt.Errorf("parse metadata: %w", err)
	}
	return meta, nil
}

func parsePrism(r *zip.Reader, meta *Metadata) error {
	cfgFile := findFile(r, "instance.cfg")
	if cfgFile != nil {
		data, err := readZipEntry(cfgFile)
		if err != nil {
			return err
		}
		meta.InstanceName = parseIniValue(string(data), "name")
	}

	packFile := findFile(r, "mmc-pack.json")
	if packFile == nil {
		return nil
	}
	data, err := readZipEntry(packFile)
	if err != nil {
		return err
	}

	var pack struct {
		Components []struct {
			UID            string `json:"uid"`
			Version        string `json:"version"`
			CachedName     string `json:"cachedName"`
			CachedRequires []struct {
				UID    string `json:"uid"`
				Equals string `json:"equals"`
			} `json:"cachedRequires"`
		} `json:"components"`
	}
	if err := json.Unmarshal(data, &pack); err != nil {
		return err
	}

	for _, c := range pack.Components {
		uid := strings.ToLower(c.UID)
		switch uid {
		case "net.minecraft":
			meta.GameVersion = firstNonEmpty(c.Version, meta.GameVersion)
		case "net.fabricmc.fabric-loader":
			meta.EngineType = "FABRIC"
			meta.LoaderVersion = firstNonEmpty(c.Version, meta.LoaderVersion)
		case "net.minecraftforge":
			meta.EngineType = "FORGE"
			meta.LoaderVersion = firstNonEmpty(c.Version, meta.LoaderVersion)
		case "net.neoforged":
			meta.EngineType = "FORGE"
			meta.LoaderVersion = firstNonEmpty(c.Version, meta.LoaderVersion)
		}
	}
	return nil
}

func parseMrpack(r *zip.Reader, meta *Metadata) error {
	idxFile := findFile(r, "modrinth.index.json")
	if idxFile == nil {
		return nil
	}
	data, err := readZipEntry(idxFile)
	if err != nil {
		return err
	}

	var idx struct {
		Name          string            `json:"name"`
		Game          string            `json:"game"`
		VersionID     string            `json:"versionId"`
		FormatVersion int               `json:"formatVersion"`
		Dependencies  map[string]string `json:"dependencies"`
	}
	if err := json.Unmarshal(data, &idx); err != nil {
		return err
	}

	meta.InstanceName = firstNonEmpty(idx.Name, meta.InstanceName)
	for dep, ver := range idx.Dependencies {
		switch strings.ToLower(dep) {
		case "minecraft":
			meta.GameVersion = firstNonEmpty(ver, meta.GameVersion)
		case "fabric-loader":
			meta.EngineType = "FABRIC"
			meta.LoaderVersion = firstNonEmpty(ver, meta.LoaderVersion)
		case "forge":
			meta.EngineType = "FORGE"
			meta.LoaderVersion = firstNonEmpty(ver, meta.LoaderVersion)
		case "neoforge":
			meta.EngineType = "FORGE"
			meta.LoaderVersion = firstNonEmpty(ver, meta.LoaderVersion)
		case "quilt-loader":
			meta.EngineType = "FABRIC"
			meta.LoaderVersion = firstNonEmpty(ver, meta.LoaderVersion)
		}
	}
	return nil
}

func parseCurseForge(r *zip.Reader, meta *Metadata) error {
	manifestFile := findFile(r, "manifest.json")
	if manifestFile == nil {
		return nil
	}
	data, err := readZipEntry(manifestFile)
	if err != nil {
		return err
	}

	var manifest struct {
		Name      string `json:"name"`
		Minecraft struct {
			Version    string `json:"version"`
			ModLoaders []struct {
				ID      string `json:"id"`
				Primary bool   `json:"primary"`
			} `json:"modLoaders"`
		} `json:"minecraft"`
	}
	if err := json.Unmarshal(data, &manifest); err != nil {
		return err
	}

	meta.InstanceName = firstNonEmpty(manifest.Name, meta.InstanceName)
	meta.GameVersion = firstNonEmpty(manifest.Minecraft.Version, meta.GameVersion)

	for _, loader := range manifest.Minecraft.ModLoaders {
		parts := strings.SplitN(loader.ID, "-", 2)
		engine := strings.ToLower(parts[0])
		var version string
		if len(parts) > 1 {
			version = parts[1]
		}
		if engine == "" {
			continue
		}
		if alias, ok := gameVersionAliases[engine]; ok {
			meta.EngineType = alias
		} else {
			meta.EngineType = strings.ToUpper(engine)
		}
		meta.LoaderVersion = firstNonEmpty(version, meta.LoaderVersion)
		if loader.Primary {
			break
		}
	}
	return nil
}

func parseIniValue(text, key string) string {
	prefix := key + "="
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, prefix) {
			return strings.TrimSpace(strings.TrimPrefix(line, prefix))
		}
	}
	return ""
}

// safeBaseName returns the base name of a path without extension.
func safeBaseName(name string) string {
	base := filepath.Base(name)
	ext := filepath.Ext(base)
	if ext != "" {
		base = base[:len(base)-len(ext)]
	}
	return base
}
