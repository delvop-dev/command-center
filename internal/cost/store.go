package cost

import (
	"database/sql"
	"time"

	_ "modernc.org/sqlite"
)

type Store struct {
	db *sql.DB
}

func NewStore(dbPath string) (*Store, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS cost_entries (
		session_id TEXT PRIMARY KEY,
		cost_usd REAL,
		tokens_in INTEGER,
		tokens_out INTEGER,
		updated_at DATETIME
	)`)
	if err != nil {
		db.Close()
		return nil, err
	}
	return &Store{db: db}, nil
}

func (s *Store) Upsert(sessionID string, costUSD float64, tokensIn, tokensOut int64) error {
	_, err := s.db.Exec(`INSERT INTO cost_entries (session_id, cost_usd, tokens_in, tokens_out, updated_at)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(session_id) DO UPDATE SET cost_usd=?, tokens_in=?, tokens_out=?, updated_at=?`,
		sessionID, costUSD, tokensIn, tokensOut, time.Now(),
		costUSD, tokensIn, tokensOut, time.Now())
	return err
}

func (s *Store) Close() error {
	return s.db.Close()
}
