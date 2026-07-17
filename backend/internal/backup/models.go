// Package backup provides server world backup creation, retention and restoration.
package backup

import "time"

// Backup represents a single world snapshot.
type Backup struct {
	Name      string    `json:"name"`
	ServerID  string    `json:"serverId"`
	SizeBytes int64     `json:"sizeBytes"`
	CreatedAt time.Time `json:"createdAt"`
}
