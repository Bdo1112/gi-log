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
			id         TEXT PRIMARY KEY,
			session_id TEXT NOT NULL,
			summary    TEXT NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
	`)
	if err != nil {
		return err
	}

	// migrate existing DB — add entities column if missing
	db.Exec("ALTER TABLE exchanges ADD COLUMN entities TEXT DEFAULT '[]'")

	return nil
}

func insertExchange(id, sessionID, userMsg, assistantMsg string, embedding []byte, entities []string) error {
	_, err := db.Exec(
		"INSERT INTO exchanges (id, session_id, user_msg, assistant_msg, embedding, entities) VALUES (?, ?, ?, ?, ?, ?)",
		id, sessionID, userMsg, assistantMsg, embedding, entitiesToJSON(entities),
	)
	return err
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
