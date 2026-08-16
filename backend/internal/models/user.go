package models

import "time"

// User represents a registered account. Maps to the `users` DB table.
type User struct {
	ID           string    `json:"id"            db:"id"`
	Username     string    `json:"username"      db:"username"`
	PasswordHash string    `json:"-"             db:"password_hash"` // Never serialized to JSON
	CreatedAt    time.Time `json:"createdAt"     db:"created_at"`
}

// Match represents a completed game session. Maps to the `matches` DB table.
type Match struct {
	ID         string    `json:"id"         db:"id"`
	LobbyCode  string    `json:"lobbyCode"  db:"lobby_code"`
	Theme      string    `json:"theme"      db:"theme"`
	WinnerID   string    `json:"winnerId"   db:"winner_id"`
	StartedAt  time.Time `json:"startedAt"  db:"started_at"`
	FinishedAt time.Time `json:"finishedAt" db:"finished_at"`
}

// MatchPlayer links a user (or bot) to a match. Maps to `match_players` DB table.
type MatchPlayer struct {
	MatchID string `json:"matchId" db:"match_id"`
	UserID  string `json:"userId"  db:"user_id"`
	IsBot   bool   `json:"isBot"   db:"is_bot"`
	Score   int    `json:"score"   db:"score"`
}

// UserProfile is the read model returned by GET /api/users/:id.
type UserProfile struct {
	User
	Wins    int            `json:"wins"`
	Losses  int            `json:"losses"`
	History []MatchSummary `json:"history"`
}

// MatchSummary is a lightweight match record for profile display.
type MatchSummary struct {
	MatchID    string    `json:"matchId"`
	Theme      string    `json:"theme"`
	Score      int       `json:"score"`
	Won        bool      `json:"won"`
	FinishedAt time.Time `json:"finishedAt"`
}

// UserRepository defines the persistence interface for user-related data.
// Implementations live in internal/db/.
type UserRepository interface {
	// CreateUser inserts a new user. Returns error if username is taken.
	CreateUser(user *User) error

	// GetByID retrieves a user by primary key.
	GetByID(id string) (*User, error)

	// GetByUsername retrieves a user by unique username (for login).
	GetByUsername(username string) (*User, error)

	// SaveMatch persists a completed match and its player records.
	SaveMatch(match *Match, players []MatchPlayer) error

	// GetMatchHistory returns the N most recent matches for a user.
	GetMatchHistory(userID string, limit int) ([]MatchSummary, error)

	// GetWinLoss returns aggregate win/loss counts for a user.
	GetWinLoss(userID string) (wins, losses int, err error)
}
