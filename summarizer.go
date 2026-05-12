package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"

	"github.com/Bdo1112/gi-log/util/prompt"
)

type Summarizer struct{}

func (s Summarizer) Process(conversationText, model, apiKey string) (string, error) {
	body, err := json.Marshal(map[string]any{
		"model": model,
		"messages": []map[string]string{
			{"role": "system", "content": prompt.SummarizePrompt},
			{"role": "user", "content": conversationText},
		},
	})
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Error *struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return "", err
	}

	if result.Error != nil {
		return "", fmt.Errorf("openai: %s", result.Error.Message)
	}
	if len(result.Choices) == 0 {
		return "", fmt.Errorf("openai: no choices returned")
	}

	return strings.TrimSpace(result.Choices[0].Message.Content), nil
}

func summarizeSession(sessionID string, cfg Config) error {
	exchanges, err := fetchExchangesBySession(sessionID)
	if err != nil {
		return fmt.Errorf("fetch exchanges: %w", err)
	}
	if len(exchanges) == 0 {
		return nil
	}

	var sb strings.Builder
	for _, e := range exchanges {
		sb.WriteString("User: " + e.UserMsg + "\nAssistant: " + e.AssistantMsg + "\n\n")
	}

	summary, err := Summarizer{}.Process(sb.String(), cfg.AI.ExtractionModel, cfg.AI.APIKey)
	if err != nil {
		return fmt.Errorf("summarize: %w", err)
	}

	var sessionVec []float32
	var sessionEntities []string
	var embedErr, extractErr error

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		sessionVec, embedErr = Embedder{}.Process(summary, cfg.AI.EmbeddingModel, cfg.AI.APIKey)
	}()
	go func() {
		defer wg.Done()
		sessionEntities, extractErr = Extractor{}.Process(summary, cfg.AI.ExtractionModel, cfg.AI.APIKey)
	}()
	wg.Wait()

	if embedErr != nil {
		return fmt.Errorf("session embed: %w", embedErr)
	}
	if extractErr != nil {
		return fmt.Errorf("session extract: %w", extractErr)
	}

	return upsertSession(sessionID, summary, sessionEntities, toBytes(sessionVec))
}
