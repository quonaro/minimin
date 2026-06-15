package archive

import "time"

// Token represents a generated archive with expiration.
type Token struct {
	Token          string         `json:"token"`
	ServerID       string         `json:"serverId"`
	ServerName     string         `json:"serverName"`
	Include        []string       `json:"include"`
	ExpiresAt      time.Time      `json:"expiresAt"`
	CreatedAt      time.Time      `json:"createdAt"`
	DownloadCounts map[string]int `json:"downloadCounts"`
	TotalDownloads int            `json:"totalDownloads"`
}

type metaEntry struct {
	Token          string         `json:"token"`
	ServerID       string         `json:"serverId"`
	ServerName     string         `json:"serverName"`
	Include        []string       `json:"include"`
	ExpiresAt      string         `json:"expiresAt"`
	CreatedAt      string         `json:"createdAt"`
	DownloadCounts map[string]int `json:"downloadCounts"`
	TotalDownloads int            `json:"totalDownloads"`
}

func (a *Token) toMetaEntry() metaEntry {
	return metaEntry{
		Token:          a.Token,
		ServerID:       a.ServerID,
		ServerName:     a.ServerName,
		Include:        a.Include,
		ExpiresAt:      a.ExpiresAt.Format(time.RFC3339),
		CreatedAt:      a.CreatedAt.Format(time.RFC3339),
		DownloadCounts: a.DownloadCounts,
		TotalDownloads: a.TotalDownloads,
	}
}

func entryToToken(e metaEntry) *Token {
	expiresAt, _ := time.Parse(time.RFC3339, e.ExpiresAt)
	createdAt, _ := time.Parse(time.RFC3339, e.CreatedAt)
	return &Token{
		Token:          e.Token,
		ServerID:       e.ServerID,
		ServerName:     e.ServerName,
		Include:        e.Include,
		ExpiresAt:      expiresAt,
		CreatedAt:      createdAt,
		DownloadCounts: e.DownloadCounts,
		TotalDownloads: e.TotalDownloads,
	}
}

// CreateRequest is the payload for creating a new archive.
type CreateRequest struct {
	TTL     int      `json:"ttl"`     // hours
	Include []string `json:"include"` // "mods", "resourcepacks", "shaderpacks"
}

// Summary is a lightweight view of an archive token.
type Summary struct {
	Token          string         `json:"token"`
	ServerName     string         `json:"serverName"`
	ExpiresAt      time.Time      `json:"expiresAt"`
	CreatedAt      time.Time      `json:"createdAt"`
	Formats        []string       `json:"formats"`
	DownloadCounts map[string]int `json:"downloadCounts"`
	TotalDownloads int            `json:"totalDownloads"`
}
