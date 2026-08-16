package db

import (
	"database/sql"
	"fmt"

	"github.com/nikhilsaxena04/meta_clash/backend/internal/models"
)

// PostgresRepo implements models.UserRepository using raw database/sql queries.
type PostgresRepo struct {
	db *sql.DB
}

// NewPostgresRepo creates a new repository backed by the given database connection.
func NewPostgresRepo(db *sql.DB) *PostgresRepo {
	return &PostgresRepo{db: db}
}

// CreateUser inserts a new user. Returns error if the username is already taken.
func (r *PostgresRepo) CreateUser(user *models.User) error {
	err := r.db.QueryRow(
		`INSERT INTO users (username, password_hash)
		 VALUES ($1, $2)
		 RETURNING id, created_at`,
		user.Username, user.PasswordHash,
	).Scan(&user.ID, &user.CreatedAt)
	if err != nil {
		return fmt.Errorf("create user: %w", err)
	}
	return nil
}

// GetByID retrieves a user by their primary key UUID.
func (r *PostgresRepo) GetByID(id string) (*models.User, error) {
	u := &models.User{}
	err := r.db.QueryRow(
		`SELECT id, username, password_hash, created_at
		 FROM users WHERE id = $1`, id,
	).Scan(&u.ID, &u.Username, &u.PasswordHash, &u.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get user by id: %w", err)
	}
	return u, nil
}

// GetByUsername retrieves a user by their unique username (used for login).
func (r *PostgresRepo) GetByUsername(username string) (*models.User, error) {
	u := &models.User{}
	err := r.db.QueryRow(
		`SELECT id, username, password_hash, created_at
		 FROM users WHERE username = $1`, username,
	).Scan(&u.ID, &u.Username, &u.PasswordHash, &u.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get user by username: %w", err)
	}
	return u, nil
}

// SaveMatch persists a completed match and all its player records in a single transaction.
func (r *PostgresRepo) SaveMatch(match *models.Match, players []models.MatchPlayer) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	err = tx.QueryRow(
		`INSERT INTO matches (lobby_code, theme, winner_id, started_at, finished_at)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id`,
		match.LobbyCode, match.Theme, nilIfEmpty(match.WinnerID),
		match.StartedAt, match.FinishedAt,
	).Scan(&match.ID)
	if err != nil {
		return fmt.Errorf("insert match: %w", err)
	}

	stmt, err := tx.Prepare(
		`INSERT INTO match_players (match_id, user_id, is_bot, score)
		 VALUES ($1, $2, $3, $4)`,
	)
	if err != nil {
		return fmt.Errorf("prepare match_players: %w", err)
	}
	defer stmt.Close()

	for _, mp := range players {
		_, err := stmt.Exec(match.ID, nilIfEmpty(mp.UserID), mp.IsBot, mp.Score)
		if err != nil {
			return fmt.Errorf("insert match_player: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}
	return nil
}

// GetMatchHistory returns the N most recent matches for a user, ordered newest first.
func (r *PostgresRepo) GetMatchHistory(userID string, limit int) ([]models.MatchSummary, error) {
	rows, err := r.db.Query(
		`SELECT m.id, m.theme, mp.score,
		        COALESCE(m.winner_id = $1, false) AS won,
		        m.finished_at
		 FROM matches m
		 JOIN match_players mp ON mp.match_id = m.id
		 WHERE mp.user_id = $1
		 ORDER BY m.finished_at DESC
		 LIMIT $2`,
		userID, limit,
	)
	if err != nil {
		return nil, fmt.Errorf("query match history: %w", err)
	}
	defer rows.Close()

	var history []models.MatchSummary
	for rows.Next() {
		var ms models.MatchSummary
		if err := rows.Scan(&ms.MatchID, &ms.Theme, &ms.Score, &ms.Won, &ms.FinishedAt); err != nil {
			return nil, fmt.Errorf("scan match summary: %w", err)
		}
		history = append(history, ms)
	}
	return history, rows.Err()
}

// GetWinLoss returns aggregate win/loss counts for a user across all their matches.
func (r *PostgresRepo) GetWinLoss(userID string) (wins, losses int, err error) {
	err = r.db.QueryRow(
		`SELECT
		   COUNT(*) FILTER (WHERE m.winner_id = $1) AS wins,
		   COUNT(*) FILTER (WHERE m.winner_id IS DISTINCT FROM $1) AS losses
		 FROM matches m
		 JOIN match_players mp ON mp.match_id = m.id
		 WHERE mp.user_id = $1`,
		userID,
	).Scan(&wins, &losses)
	if err != nil {
		return 0, 0, fmt.Errorf("get win/loss: %w", err)
	}
	return wins, losses, nil
}

// Compile-time interface check.
var _ models.UserRepository = (*PostgresRepo)(nil)

// nilIfEmpty converts an empty string to a sql.NullString with Valid=false,
// so PostgreSQL receives NULL instead of an empty UUID string.
func nilIfEmpty(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}
