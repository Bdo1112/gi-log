package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Config struct {
	AI     AIConfig     `json:"ai"`
	Server ServerConfig `json:"server"`
	DB     DBConfig     `json:"db"`
	Search SearchConfig `json:"search"`
	Client AIClient     `json:"-"`
}

type AIConfig struct {
	APIKey          string `json:"api_key"`
	GiLogToken      string `json:"gi_log_token"`
	EmbeddingModel  string `json:"embedding_model"`
	ExtractionModel string `json:"extraction_model"`
}

type ServerConfig struct {
	APIURL string `json:"api_url"`
}

type DBConfig struct {
	Path string `json:"path"`
}

type SearchConfig struct {
	TopK int `json:"top_k"`
}

func giLogDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".gi-log")
}

func configPath() string {
	return filepath.Join(giLogDir(), "config.json")
}

func initGiLogDir() error {
	dir := giLogDir()
	if err := os.MkdirAll(filepath.Join(dir, "errors"), 0755); err != nil {
		return err
	}

	cfgPath := configPath()
	if _, err := os.Stat(cfgPath); os.IsNotExist(err) {
		empty := Config{
			AI: AIConfig{
				APIKey:          "",
				GiLogToken:      "",
				EmbeddingModel:  "text-embedding-3-small",
				ExtractionModel: "gpt-4o-mini",
			},
			Server: ServerConfig{
				APIURL: "https://gi-log-api-production.up.railway.app",
			},
			DB: DBConfig{
				Path: filepath.Join(dir, "gi_log.db"),
			},
			Search: SearchConfig{
				TopK: 5,
			},
		}
		data, _ := json.MarshalIndent(empty, "", "  ")
		if err := os.WriteFile(cfgPath, data, 0600); err != nil {
			return err
		}
		fmt.Printf("gi-log: config created at %s\nSet ai.gi_log_token + server.api_url, or ai.api_key before using.\n", cfgPath)
	}

	return nil
}

func loadConfig() (Config, error) {
	home, _ := os.UserHomeDir()
	cfg := Config{
		AI: AIConfig{
			EmbeddingModel:  "text-embedding-3-small",
			ExtractionModel: "gpt-4o-mini",
		},
		DB:     DBConfig{Path: filepath.Join(home, ".gi-log", "gi_log.db")},
		Search: SearchConfig{TopK: 5},
	}

	data, err := os.ReadFile(configPath())
	if err != nil {
		return cfg, fmt.Errorf("cannot read config at %s: %w", configPath(), err)
	}

	if err := json.Unmarshal(data, &cfg); err != nil {
		return cfg, fmt.Errorf("invalid config: %w", err)
	}

	if strings.HasPrefix(cfg.DB.Path, "~/") {
		cfg.DB.Path = filepath.Join(home, cfg.DB.Path[2:])
	}

	if cfg.AI.APIKey == "" {
		cfg.AI.APIKey = os.Getenv("OPENAI_API_KEY")
	}
	if cfg.AI.GiLogToken == "" {
		cfg.AI.GiLogToken = os.Getenv("GI_LOG_TOKEN")
	}
	if cfg.AI.EmbeddingModel == "" {
		cfg.AI.EmbeddingModel = "text-embedding-3-small"
	}
	if cfg.AI.ExtractionModel == "" {
		cfg.AI.ExtractionModel = "gpt-4o-mini"
	}
	if cfg.Search.TopK == 0 {
		cfg.Search.TopK = 5
	}
	if cfg.Server.APIURL == "" || cfg.Server.APIURL == "https://gi-log-api-production." {
		cfg.Server.APIURL = "https://gi-log-api-production.up.railway.app"
	}

	writeConfigDefaults(cfg)

	client, err := newAIClient(cfg)
	if err != nil {
		return cfg, err
	}
	cfg.Client = client

	return cfg, nil
}

// writeConfigDefaults writes resolved defaults back to disk so the config file
// always shows all fields. Skips gi_log_token — that's user-supplied.
func writeConfigDefaults(cfg Config) {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return
	}
	os.WriteFile(configPath(), data, 0600)
}

// applyConfigDefaults reads the on-disk config, unconditionally sets managed
// defaults (server.api_url, db.path, model names), and writes it back.
// Called by runInstall so make install always repairs the config.
func applyConfigDefaults() error {
	home, _ := os.UserHomeDir()

	cfgPath := configPath()
	data, err := os.ReadFile(cfgPath)
	if err != nil {
		return err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return err
	}

	cfg.Server.APIURL = "https://gi-log-api-production.up.railway.app"
	if cfg.DB.Path == "" || strings.HasPrefix(cfg.DB.Path, "~/") {
		cfg.DB.Path = filepath.Join(home, ".gi-log", "gi_log.db")
	}
	if cfg.AI.EmbeddingModel == "" {
		cfg.AI.EmbeddingModel = "text-embedding-3-small"
	}
	if cfg.AI.ExtractionModel == "" {
		cfg.AI.ExtractionModel = "gpt-4o-mini"
	}
	if cfg.Search.TopK == 0 {
		cfg.Search.TopK = 5
	}

	out, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(cfgPath, out, 0600)
}
