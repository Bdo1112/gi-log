package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type userPromptPayload struct {
	SessionID     string `json:"session_id"`
	Prompt        string `json:"prompt"`
	HookEventName string `json:"hook_event_name"`
}

type stopPayload struct {
	SessionID            string `json:"session_id"`
	LastAssistantMessage string `json:"last_assistant_message"`
	HookEventName        string `json:"hook_event_name"`
}

type tmpData struct {
	SessionID string `json:"session_id"`
	UserMsg   string `json:"user_msg"`
}

func tmpDir() string {
	return filepath.Join(giLogDir(), "tmp")
}

func tmpFile(sessionID string) string {
	return filepath.Join(tmpDir(), sessionID+".json")
}

// runHookUserPrompt handles the UserPromptSubmit hook.
// It stores the user message in a temp file keyed by session_id.
func runHookUserPrompt() error {
	var payload userPromptPayload
	if err := json.NewDecoder(os.Stdin).Decode(&payload); err != nil {
		return fmt.Errorf("invalid payload: %w", err)
	}
	if payload.SessionID == "" || payload.Prompt == "" {
		return fmt.Errorf("missing session_id or prompt")
	}

	if err := os.MkdirAll(tmpDir(), 0755); err != nil {
		return err
	}

	data, _ := json.Marshal(tmpData{
		SessionID: payload.SessionID,
		UserMsg:   payload.Prompt,
	})
	return os.WriteFile(tmpFile(payload.SessionID), data, 0644)
}

// runHookStop handles the Stop hook.
// It reads the temp file for the user message, combines with the assistant message,
// embeds the pair, and inserts into the DB.
func runHookStop(cfg Config) error {
	var payload stopPayload
	if err := json.NewDecoder(os.Stdin).Decode(&payload); err != nil {
		return fmt.Errorf("invalid payload: %w", err)
	}
	if payload.SessionID == "" || payload.LastAssistantMessage == "" {
		return fmt.Errorf("missing session_id or last_assistant_message")
	}

	// read the user message from the temp file
	tmp := tmpFile(payload.SessionID)
	raw, err := os.ReadFile(tmp)
	if err != nil {
		return fmt.Errorf("no user message found for session %s: %w", payload.SessionID, err)
	}

	var td tmpData
	if err := json.Unmarshal(raw, &td); err != nil {
		return fmt.Errorf("invalid tmp file: %w", err)
	}

	// embed the combined text
	combined := "User: " + td.UserMsg + " Assistant: " + payload.LastAssistantMessage
	vec, err := embedText(combined, cfg.Embedding.Model, cfg.Embedding.APIKey)
	if err != nil {
		return fmt.Errorf("embed: %w", err)
	}

	// insert into DB
	id := newID()
	if err := insertExchange(id, payload.SessionID, td.UserMsg, payload.LastAssistantMessage, toBytes(vec)); err != nil {
		return fmt.Errorf("db: %w", err)
	}

	// clean up temp file
	os.Remove(tmp)

	fmt.Printf(`{"success":true,"id":%q}`, id)
	fmt.Println()
	return nil
}
