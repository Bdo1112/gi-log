package main

import (
	"math"
	"sort"
)

type SearchResult struct {
	ID           string
	SessionID    string
	UserMsg      string
	AssistantMsg string
	Similarity   float32
	CreatedAt    string
}

const similarityThreshold = 0.5

func cosineSimilarity(a, b []float32) float32 {
	var dot, normA, normB float32
	for i := range a {
		dot += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}
	if normA == 0 || normB == 0 {
		return 0
	}
	return dot / (float32(math.Sqrt(float64(normA))) * float32(math.Sqrt(float64(normB))))
}

func rankMemories(queryVec []float32, exchanges []Exchange, topK int) []SearchResult {
	var results []SearchResult
	for _, e := range exchanges {
		vec := fromBytes(e.Embedding)
		sim := cosineSimilarity(queryVec, vec)
		if sim >= similarityThreshold {
			results = append(results, SearchResult{
				ID:           e.ID,
				SessionID:    e.SessionID,
				UserMsg:      e.UserMsg,
				AssistantMsg: e.AssistantMsg,
				Similarity:   sim,
				CreatedAt:    e.CreatedAt,
			})
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Similarity > results[j].Similarity
	})

	if len(results) > topK {
		results = results[:topK]
	}
	return results
}
