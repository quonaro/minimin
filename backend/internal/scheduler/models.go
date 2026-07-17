// Package scheduler provides task scheduling and automation for servers.
package scheduler

import (
	"time"
)

// TriggerType identifies how a task is fired.
type TriggerType string

const (
	TriggerInterval TriggerType = "interval"
	TriggerCron     TriggerType = "cron"
	TriggerEvent    TriggerType = "event"
)

// ActionType identifies what a task does when fired.
type ActionType string

const (
	ActionRCON          ActionType = "rcon"
	ActionContainerExec ActionType = "container_exec"
	ActionLifecycle     ActionType = "lifecycle"
	ActionBackup        ActionType = "backup"
)

// Trigger defines when a task should run.
type Trigger struct {
	Type     TriggerType `json:"type" yaml:"type"`
	Interval string      `json:"interval,omitempty" yaml:"interval,omitempty"` // e.g. "5m", "1h"
	Cron     string      `json:"cron,omitempty" yaml:"cron,omitempty"`           // e.g. "0 4 * * *"
	Event    string      `json:"event,omitempty" yaml:"event,omitempty"`         // e.g. "server_start"
}

// Action defines what a task does.
type Action struct {
	Type         ActionType `json:"type" yaml:"type"`
	Command      string     `json:"command,omitempty" yaml:"command,omitempty"`                // rcon or exec
	Lifecycle    string     `json:"lifecycle,omitempty" yaml:"lifecycle,omitempty"`            // start | stop | restart
	BackupScope  string     `json:"backupScope,omitempty" yaml:"backup_scope,omitempty"`       // world only for now
	KeepLast     int        `json:"keepLast,omitempty" yaml:"keep_last,omitempty"`             // retention count
	KeepDays     int        `json:"keepDays,omitempty" yaml:"keep_days,omitempty"`             // retention age
}

// Task is a single scheduled or event-driven automation rule.
type Task struct {
	ID        string    `json:"id" yaml:"id"`
	ServerID  string    `json:"serverId" yaml:"server_id"`
	Name      string    `json:"name" yaml:"name"`
	Enabled   bool      `json:"enabled" yaml:"enabled"`
	Trigger   Trigger   `json:"trigger" yaml:"trigger"`
	Action    Action    `json:"action" yaml:"action"`
	LastRun   time.Time `json:"lastRun,omitempty" yaml:"last_run,omitempty"`
	NextRun   time.Time `json:"nextRun,omitempty" yaml:"next_run,omitempty"`
}

// ActionsFile is the root YAML structure.
type ActionsFile struct {
	Tasks []Task `json:"tasks" yaml:"tasks"`
}
