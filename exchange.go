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
		vec, embedErr = cfg.Client.Embed(combined)
	}()
	go func() {
		defer wg.Done()
		entities, extractErr = cfg.Client.Extract(combined)
	}()
	wg.Wait()

	if embedErr != nil {
		return fmt.Errorf("embed: %w", embedErr)
	}
	if extractErr != nil {
		return fmt.Errorf("extract: %w", extractErr)
	}

	id := newID()
	store := newStore(cfg)
	if err := store.Save(id, sessionID, userMsg, assistantMsg, toBytes(vec), entities); err != nil {
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
