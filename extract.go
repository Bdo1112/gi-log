package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"gi-log/util/prompt"
)

type Extractor struct{}

func (e Extractor) Process(text, model, apiKey string) ([]string, error) {
	body, err := json.Marshal(map[string]any{
		"model": model,
		"messages": []map[string]string{
			{
				"role":    "system",
				"content": prompt.ExtractPrompt,
			},
			{
				"role":    "user",
				"content": text,
			},
		},
	})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
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
		return nil, err
	}

	if result.Error != nil {
		return nil, fmt.Errorf("openai: %s", result.Error.Message)
	}
	if len(result.Choices) == 0 {
		return nil, fmt.Errorf("openai: no choices returned")
	}

	content := strings.TrimSpace(result.Choices[0].Message.Content)
	if strings.HasPrefix(content, "```") {
		if i := strings.Index(content, "\n"); i != -1 {
			content = content[i+1:]
		}
		content = strings.TrimSpace(content)
		content = strings.TrimSuffix(content, "```")
		content = strings.TrimSpace(content)
	}

	// try parsing as plain array first
	var entities []string
	if err := json.Unmarshal([]byte(content), &entities); err != nil {
		// fall back: model may have returned {"entities": [...]}
		var wrapper struct {
			Entities []string `json:"entities"`
		}
		if err2 := json.Unmarshal([]byte(content), &wrapper); err2 != nil {
			return nil, fmt.Errorf("cannot parse entities from: %s", content)
		}
		entities = wrapper.Entities
	}

	return entities, nil
}

func entitiesToJSON(entities []string) string {
	data, _ := json.Marshal(entities)
	return string(data)
}

func entitiesFromJSON(s string) []string {
	if s == "" {
		return nil
	}
	var entities []string
	json.Unmarshal([]byte(s), &entities)
	return entities
}

func entitiesOverlap(queryKeywords, exchangeEntities []string) bool {
	for _, q := range queryKeywords {
		ql := strings.ToLower(q)
		for _, e := range exchangeEntities {
			el := strings.ToLower(e)
			if el == ql || strings.Contains(el, ql) || strings.Contains(ql, el) {
				return true
			}
		}
	}
	return false
}
