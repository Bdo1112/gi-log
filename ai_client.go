package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type AIClient interface {
	Embed(text string) ([]float32, error)
	Extract(text string) ([]string, error)
	Summarize(text string) (string, error)
}

// OpenAIClient calls OpenAI directly using the user's own API key.
type OpenAIClient struct {
	ai AIConfig
}

func (c OpenAIClient) Embed(text string) ([]float32, error) {
	return Embedder{}.Process(text, c.ai.EmbeddingModel, c.ai.APIKey)
}

func (c OpenAIClient) Extract(text string) ([]string, error) {
	return Extractor{}.Process(text, c.ai.ExtractionModel, c.ai.APIKey)
}

func (c OpenAIClient) Summarize(text string) (string, error) {
	return Summarizer{}.Process(text, c.ai.ExtractionModel, c.ai.APIKey)
}

// GiLogClient calls the hosted gi-log API — no OpenAI key needed.
type GiLogClient struct {
	token   string
	baseURL string
}

func (c GiLogClient) Embed(text string) ([]float32, error) {
	var resp struct {
		Embedding []float32 `json:"embedding"`
	}
	if err := c.post("/ai/embed", map[string]string{"text": text}, &resp); err != nil {
		return nil, err
	}
	return resp.Embedding, nil
}

func (c GiLogClient) Extract(text string) ([]string, error) {
	var resp struct {
		Keywords []string `json:"keywords"`
	}
	if err := c.post("/ai/extract", map[string]string{"text": text}, &resp); err != nil {
		return nil, err
	}
	return resp.Keywords, nil
}

func (c GiLogClient) Summarize(text string) (string, error) {
	var resp struct {
		Summary string `json:"summary"`
	}
	// TODO: add retry logic for transient failures
	if err := c.post("/ai/summarize", map[string]string{"text": text}, &resp); err != nil {
		return "", err
	}
	return resp.Summary, nil
}

func (c GiLogClient) post(path string, body any, out any) error {
	data, err := json.Marshal(body)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", c.baseURL+path, bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		var e struct {
			Detail string `json:"detail"`
		}
		json.Unmarshal(raw, &e)
		return fmt.Errorf("gi-log API %s: %s", path, e.Detail)
	}

	return json.Unmarshal(raw, out)
}

func newAIClient(cfg Config) (AIClient, error) {
	if cfg.AI.GiLogToken != "" {
		if cfg.Server.APIURL == "" {
			return nil, fmt.Errorf("server.api_url is not set in %s", configPath())
		}
		return GiLogClient{token: cfg.AI.GiLogToken, baseURL: cfg.Server.APIURL}, nil
	}
	if cfg.AI.APIKey != "" {
		return OpenAIClient{ai: cfg.AI}, nil
	}
	return nil, fmt.Errorf("no AI credentials set — add gi_log_token or api_key to %s", configPath())
}
