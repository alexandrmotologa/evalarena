package tournament_test

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/alexandrmotologa/evalarena/internal/config"
	"github.com/alexandrmotologa/evalarena/internal/dataset"
	"github.com/alexandrmotologa/evalarena/internal/provider"
	"github.com/alexandrmotologa/evalarena/internal/tournament"
)

func TestMatchmaker_RoundRobin(t *testing.T) {
	models := []string{"m1", "m2", "m3"}
	prompts := []tournament.PromptRef{
		{ID: "p1", Text: "Prompt 1"},
		{ID: "p2", Text: "Prompt 2"},
	}

	// 3 models -> 3 pairs (m1-m2, m1-m3, m2-m3) * 2 prompts * 1 round = 6 matches
	matches := tournament.GenerateRoundRobin(models, prompts, 1)
	if len(matches) != 6 {
		t.Fatalf("expected 6 matches, got %d", len(matches))
	}
}

func TestRunner_EndToEndSimulation(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "sim_test.db")

	store, err := tournament.NewStore(dbPath)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer store.Close()

	cfg := &config.Config{
		Tournament: config.TournamentConfig{
			Name:        "Sim Test Tournament",
			Schedule:    "round-robin",
			Rounds:      1,
			Concurrency: 2,
			Database:    dbPath,
		},
		Judge: config.JudgeConfig{
			Provider:       "mock",
			Model:          "mock-judge",
			Temperature:    0.1,
			DebiasPosition: true,
		},
		Models: []config.ModelConfig{
			{ID: "m1", Provider: "mock", Model: "gpt-4o", Pricing: config.PricingConfig{InputPerMillion: 2.5, OutputPerMillion: 10.0}},
			{ID: "m2", Provider: "mock", Model: "gemini-2.0-flash", Pricing: config.PricingConfig{InputPerMillion: 0.1, OutputPerMillion: 0.4}},
		},
	}

	items := []dataset.Item{
		{ID: "prompt_1", Prompt: "Explain enterprise database indexing."},
	}

	pool, err := provider.NewPool(cfg.Models, cfg.Judge)
	if err != nil {
		t.Fatalf("failed to create pool: %v", err)
	}

	bus := tournament.NewEventBus()
	runner := tournament.NewRunner(cfg, items, pool, store, bus)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	completed, err := runner.Run(ctx)
	if err != nil {
		t.Fatalf("runner failed: %v", err)
	}

	if len(completed) != 1 {
		t.Fatalf("expected 1 completed match, got %d", len(completed))
	}

	if completed[0].ResponseA == "" || completed[0].ResponseB == "" {
		t.Errorf("empty response generated")
	}
	if completed[0].FinalWinner == "" {
		t.Errorf("expected final winner to be set")
	}
}
