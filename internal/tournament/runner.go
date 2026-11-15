package tournament

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/alexandrmotologa/evalarena/internal/config"
	"github.com/alexandrmotologa/evalarena/internal/dataset"
	"github.com/alexandrmotologa/evalarena/internal/judge"
	"github.com/alexandrmotologa/evalarena/internal/provider"
)

// Runner orchestrates parallel pairwise tournament execution.
type Runner struct {
	cfg      *config.Config
	dataset  []dataset.Item
	pool     *provider.Pool
	store    *Store
	eventBus *EventBus
}

// NewRunner creates a tournament execution runner.
func NewRunner(cfg *config.Config, items []dataset.Item, pool *provider.Pool, store *Store, bus *EventBus) *Runner {
	if bus == nil {
		bus = NewEventBus()
	}
	return &Runner{
		cfg:      cfg,
		dataset:  items,
		pool:     pool,
		store:    store,
		eventBus: bus,
	}
}

// EventBus returns the attached event broadcaster.
func (r *Runner) EventBus() *EventBus {
	return r.eventBus
}

// Run executes the complete tournament schedule and stores all battle results.
func (r *Runner) Run(ctx context.Context) ([]MatchRecord, error) {
	var modelIDs []string
	for _, m := range r.cfg.Models {
		modelIDs = append(modelIDs, m.ID)
	}

	var prompts []PromptRef
	for _, item := range r.dataset {
		var msgs []provider.Message
		for _, m := range item.Messages {
			msgs = append(msgs, provider.Message{Role: m.Role, Content: m.Content})
		}
		prompts = append(prompts, PromptRef{
			ID:       item.ID,
			Category: item.Category,
			Text:     item.Prompt,
			Messages: msgs,
		})
	}

	var schedule []ScheduledMatch
	switch r.cfg.Tournament.Schedule {
	case "round-robin":
		schedule = GenerateRoundRobin(modelIDs, prompts, r.cfg.Tournament.Rounds)
	case "swiss":
		schedule = GenerateSwiss(modelIDs, prompts, 1, nil)
	default:
		schedule = GenerateRoundRobin(modelIDs, prompts, r.cfg.Tournament.Rounds)
	}

	concurrency := r.cfg.Tournament.Concurrency
	if concurrency <= 0 {
		concurrency = 4
	}

	sem := make(chan struct{}, concurrency)
	var wg sync.WaitGroup
	var mu sync.Mutex

	var completedMatches []MatchRecord
	var runErr error

	for _, sm := range schedule {
		if ctx.Err() != nil {
			break
		}

		wg.Add(1)
		sem <- struct{}{}

		go func(match ScheduledMatch) {
			defer wg.Done()
			defer func() { <-sem }()

			record, err := r.executeMatch(ctx, match)
			if err != nil {
				mu.Lock()
				if runErr == nil {
					runErr = fmt.Errorf("match %s failed: %w", match.ID, err)
				}
				mu.Unlock()
				return
			}

			mu.Lock()
			completedMatches = append(completedMatches, *record)
			mu.Unlock()
		}(sm)
	}

	wg.Wait()

	if runErr != nil {
		return completedMatches, runErr
	}

	return completedMatches, nil
}

func (r *Runner) executeMatch(ctx context.Context, sm ScheduledMatch) (*MatchRecord, error) {
	r.eventBus.Publish("match_start", map[string]interface{}{
		"match_id": sm.ID,
		"model_a":  sm.ModelAID,
		"model_b":  sm.ModelBID,
		"prompt":   sm.PromptText,
	})

	// 1. Dispatch prompt to Model A and Model B in parallel
	req := provider.CompletionRequest{
		Prompt:      sm.PromptText,
		Messages:    sm.Messages,
		Temperature: 0.7,
	}

	respA, respB, err := r.pool.DispatchPair(ctx, sm.ModelAID, sm.ModelBID, req)
	if err != nil {
		return nil, fmt.Errorf("parallel dispatch failed: %w", err)
	}

	// 2. Automated evaluation with double-blind position debiasing
	judgeClient := r.pool.GetJudge()
	judgeResult, err := judge.EvaluateDebiased(
		ctx,
		judgeClient,
		sm.ModelAID,
		sm.ModelBID,
		sm.PromptText,
		respA.Text,
		respB.Text,
		r.cfg.Judge.RubricContent,
		r.cfg.Judge.Temperature,
		r.cfg.Judge.DebiasPosition,
	)
	if err != nil {
		return nil, fmt.Errorf("judge evaluation failed: %w", err)
	}

	// 3. Construct match record
	record := &MatchRecord{
		ID:                   sm.ID,
		TournamentID:         r.cfg.Tournament.Name,
		PromptID:             sm.PromptID,
		Category:             sm.Category,
		PromptText:           sm.PromptText,
		RoundNum:             sm.RoundNum,
		ModelAID:             sm.ModelAID,
		ModelBID:             sm.ModelBID,
		ResponseA:            respA.Text,
		ResponseB:            respB.Text,
		LatencyAMs:           respA.LatencyMs,
		LatencyBMs:           respB.LatencyMs,
		CostA:                respA.TotalCost,
		CostB:                respB.TotalCost,
		TokensA:              respA.InputTokens + respA.OutputTokens,
		TokensB:              respB.InputTokens + respB.OutputTokens,
		Pass1Winner:          judgeResult.Pass1Winner,
		Pass1Reasoning:       judgeResult.Pass1Reasoning,
		Pass2Winner:          judgeResult.Pass2Winner,
		Pass2Reasoning:       judgeResult.Pass2Reasoning,
		FinalWinner:          judgeResult.FinalWinner,
		PositionBiasDetected: judgeResult.PositionBiasDetected,
		CreatedAt:            time.Now(),
	}

	// 4. Persist to SQLite
	if err := r.store.SaveMatch(record); err != nil {
		return nil, fmt.Errorf("failed to save match record: %w", err)
	}

	r.eventBus.Publish("match_finish", record)

	return record, nil
}
