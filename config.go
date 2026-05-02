package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Config struct {
	Embedding EmbeddingConfig `json:"embedding"`
	DB        DBConfig        `json:"db"`
	Search    SearchConfig    `json:"search"`
}

type EmbeddingConfig struct {
	APIKey string `json:"api_key"`
	Model  string `json:"model"`
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
			Embedding: EmbeddingConfig{
				APIKey: "",
				Model:  "text-embedding-3-small",
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
		fmt.Printf("gi-log: config created at %s\nPlease set embedding.api_key before using.\n", cfgPath)
	}

	return nil
}

func loadConfig() (Config, error) {
	home, _ := os.UserHomeDir()
	cfg := Config{
		Embedding: EmbeddingConfig{Model: "text-embedding-3-small"},
		DB:        DBConfig{Path: filepath.Join(home, ".gi-log", "gi_log.db")},
		Search:    SearchConfig{TopK: 5},
	}

	data, err := os.ReadFile(configPath())
	if err != nil {
		return cfg, fmt.Errorf("cannot read config at %s: %w", configPath(), err)
	}

	if err := json.Unmarshal(data, &cfg); err != nil {
		return cfg, fmt.Errorf("invalid config: %w", err)
	}

	if cfg.Embedding.APIKey == "" {
		return cfg, fmt.Errorf("embedding.api_key is not set in %s", configPath())
	}
	if cfg.Embedding.Model == "" {
		cfg.Embedding.Model = "text-embedding-3-small"
	}
	if cfg.Search.TopK == 0 {
		cfg.Search.TopK = 5
	}

	return cfg, nil
}
