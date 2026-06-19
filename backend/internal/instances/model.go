// Package instances parses and extracts Minecraft modpack / instance archives
// for importing into a server volume.
package instances

import "time"

// Format describes the detected kind of archive.
type Format string

const (
	FormatPrism      Format = "prism"
	FormatMrpack     Format = "mrpack"
	FormatCurseForge Format = "curseforge"
	FormatPlain      Format = "plain"
)

// Metadata contains values that can be inferred from the archive.
type Metadata struct {
	Format        Format   `json:"format"`
	InstanceName  string   `json:"instanceName,omitempty"`
	GameVersion   string   `json:"gameVersion,omitempty"`
	EngineType    string   `json:"engineType,omitempty"`
	LoaderVersion string   `json:"loaderVersion,omitempty"`
	DetectedPaths []string `json:"detectedPaths"`
}

// TempEntry holds an uploaded archive awaiting extraction.
type TempEntry struct {
	Token     string
	Path      string
	CreatedAt time.Time
	Size      int64
}

// ExtractOptions controls how an archive is unpacked into a server volume.
type ExtractOptions struct {
	// TargetDir is the server volume directory.
	TargetDir string

	// AllowedDirs restricts extraction to these top-level directory names.
	AllowedDirs []string

	// StripComponents removes leading path components when matching AllowedDirs.
	// 0 means the directory must appear at the archive root.
	StripComponents int
}

// allowedTopLevelDirs is the default list of directories that may be written
// to a server volume when importing an archive.
var allowedTopLevelDirs = []string{
	"mods",
	"resourcepacks",
	"shaderpacks",
	"kubejs",
	"defaultconfigs",
	"scripts",
	"libraries",
}

// gameVersionAliases maps detected engine names to the engine types used by
// the orchestrator. Empty value means "keep the raw value".
var gameVersionAliases = map[string]string{
	"fabric":   "FABRIC",
	"forge":    "FORGE",
	"neoforge": "FORGE",
	"quilt":    "FABRIC",
}
