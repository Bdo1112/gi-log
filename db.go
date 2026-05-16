package main

import (
	"database/sql"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

type Exchange struct {
	ID           string
	SessionID    string
	UserMsg      string
	AssistantMsg string
	Embedding    []byte
	Entities     []string
	CreatedAt    string
}

type Session struct {
	SessionID string
	Summary   string
	Entities  []string
	Embedding []byte
	CreatedAt string
	UpdatedAt string
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
		CREATE TABLE IF NOT EXISTS exchanges (
			id            TEXT PRIMARY KEY,
			session_id    TEXT NOT NULL,
			user_msg      TEXT NOT NULL,
			assistant_msg TEXT NOT NULL,
			embedding     BLOB NOT NULL,
			entities      TEXT DEFAULT '[]',
			created_at    TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
		CREATE TABLE IF NOT EXISTS sessions (
			session_id TEXT PRIMARY KEY,
			summary    TEXT NOT NULL DEFAULT '',
			entities   TEXT NOT NULL DEFAULT '[]',
			embedding  BLOB,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
	`)
	if err != nil {
		return err
	}

	// migrations for existing DBs
	db.Exec("ALTER TABLE exchanges ADD COLUMN entities TEXT DEFAULT '[]'")
	db.Exec("ALTER TABLE sessions ADD COLUMN entities TEXT NOT NULL DEFAULT '[]'")
	db.Exec("ALTER TABLE sessions ADD COLUMN embedding BLOB")

	return nil
}

func insertExchange(id, sessionID, userMsg, assistantMsg string, embedding []byte, entities []string) error {
	_, err := db.Exec(
		"INSERT INTO exchanges (id, session_id, user_msg, assistant_msg, embedding, entities) VALUES (?, ?, ?, ?, ?, ?)",
		id, sessionID, userMsg, assistantMsg, embedding, entitiesToJSON(entities),
	)
	return err
}

func fetchSession(sessionID string) (*Session, error) {
	row := db.QueryRow("SELECT session_id, summary, entities, embedding, created_at, updated_at FROM sessions WHERE session_id = ?", sessionID)
	var s Session
	var entitiesJSON string
	err := row.Scan(&s.SessionID, &s.Summary, &entitiesJSON, &s.Embedding, &s.CreatedAt, &s.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	s.Entities = entitiesFromJSON(entitiesJSON)
	return &s, nil
}

func upsertSession(sessionID, summary string, entities []string, embedding []byte) error {
	_, err := db.Exec(`
		INSERT INTO sessions (session_id, summary, entities, embedding)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(session_id) DO UPDATE SET
			summary    = excluded.summary,
			entities   = excluded.entities,
			embedding  = excluded.embedding,
			updated_at = CURRENT_TIMESTAMP
	`, sessionID, summary, entitiesToJSON(entities), embedding)
	return err
}

func fetchAllSessions() ([]Session, error) {
	rows, err := db.Query("SELECT session_id, summary, entities, embedding, created_at, updated_at FROM sessions")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []Session
	for rows.Next() {
		var s Session
		var entitiesJSON string
		if err := rows.Scan(&s.SessionID, &s.Summary, &entitiesJSON, &s.Embedding, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, err
		}
		s.Entities = entitiesFromJSON(entitiesJSON)
		result = append(result, s)
	}
	return result, rows.Err()
}

func countExchangesBySession(sessionID string) (int, error) {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM exchanges WHERE session_id = ?", sessionID).Scan(&count)
	return count, err
}

func fetchExchangesBySession(sessionID string) ([]Exchange, error) {
	rows, err := db.Query(
		"SELECT id, session_id, user_msg, assistant_msg, embedding, entities, created_at FROM exchanges WHERE session_id = ? ORDER BY created_at ASC",
		sessionID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []Exchange
	for rows.Next() {
		var e Exchange
		var entitiesJSON string
		if err := rows.Scan(&e.ID, &e.SessionID, &e.UserMsg, &e.AssistantMsg, &e.Embedding, &entitiesJSON, &e.CreatedAt); err != nil {
			return nil, err
		}
		e.Entities = entitiesFromJSON(entitiesJSON)
		result = append(result, e)
	}
	return result, rows.Err()
}

func fetchAllExchanges() ([]Exchange, error) {
	rows, err := db.Query("SELECT id, session_id, user_msg, assistant_msg, embedding, entities, created_at FROM exchanges")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []Exchange
	for rows.Next() {
		var e Exchange
		var entitiesJSON string
		if err := rows.Scan(&e.ID, &e.SessionID, &e.UserMsg, &e.AssistantMsg, &e.Embedding, &entitiesJSON, &e.CreatedAt); err != nil {
			return nil, err
		}
		e.Entities = entitiesFromJSON(entitiesJSON)
		result = append(result, e)
	}
	return result, rows.Err()
}
