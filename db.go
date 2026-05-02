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
			id           TEXT PRIMARY KEY,
			session_id   TEXT NOT NULL,
			user_msg     TEXT NOT NULL,
			assistant_msg TEXT NOT NULL,
			embedding    BLOB NOT NULL,
			created_at   TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
		CREATE TABLE IF NOT EXISTS sessions (
			id         TEXT PRIMARY KEY,
			session_id TEXT NOT NULL,
			summary    TEXT NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
	`)
	return err
}

func insertExchange(id, sessionID, userMsg, assistantMsg string, embedding []byte) error {
	_, err := db.Exec(
		"INSERT INTO exchanges (id, session_id, user_msg, assistant_msg, embedding) VALUES (?, ?, ?, ?, ?)",
		id, sessionID, userMsg, assistantMsg, embedding,
	)
	return err
}

func fetchAllExchanges() ([]Exchange, error) {
	rows, err := db.Query("SELECT id, session_id, user_msg, assistant_msg, embedding, created_at FROM exchanges")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []Exchange
	for rows.Next() {
		var e Exchange
		if err := rows.Scan(&e.ID, &e.SessionID, &e.UserMsg, &e.AssistantMsg, &e.Embedding, &e.CreatedAt); err != nil {
			return nil, err
		}
		result = append(result, e)
	}
	return result, rows.Err()
}
