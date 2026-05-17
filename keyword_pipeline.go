package main

import "sort"

type KeywordPipeline struct{}

func (p KeywordPipeline) Search(ctx SearchContext, sessions []Session, exchanges []Exchange, topK int) []SearchResult {
	// find sessions whose entities match query keywords
	matched := map[string]bool{}
	for _, s := range sessions {
		if entitiesOverlap(ctx.Keywords, s.Entities) {
			matched[s.ID] = true
		}
	}

	if len(matched) == 0 {
		return nil
	}

	// collect exchanges from matching sessions, re-rank by cosine similarity
	var results []SearchResult
	for _, e := range exchanges {
		if !matched[e.SessionID] {
			continue
		}
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
