package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

func runInstall() error {
	settingsPath := filepath.Join(os.Getenv("HOME"), ".claude", "settings.json")
	claudeJSONPath := filepath.Join(os.Getenv("HOME"), ".claude.json")

	// --- hooks: read/write settings.json ---
	settingsData, err := os.ReadFile(settingsPath)
	if err != nil {
		return fmt.Errorf("cannot read %s: %w", settingsPath, err)
	}
	var settings map[string]any
	if err := json.Unmarshal(settingsData, &settings); err != nil {
		return fmt.Errorf("invalid settings.json: %w", err)
	}

	hooksChanged := false
	hooksChanged = ensureHook(settings, "UserPromptSubmit", "gi-log hook-user-prompt") || hooksChanged
	hooksChanged = ensureHook(settings, "Stop", "gi-log hook-stop") || hooksChanged

	if hooksChanged {
		out, err := json.MarshalIndent(settings, "", "  ")
		if err != nil {
			return err
		}
		if err := os.WriteFile(settingsPath, out, 0644); err != nil {
			return fmt.Errorf("cannot write settings.json: %w", err)
		}
		fmt.Println("gi-log: hooks updated in settings.json")
	} else {
		fmt.Println("gi-log: hooks already registered")
	}

	// --- mcp: read/write ~/.claude.json ---
	claudeData, err := os.ReadFile(claudeJSONPath)
	if err != nil {
		return fmt.Errorf("cannot read %s: %w", claudeJSONPath, err)
	}
	var claudeJSON map[string]any
	if err := json.Unmarshal(claudeData, &claudeJSON); err != nil {
		return fmt.Errorf("invalid .claude.json: %w", err)
	}

	mcpChanged := ensureMCP(claudeJSON)

	if mcpChanged {
		out, err := json.MarshalIndent(claudeJSON, "", "  ")
		if err != nil {
			return err
		}
		if err := os.WriteFile(claudeJSONPath, out, 0644); err != nil {
			return fmt.Errorf("cannot write .claude.json: %w", err)
		}
		fmt.Println("gi-log: MCP server registered in ~/.claude.json")
	} else {
		fmt.Println("gi-log: MCP server already registered")
	}

	return nil
}

func ensureHook(settings map[string]any, event, command string) bool {
	hooks, _ := settings["hooks"].(map[string]any)
	if hooks == nil {
		hooks = map[string]any{}
		settings["hooks"] = hooks
	}

	entries, _ := hooks[event].([]any)
	if len(entries) == 0 {
		hooks[event] = []any{
			map[string]any{"hooks": []any{map[string]any{"type": "command", "command": command}}},
		}
		fmt.Printf("  added hook: %s → %s\n", event, command)
		return true
	}

	entry, _ := entries[0].(map[string]any)
	hookList, _ := entry["hooks"].([]any)
	for _, h := range hookList {
		hm, _ := h.(map[string]any)
		if hm["command"] == command {
			fmt.Printf("  already registered: %s → %s\n", event, command)
			return false
		}
	}

	entry["hooks"] = append(hookList, map[string]any{"type": "command", "command": command})
	entries[0] = entry
	hooks[event] = entries
	fmt.Printf("  added hook: %s → %s\n", event, command)
	return true
}

func ensureMCP(settings map[string]any) bool {
	mcpServers, _ := settings["mcpServers"].(map[string]any)
	if mcpServers == nil {
		mcpServers = map[string]any{}
		settings["mcpServers"] = mcpServers
	}

	if _, exists := mcpServers["gi-log"]; exists {
		fmt.Println("  already registered: mcpServers → gi-log")
		return false
	}

	mcpServers["gi-log"] = map[string]any{
		"command": "gi-log",
		"args":    []string{"mcp"},
	}
	fmt.Println("  added mcpServers → gi-log")
	return true
}
