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

	CREATE TABLE IF NOT EXISTS servers_cache (
		id TEXT PRIMARY KEY,
		agent_id TEXT NOT NULL,
		server_id TEXT NOT NULL,
		status TEXT NOT NULL,
		game_port INTEGER NOT NULL,
		engine TEXT NOT NULL,
		version TEXT NOT NULL,
		updated_at TEXT NOT NULL,
		FOREIGN KEY (agent_id) REFERENCES agents(id) ON DELETE CASCADE
	);

	CREATE INDEX IF NOT EXISTS idx_servers_cache_agent_id ON servers_cache(agent_id);
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
	APIKey    string    `json:"api_key"`
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
	query := `SELECT id, name, host, api_key, created_at FROM agents WHERE id = ?`
	err := db.conn.QueryRow(query, id).Scan(&agent.ID, &agent.Name, &agent.Host, &agent.APIKey, &agent.CreatedAt)
	if err != nil {
		return Agent{}, false
	}
	agent.CreatedAt, _ = time.Parse(time.RFC3339, agent.CreatedAt.Format(time.RFC3339))
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

// CachedServer represents a cached server state from an agent.
type CachedServer struct {
	ID        string    `json:"id"`
	AgentID   string    `json:"agent_id"`
	ServerID  string    `json:"server_id"`
	Status    string    `json:"status"`
	GamePort  int       `json:"game_port"`
	Engine    string    `json:"engine"`
	Version   string    `json:"version"`
	UpdatedAt time.Time `json:"updated_at"`
}

// UpsertServerCache inserts or updates a cached server.
func (db *DB) UpsertServerCache(server CachedServer) error {
	query := `
	INSERT INTO servers_cache (id, agent_id, server_id, status, game_port, engine, version, updated_at)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	ON CONFLICT(id) DO UPDATE SET
		status = excluded.status,
		game_port = excluded.game_port,
		engine = excluded.engine,
		version = excluded.version,
		updated_at = excluded.updated_at
	`
	_, err := db.conn.Exec(query, server.ID, server.AgentID, server.ServerID, server.Status, server.GamePort, server.Engine, server.Version, server.UpdatedAt.UTC().Format(time.RFC3339))
	if err != nil {
		return fmt.Errorf("failed to upsert server cache: %w", err)
	}
	return nil
}

// ListCachedServers returns all cached servers.
func (db *DB) ListCachedServers() ([]CachedServer, error) {
	query := `SELECT id, agent_id, server_id, status, game_port, engine, version, updated_at FROM servers_cache ORDER BY updated_at DESC`
	rows, err := db.conn.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to list cached servers: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var servers []CachedServer
	for rows.Next() {
		var server CachedServer
		var updatedAtStr string
		if err := rows.Scan(&server.ID, &server.AgentID, &server.ServerID, &server.Status, &server.GamePort, &server.Engine, &server.Version, &updatedAtStr); err != nil {
			return nil, fmt.Errorf("failed to scan cached server: %w", err)
		}
		server.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAtStr)
		servers = append(servers, server)
	}
	return servers, nil
}

// DeleteAgentCache removes all cached servers for an agent.
func (db *DB) DeleteAgentCache(agentID string) error {
	query := `DELETE FROM servers_cache WHERE agent_id = ?`
	_, err := db.conn.Exec(query, agentID)
	if err != nil {
		return fmt.Errorf("failed to delete agent cache: %w", err)
	}
	return nil
}
