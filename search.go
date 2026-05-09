package main

import "math"

type SearchContext struct {
	RawQuery string
	QueryVec []float32
	Keywords []string
}

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
