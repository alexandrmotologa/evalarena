package tournament

import (
	"fmt"
	"math/rand"
	"sort"
	"time"

	"github.com/alexandrmotologa/evalarena/internal/provider"
)

// ScheduledMatch represents an allocated battle pairing ready for execution.
type ScheduledMatch struct {
	ID         string
	RoundNum   int
	PromptID   string
	Category   string
	PromptText string
	Messages   []provider.Message
	ModelAID   string
	ModelBID   string
}

// GenerateRoundRobin creates exhaustive all-vs-all pairings for all prompts.
func GenerateRoundRobin(modelIDs []string, prompts []PromptRef, rounds int) []ScheduledMatch {
	var schedule []ScheduledMatch
	matchIndex := 1

	for r := 1; r <= rounds; r++ {
		for i := 0; i < len(modelIDs); i++ {
			for j := i + 1; j < len(modelIDs); j++ {
				for _, p := range prompts {
					schedule = append(schedule, ScheduledMatch{
						ID:         fmt.Sprintf("m_%06d", matchIndex),
						RoundNum:   r,
						PromptID:   p.ID,
						Category:   p.Category,
						PromptText: p.Text,
						Messages:   p.Messages,
						ModelAID:   modelIDs[i],
						ModelBID:   modelIDs[j],
					})
					matchIndex++
				}
			}
		}
	}

	return schedule
}

// GenerateSwiss creates pairings between models with closest current skill ratings.
func GenerateSwiss(modelIDs []string, prompts []PromptRef, roundNum int, ratings map[string]float64) []ScheduledMatch {
	type modelScore struct {
		id    string
		score float64
	}

	scored := make([]modelScore, len(modelIDs))
	for i, id := range modelIDs {
		score := 1000.0
		if r, ok := ratings[id]; ok {
			score = r
		}
		scored[i] = modelScore{id: id, score: score}
	}

	// Sort descending by score
	sort.Slice(scored, func(i, j int) bool {
		return scored[i].score > scored[j].score
	})

	var schedule []ScheduledMatch
	matchIndex := 1

	for i := 0; i < len(scored)-1; i += 2 {
		mA := scored[i].id
		mB := scored[i+1].id

		for _, p := range prompts {
			schedule = append(schedule, ScheduledMatch{
				ID:         fmt.Sprintf("m_r%d_%04d", roundNum, matchIndex),
				RoundNum:   roundNum,
				PromptID:   p.ID,
				Category:   p.Category,
				PromptText: p.Text,
				Messages:   p.Messages,
				ModelAID:   mA,
				ModelBID:   mB,
			})
			matchIndex++
		}
	}

	return schedule
}

// GenerateRandom creates randomly sampled pairwise matches.
func GenerateRandom(modelIDs []string, prompts []PromptRef, totalMatches int) []ScheduledMatch {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	var schedule []ScheduledMatch

	for i := 1; i <= totalMatches; i++ {
		idxA := r.Intn(len(modelIDs))
		idxB := r.Intn(len(modelIDs))
		for idxB == idxA {
			idxB = r.Intn(len(modelIDs))
		}

		prompt := prompts[r.Intn(len(prompts))]

		schedule = append(schedule, ScheduledMatch{
			ID:         fmt.Sprintf("m_rnd_%04d", i),
			RoundNum:   1,
			PromptID:   prompt.ID,
			Category:   prompt.Category,
			PromptText: prompt.Text,
			Messages:   prompt.Messages,
			ModelAID:   modelIDs[idxA],
			ModelBID:   modelIDs[idxB],
		})
	}

	return schedule
}

// PromptRef provides a lightweight reference to a benchmark prompt.
type PromptRef struct {
	ID       string
	Category string
	Text     string
	Messages []provider.Message
}
