package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

func main() {
	if err := initGiLogDir(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(1)
	}

	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "Usage: gi-log <save|end-session|search|mcp>")
		os.Exit(1)
	}

	if os.Args[1] == "install" {
		if err := runInstall(); err != nil {
			fmt.Fprintf(os.Stderr, "install error: %s\n", err)
			os.Exit(1)
		}
		return
	}

	cfg, err := loadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "config error: %s\n", err)
		os.Exit(1)
	}

	if err := initDB(cfg.DB.Path); err != nil {
		logError(err)
		fmt.Fprintf(os.Stderr, "db error: %s\n", err)
		os.Exit(1)
	}
	fmt.Println(os.Args[1])
	switch os.Args[1] {
	case "save":
		if err := runSave(cfg); err != nil {
			logError(err)
			fmt.Fprintf(os.Stderr, "save error: %s\n", err)
			os.Exit(1)
		}
	case "end-session":
		if err := runEndSession(cfg); err != nil {
			logError(err)
			fmt.Fprintf(os.Stderr, "end-session error: %s\n", err)
			os.Exit(1)
		}
	case "search":
		if err := runSearch(cfg); err != nil {
			logError(err)
			fmt.Fprintf(os.Stderr, "search error: %s\n", err)
			os.Exit(1)
		}
	case "hook-session-end":
		if err := runHookSessionEnd(cfg); err != nil {
			logError(err)
			fmt.Fprintf(os.Stderr, "hook-session-end error: %s\n", err)
			os.Exit(1)
		}
	case "hook-user-prompt":
		if err := runHookUserPrompt(); err != nil {
			logError(err)
			fmt.Fprintf(os.Stderr, "hook-user-prompt error: %s\n", err)
			os.Exit(1)
		}
	case "hook-stop":
		if err := runHookStop(cfg); err != nil {
			logError(err)
			fmt.Fprintf(os.Stderr, "hook-stop error: %s\n", err)
			os.Exit(1)
		}
	case "mcp":
		runMCP(cfg)
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", os.Args[1])
		os.Exit(1)
	}
}

func runSave(cfg Config) error {
	var input struct {
		SessionID    string `json:"session_id"`
		UserMsg      string `json:"user_msg"`
		AssistantMsg string `json:"assistant_msg"`
	}
	if err := json.NewDecoder(os.Stdin).Decode(&input); err != nil {
		return fmt.Errorf("invalid input: %w", err)
	}
	if input.SessionID == "" || input.UserMsg == "" || input.AssistantMsg == "" {
		return fmt.Errorf("session_id, user_msg, and assistant_msg are required")
	}

	if err := saveExchange(input.SessionID, input.UserMsg, input.AssistantMsg, cfg); err != nil {
		return err
	}

	fmt.Printf(`{"success":true}`)
	fmt.Println()
	return nil
}

func runEndSession(cfg Config) error {
	var input struct {
		SessionID      string `json:"session_id"`
		TranscriptPath string `json:"transcript_path"`
	}
	if err := json.NewDecoder(os.Stdin).Decode(&input); err != nil {
		return fmt.Errorf("invalid input: %w", err)
	}
	// TODO: read transcript, generate summary via OpenAI chat completions, store in sessions table
	fmt.Printf(`{"success":true,"session_id":%q,"note":"summary not yet implemented"}`, input.SessionID)
	fmt.Println()
	return nil
}

func runSearch(cfg Config) error {
	var input struct {
		Query string `json:"query"`
	}
	if err := json.NewDecoder(os.Stdin).Decode(&input); err != nil {
		return fmt.Errorf("invalid input: %w", err)
	}
	if input.Query == "" {
		return fmt.Errorf("query is required")
	}

	results, err := doSearch(input.Query, cfg)
	if err != nil {
		return err
	}

	fmt.Println(formatResults(results))
	return nil
}

func doSearch(query string, cfg Config) ([]SearchResult, error) {
	// extract keywords first to decide pipeline
	keywords, err := cfg.Client.Extract(query)
	if err != nil {
		return nil, fmt.Errorf("extract: %w", err)
	}

	vec, err := cfg.Client.Embed(query)
	if err != nil {
		return nil, fmt.Errorf("embed: %w", err)
	}

	ctx := SearchContext{
		RawQuery: query,
		QueryVec: vec,
		Keywords: keywords,
	}

	sessions, err := fetchAllSessions()
	if err != nil {
		return nil, fmt.Errorf("db sessions: %w", err)
	}

	exchanges, err := fetchAllExchanges()
	if err != nil {
		return nil, fmt.Errorf("db exchanges: %w", err)
	}

	// try keyword pipeline first
	if len(keywords) > 0 {
		results := KeywordPipeline{}.Search(ctx, sessions, exchanges, cfg.Search.TopK)
		if len(results) > 0 {
			return results, nil
		}
	}

	// fall back to semantic pipeline
	return SemanticPipeline{}.Search(ctx, exchanges, cfg.Search.TopK), nil
}

func formatResults(results []SearchResult) string {
	if len(results) == 0 {
		return "No relevant memories found."
	}
	out := ""
	for i, r := range results {
		out += fmt.Sprintf("[Memory %d - similarity: %.2f]\nUser: %s\nAssistant: %s\n\n",
			i+1, r.Similarity, r.UserMsg, r.AssistantMsg)
	}
	return out
}

func logError(err error) {
	dir := filepath.Join(giLogDir(), "errors")
	os.MkdirAll(dir, 0755)
	fname := filepath.Join(dir, time.Now().Format("2006-01-02T15-04-05")+".log")
	os.WriteFile(fname, []byte(err.Error()+"\n"), 0644)
}

func newID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}
