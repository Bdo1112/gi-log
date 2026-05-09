package main

import "sort"

type SemanticPipeline struct{}

func (p SemanticPipeline) Search(ctx SearchContext, exchanges []Exchange, topK int) []SearchResult {
	var results []SearchResult
	for _, e := range exchanges {
		vec := fromBytes(e.Embedding)
		sim := cosineSimilarity(ctx.QueryVec, vec)
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
