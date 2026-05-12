package main

import (
	"fmt"
	"sync"
)

func saveExchange(sessionID, userMsg, assistantMsg string, cfg Config) error {
	combined := "User: " + userMsg + " Assistant: " + assistantMsg

	var vec []float32
	var entities []string
	var embedErr, extractErr error

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		vec, embedErr = Embedder{}.Process(combined, cfg.AI.EmbeddingModel, cfg.AI.APIKey)
	}()
	go func() {
		defer wg.Done()
		entities, extractErr = Extractor{}.Process(combined, cfg.AI.ExtractionModel, cfg.AI.APIKey)
	}()
	wg.Wait()

	if embedErr != nil {
		return fmt.Errorf("embed: %w", embedErr)
	}
	if extractErr != nil {
		return fmt.Errorf("extract: %w", extractErr)
	}

	id := newID()
	if err := insertExchange(id, sessionID, userMsg, assistantMsg, toBytes(vec), entities); err != nil {
		return err
	}

	count, err := countExchangesBySession(sessionID)
	if err != nil {
		return err
	}
	if count%5 == 0 {
		return summarizeSession(sessionID, cfg)
	}
	return nil
}
