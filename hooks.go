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

func runHookSessionEnd(cfg Config) error {
	var payload struct {
		SessionID     string `json:"session_id"`
		HookEventName string `json:"hook_event_name"`
	}
	if err := json.NewDecoder(os.Stdin).Decode(&payload); err != nil {
		return fmt.Errorf("invalid payload: %w", err)
	}
	if payload.SessionID == "" {
		return fmt.Errorf("missing session_id")
	}

	return summarizeSession(payload.SessionID, cfg)
}

func runHookStop(cfg Config) error {
	var payload stopPayload
	if err := json.NewDecoder(os.Stdin).Decode(&payload); err != nil {
		return fmt.Errorf("invalid payload: %w", err)
	}
	if payload.SessionID == "" || payload.LastAssistantMessage == "" {
		return fmt.Errorf("missing session_id or last_assistant_message")
	}

	tmp := tmpFile(payload.SessionID)
	raw, err := os.ReadFile(tmp)
	if err != nil {
		return fmt.Errorf("no user message found for session %s: %w", payload.SessionID, err)
	}

	var td tmpData
	if err := json.Unmarshal(raw, &td); err != nil {
		return fmt.Errorf("invalid tmp file: %w", err)
	}

	if err := saveExchange(payload.SessionID, td.UserMsg, payload.LastAssistantMessage, cfg); err != nil {
		return err
	}

	os.Remove(tmp)

	fmt.Printf(`{"success":true,"id":%q}`, payload.SessionID)
	fmt.Println()
	return nil
}
