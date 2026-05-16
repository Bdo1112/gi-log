package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type ExchangeStore interface {
	Save(id, sessionID, userMsg, assistantMsg string, vec []byte, entities []string) error
}

// LocalStore writes to SQLite on the user's machine.
type LocalStore struct{}

func (s LocalStore) Save(id, sessionID, userMsg, assistantMsg string, vec []byte, entities []string) error {
	return insertExchange(id, sessionID, userMsg, assistantMsg, vec, entities)
}

// CloudStore sends the exchange to the hosted API.
// vec and entities are ignored — the server recomputes them.
type CloudStore struct {
	baseURL string
	token   string
}

func (s CloudStore) Save(id, sessionID, userMsg, assistantMsg string, _ []byte, _ []string) error {
	body, err := json.Marshal(map[string]string{
		"user_id":             s.token,
		"external_session_id": sessionID,
		"platform":            "claude",
		"user_msg":            userMsg,
		"assistant_msg":       assistantMsg,
	})
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", s.baseURL+"/exchanges", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		raw, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("cloud save failed (%d): %s", resp.StatusCode, string(raw))
	}
	return nil
}

// MultiStore writes to primary (must succeed) then secondary (best effort, async).
type MultiStore struct {
	primary   ExchangeStore
	secondary ExchangeStore
}

func (s MultiStore) Save(id, sessionID, userMsg, assistantMsg string, vec []byte, entities []string) error {
	if err := s.primary.Save(id, sessionID, userMsg, assistantMsg, vec, entities); err != nil {
		return err
	}
	go func() {
		if err := s.secondary.Save(id, sessionID, userMsg, assistantMsg, vec, entities); err != nil {
			logError(fmt.Errorf("cloud sync: %w", err))
		}
	}()
	return nil
}

func newStore(cfg Config) ExchangeStore {
	local := LocalStore{}
	if cfg.AI.GiLogToken != "" && cfg.Server.APIURL != "" {
		return MultiStore{
			primary:   local,
			secondary: CloudStore{baseURL: cfg.Server.APIURL, token: cfg.AI.GiLogToken},
		}
	}
	return local
}
