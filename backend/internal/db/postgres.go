// Package db provides PostgreSQL connectivity and schema management
// for the Meta Clash backend. Runtime game state stays in-memory;
// this layer persists only user accounts and match history.
package db

import (
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	_ "github.com/lib/pq" // PostgreSQL driver registration
)

// Connect opens a PostgreSQL connection pool and verifies connectivity.
// Pool limits are tuned for a small-to-medium game server workload.
func Connect(dsn string) (*sql.DB, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Pool tuning
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Verify the connection is alive
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	slog.Info("database connected", "dsn", redactDSN(dsn))
	return db, nil
}

// Migrate creates the required tables if they do not already exist.
// Uses IF NOT EXISTS so it is safe to call on every startup.
func Migrate(db *sql.DB) error {
	migrations := []struct {
		name string
		stmt string
	}{
		{
			name: "create_users",
			stmt: `
				CREATE TABLE IF NOT EXISTS users (
					id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
					username      VARCHAR(50) UNIQUE NOT NULL,
					password_hash TEXT NOT NULL,
					created_at    TIMESTAMPTZ DEFAULT NOW()
				);`,
		},
		{
			name: "create_matches",
			stmt: `
				CREATE TABLE IF NOT EXISTS matches (
					id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
					lobby_code  VARCHAR(10) NOT NULL,
					theme       VARCHAR(100) NOT NULL,
					winner_id   UUID REFERENCES users(id),
					started_at  TIMESTAMPTZ NOT NULL,
					finished_at TIMESTAMPTZ NOT NULL
				);`,
		},
		{
			name: "create_match_players",
			stmt: `
				CREATE TABLE IF NOT EXISTS match_players (
					id       UUID PRIMARY KEY DEFAULT gen_random_uuid(),
					match_id UUID NOT NULL REFERENCES matches(id) ON DELETE CASCADE,
					user_id  UUID,
					is_bot   BOOLEAN NOT NULL DEFAULT FALSE,
					score    INTEGER NOT NULL DEFAULT 0
				);`,
		},
		{
			name: "index_match_players_match_id",
			stmt: `CREATE INDEX IF NOT EXISTS idx_match_players_match_id ON match_players(match_id);`,
		},
		{
			name: "index_match_players_user_id",
			stmt: `CREATE INDEX IF NOT EXISTS idx_match_players_user_id ON match_players(user_id);`,
		},
	}

	for _, m := range migrations {
		if _, err := db.Exec(m.stmt); err != nil {
			return fmt.Errorf("migration %q failed: %w", m.name, err)
		}
		slog.Info("migration applied", "name", m.name)
	}

	return nil
}

// redactDSN hides the password portion of a DSN for safe logging.
func redactDSN(dsn string) string {
	if len(dsn) > 30 {
		return dsn[:30] + "..."
	}
	return dsn
}
