package main

import (
	"database/sql"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

const localUserID = "local"

type Exchange struct {
	ID            string
	UserID        string
	SessionID     string
	SequenceIndex int
	UserMsg       string
	AssistantMsg  string
	Embedding     []byte
	Entities      []string
	CreatedAt     string
}

type Session struct {
	ID                 string // UUID primary key
	ExternalSessionID  string // logical session key (e.g. Claude conversation ID)
	UserID             string
	Summary            string
	Entities           []string
	Embedding          []byte
	CreatedAt          string
	UpdatedAt          string
}

var db *sql.DB

func initDB(path string) error {
	if len(path) > 1 && path[:2] == "~/" {
		home, _ := os.UserHomeDir()
		path = filepath.Join(home, path[2:])
	}

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	var err error
	db, err = sql.Open("sqlite", path)
	if err != nil {
		return err
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS sessions (
			id                   TEXT PRIMARY KEY,
			user_id              TEXT NOT NULL,
			external_session_id  TEXT NOT NULL,
			platform             TEXT NOT NULL DEFAULT 'claude',
			title                TEXT,
			summary              TEXT NOT NULL DEFAULT '',
			entities             TEXT NOT NULL DEFAULT '[]',
			embedding            BLOB,
			exchange_count       INTEGER NOT NULL DEFAULT 0,
			started_at           TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			last_exchange_at     TIMESTAMP,
			created_at           TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at           TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			UNIQUE (user_id, platform, external_session_id)
		);
		CREATE TABLE IF NOT EXISTS exchanges (
			id             TEXT PRIMARY KEY,
			user_id        TEXT NOT NULL,
			session_id     TEXT NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
			sequence_index INTEGER NOT NULL,
			user_msg       TEXT NOT NULL,
			assistant_msg  TEXT NOT NULL,
			embedding      BLOB,
			created_at     TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			UNIQUE (session_id, sequence_index)
		);
		CREATE TABLE IF NOT EXISTS tokens (
			token           TEXT PRIMARY KEY,
			label           TEXT NOT NULL DEFAULT '',
			requests_today  INTEGER NOT NULL DEFAULT 0,
			last_reset_date TEXT NOT NULL DEFAULT (date('now')),
			created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
		CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON sessions(user_id);
		CREATE INDEX IF NOT EXISTS idx_sessions_user_platform ON sessions(user_id, platform);
		CREATE INDEX IF NOT EXISTS idx_exchanges_session_id ON exchanges(session_id);
		CREATE INDEX IF NOT EXISTS idx_exchanges_user_id ON exchanges(user_id);
	`)
	return err
}

func insertExchange(id, externalSessionID, userMsg, assistantMsg string, embedding []byte, entities []string) error {
	session, err := findOrCreateSession(externalSessionID)
	if err != nil {
		return err
	}

	var seqIndex int
	err = db.QueryRow("SELECT COUNT(*) FROM exchanges WHERE session_id = ?", session.ID).Scan(&seqIndex)
	if err != nil {
		return err
	}

	_, err = db.Exec(
		"INSERT INTO exchanges (id, user_id, session_id, sequence_index, user_msg, assistant_msg, embedding) VALUES (?, ?, ?, ?, ?, ?, ?)",
		id, localUserID, session.ID, seqIndex, userMsg, assistantMsg, embedding,
	)
	if err != nil {
		return err
	}

	_, err = db.Exec(
		"UPDATE sessions SET exchange_count = exchange_count + 1, last_exchange_at = CURRENT_TIMESTAMP WHERE id = ?",
		session.ID,
	)
	return err
}

func findOrCreateSession(externalSessionID string) (*Session, error) {
	_, err := db.Exec(
		"INSERT OR IGNORE INTO sessions (id, user_id, external_session_id, platform) VALUES (lower(hex(randomblob(4)))||'-'||lower(hex(randomblob(2)))||'-4'||substr(lower(hex(randomblob(2))),2)||'-'||substr('89ab',abs(random())%4+1,1)||substr(lower(hex(randomblob(2))),2)||'-'||lower(hex(randomblob(6))), ?, ?, 'claude')",
		localUserID, externalSessionID,
	)
	if err != nil {
		return nil, err
	}

	return fetchSession(externalSessionID)
}

func fetchSession(externalSessionID string) (*Session, error) {
	row := db.QueryRow(
		"SELECT id, user_id, external_session_id, summary, entities, embedding, created_at, updated_at FROM sessions WHERE user_id = ? AND external_session_id = ?",
		localUserID, externalSessionID,
	)
	var s Session
	var entitiesJSON string
	err := row.Scan(&s.ID, &s.UserID, &s.ExternalSessionID, &s.Summary, &entitiesJSON, &s.Embedding, &s.CreatedAt, &s.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	s.Entities = entitiesFromJSON(entitiesJSON)
	return &s, nil
}

func upsertSession(externalSessionID, summary string, entities []string, embedding []byte) error {
	session, err := findOrCreateSession(externalSessionID)
	if err != nil {
		return err
	}
	_, err = db.Exec(`
		UPDATE sessions SET
			summary    = ?,
			entities   = ?,
			embedding  = ?,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`, summary, entitiesToJSON(entities), embedding, session.ID)
	return err
}

func fetchAllSessions() ([]Session, error) {
	rows, err := db.Query(
		"SELECT id, user_id, external_session_id, summary, entities, embedding, created_at, updated_at FROM sessions WHERE user_id = ?",
		localUserID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []Session
	for rows.Next() {
		var s Session
		var entitiesJSON string
		if err := rows.Scan(&s.ID, &s.UserID, &s.ExternalSessionID, &s.Summary, &entitiesJSON, &s.Embedding, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, err
		}
		s.Entities = entitiesFromJSON(entitiesJSON)
		result = append(result, s)
	}
	return result, rows.Err()
}

func countExchangesBySession(externalSessionID string) (int, error) {
	session, err := fetchSession(externalSessionID)
	if err != nil || session == nil {
		return 0, err
	}
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM exchanges WHERE session_id = ?", session.ID).Scan(&count)
	return count, err
}

func fetchExchangesBySession(externalSessionID string) ([]Exchange, error) {
	session, err := fetchSession(externalSessionID)
	if err != nil || session == nil {
		return nil, err
	}
	rows, err := db.Query(
		"SELECT id, user_id, session_id, sequence_index, user_msg, assistant_msg, embedding, created_at FROM exchanges WHERE session_id = ? ORDER BY sequence_index ASC",
		session.ID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []Exchange
	for rows.Next() {
		var e Exchange
		if err := rows.Scan(&e.ID, &e.UserID, &e.SessionID, &e.SequenceIndex, &e.UserMsg, &e.AssistantMsg, &e.Embedding, &e.CreatedAt); err != nil {
			return nil, err
		}
		result = append(result, e)
	}
	return result, rows.Err()
}

func fetchAllExchanges() ([]Exchange, error) {
	rows, err := db.Query(
		"SELECT id, user_id, session_id, sequence_index, user_msg, assistant_msg, embedding, created_at FROM exchanges WHERE user_id = ?",
		localUserID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []Exchange
	for rows.Next() {
		var e Exchange
		if err := rows.Scan(&e.ID, &e.UserID, &e.SessionID, &e.SequenceIndex, &e.UserMsg, &e.AssistantMsg, &e.Embedding, &e.CreatedAt); err != nil {
			return nil, err
		}
		result = append(result, e)
	}
	return result, rows.Err()
}
