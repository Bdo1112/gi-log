package main

import "sort"

type KeywordPipeline struct{}

func (p KeywordPipeline) Search(ctx SearchContext, exchanges []Exchange, topK int) []SearchResult {
	// filter exchanges whose entities overlap with query keywords
	var filtered []Exchange
	for _, e := range exchanges {
		if entitiesOverlap(ctx.Keywords, e.Entities) {
			filtered = append(filtered, e)
		}
	}

	if len(filtered) == 0 {
		return nil
	}

	// re-rank filtered set by cosine similarity
	var results []SearchResult
	for _, e := range filtered {
		vec := fromBytes(e.Embedding)
		sim := cosineSimilarity(ctx.QueryVec, vec)
		results = append(results, SearchResult{
			ID:           e.ID,
			SessionID:    e.SessionID,
			UserMsg:      e.UserMsg,
			AssistantMsg: e.AssistantMsg,
			Similarity:   sim,
			CreatedAt:    e.CreatedAt,
		})
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Similarity > results[j].Similarity
	})

	if len(results) > topK {
		results = results[:topK]
	}
	return results
}
