package main

import (
	"fmt"
	"net/http"
	"time"
)

func runStatus(cfg Config) error {
	fmt.Println("gi-log status")
	fmt.Println("─────────────────────────────")

	// config
	fmt.Println("\nConfig")
	if cfg.AI.GiLogToken != "" {
		fmt.Println("  mode:  gi-log token")
		fmt.Printf("  token: %s...%s\n", cfg.AI.GiLogToken[:8], cfg.AI.GiLogToken[len(cfg.AI.GiLogToken)-4:])
		fmt.Printf("  url:   %s\n", cfg.Server.APIURL)
	} else if cfg.AI.APIKey != "" {
		fmt.Println("  mode:  OpenAI direct")
		fmt.Printf("  key:   %s...%s\n", cfg.AI.APIKey[:7], cfg.AI.APIKey[len(cfg.AI.APIKey)-4:])
	} else {
		fmt.Println("  mode:  [NOT CONFIGURED] — set gi_log_token or api_key in ~/.gi-log/config.json")
	}

	// api reachability (token mode only)
	if cfg.AI.GiLogToken != "" {
		fmt.Println("\nAPI")
		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Get(cfg.Server.APIURL + "/health")
		if err != nil {
			fmt.Printf("  reachable: NO — %s\n", err)
		} else {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				fmt.Println("  reachable: YES")
			} else {
				fmt.Printf("  reachable: NO — HTTP %d\n", resp.StatusCode)
			}
		}
	}

	// local db stats
	fmt.Println("\nLocal database")
	exchanges, err := fetchAllExchanges()
	if err != nil {
		fmt.Printf("  error reading db: %s\n", err)
	} else {
		fmt.Printf("  exchanges: %d\n", len(exchanges))
	}
	sessions, err := fetchAllSessions()
	if err != nil {
		fmt.Printf("  error reading sessions: %s\n", err)
	} else {
		withSummary := 0
		for _, s := range sessions {
			if s.Summary != "" {
				withSummary++
			}
		}
		fmt.Printf("  sessions:  %d (%d with summary)\n", len(sessions), withSummary)
	}
	fmt.Printf("  db path:   %s\n", cfg.DB.Path)

	// hooks
	fmt.Println("\nHooks")
	settings, err := readSettings()
	if err != nil {
		fmt.Printf("  error reading settings.json: %s\n", err)
	} else {
		fmt.Printf("  UserPromptSubmit: %s\n", hookStatus(settings, "UserPromptSubmit", "gi-log hook-user-prompt"))
		fmt.Printf("  Stop:             %s\n", hookStatus(settings, "Stop", "gi-log hook-stop"))
	}

	fmt.Println()
	return nil
}

func hookStatus(settings map[string]any, event, command string) string {
	hooks, _ := settings["hooks"].(map[string]any)
	if hooks == nil {
		return "NOT registered"
	}
	entries, _ := hooks[event].([]any)
	for _, entry := range entries {
		em, _ := entry.(map[string]any)
		hookList, _ := em["hooks"].([]any)
		for _, h := range hookList {
			hm, _ := h.(map[string]any)
			if hm["command"] == command {
				return "registered"
			}
		}
	}
	return "NOT registered"
}
