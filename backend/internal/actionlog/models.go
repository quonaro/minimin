package actionlog

import "time"

// Entry records a single executed action.
type Entry struct {
	ID        string    `json:"id" yaml:"id"`
	Timestamp time.Time `json:"timestamp" yaml:"timestamp"`
	ServerID  string    `json:"serverId" yaml:"server_id"`
	TaskID    string    `json:"taskId,omitempty" yaml:"task_id,omitempty"`
	Source    string    `json:"source" yaml:"source"` // "manual", "interval", "cron", "event"
	Action    string    `json:"action" yaml:"action"` // "rcon", "container_exec", "lifecycle", "backup", "restore", "delete_backup"
	Detail    string    `json:"detail,omitempty" yaml:"detail,omitempty"`
	Status    string    `json:"status" yaml:"status"` // "success", "error", "skipped"
	Message   string    `json:"message,omitempty" yaml:"message,omitempty"`
}

// LogFile is the top-level YAML structure.
type LogFile struct {
	Entries []Entry `yaml:"entries"`
}
