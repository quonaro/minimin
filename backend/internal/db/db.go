package db

import (
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	_ "modernc.org/sqlite"
)

// DB wraps SQLite database connection.
type DB struct {
	conn *sql.DB
}

// Open opens SQLite database at given path and initializes schema.
func Open(path string) (*DB, error) {
	conn, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	db := &DB{conn: conn}
	if err := db.initSchema(); err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("failed to init schema: %w", err)
	}

	return db, nil
}

// Close closes the database connection.
func (db *DB) Close() error {
	return db.conn.Close()
}

// initSchema creates tables if they don't exist.
func (db *DB) initSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS agents (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		host TEXT NOT NULL,
		api_key TEXT NOT NULL,
		created_at TEXT NOT NULL
	);

	`

	if _, err := db.conn.Exec(schema); err != nil {
		return fmt.Errorf("failed to create schema: %w", err)
	}

	return nil
}

// Agent represents a registered Minecraft agent.
type Agent struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Host      string    `json:"host"`
	APIKey    string    `json:"api_key,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// CreateAgent inserts a new agent into the database.
func (db *DB) CreateAgent(agent Agent) error {
	query := `
	INSERT INTO agents (id, name, host, api_key, created_at)
	VALUES (?, ?, ?, ?, ?)
	`
	_, err := db.conn.Exec(query, agent.ID, agent.Name, agent.Host, agent.APIKey, agent.CreatedAt.UTC().Format(time.RFC3339))
	if err != nil {
		return fmt.Errorf("failed to create agent: %w", err)
	}
	slog.Info("agent created", "id", agent.ID, "name", agent.Name, "host", agent.Host)
	return nil
}

// GetAgent retrieves an agent by ID.
func (db *DB) GetAgent(id string) (Agent, bool) {
	var agent Agent
	var createdAtStr string
	query := `SELECT id, name, host, api_key, created_at FROM agents WHERE id = ?`
	err := db.conn.QueryRow(query, id).Scan(&agent.ID, &agent.Name, &agent.Host, &agent.APIKey, &createdAtStr)
	if err != nil {
		slog.Warn("GetAgent failed", "id", id, "error", err)
		return Agent{}, false
	}
	agent.CreatedAt, _ = time.Parse(time.RFC3339, createdAtStr)
	return agent, true
}

// ListAgents returns all agents.
func (db *DB) ListAgents() ([]Agent, error) {
	query := `SELECT id, name, host, api_key, created_at FROM agents ORDER BY created_at DESC`
	rows, err := db.conn.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to list agents: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var agents []Agent
	for rows.Next() {
		var agent Agent
		var createdAtStr string
		if err := rows.Scan(&agent.ID, &agent.Name, &agent.Host, &agent.APIKey, &createdAtStr); err != nil {
			return nil, fmt.Errorf("failed to scan agent: %w", err)
		}
		agent.CreatedAt, _ = time.Parse(time.RFC3339, createdAtStr)
		agents = append(agents, agent)
	}
	return agents, nil
}

// DeleteAgent removes an agent by ID.
func (db *DB) DeleteAgent(id string) error {
	query := `DELETE FROM agents WHERE id = ?`
	result, err := db.conn.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete agent: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("agent not found")
	}
	slog.Info("agent deleted", "id", id)
	return nil
}
