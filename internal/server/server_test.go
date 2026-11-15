package server_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/alexandrmotologa/evalarena/internal/config"
	"github.com/alexandrmotologa/evalarena/internal/dataset"
	"github.com/alexandrmotologa/evalarena/internal/provider"
	"github.com/alexandrmotologa/evalarena/internal/server"
	"github.com/alexandrmotologa/evalarena/internal/tournament"
)

func TestServer_HealthAndEndpoints(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "server_test.db")

	store, err := tournament.NewStore(dbPath)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer store.Close()

	cfg := &config.Config{
		Tournament: config.TournamentConfig{
			Name:     "API Test Tournament",
			Schedule: "round-robin",
			Database: dbPath,
		},
		Judge: config.JudgeConfig{
			Provider: "mock",
			Model:    "mock-judge",
		},
		Models: []config.ModelConfig{
			{ID: "m1", Provider: "mock", Model: "gpt-4o"},
			{ID: "m2", Provider: "mock", Model: "claude-3-5-sonnet"},
		},
	}

	items := []dataset.Item{
		{ID: "p1", Category: "support", Prompt: "Test prompt"},
	}

	pool, err := provider.NewPool(cfg.Models, cfg.Judge)
	if err != nil {
		t.Fatalf("failed to create pool: %v", err)
	}

	bus := tournament.NewEventBus()

	// Seed one match so arena endpoints have data
	match := tournament.MatchRecord{
		ID:           "test_match_1",
		TournamentID: "API Test Tournament",
		PromptID:     "p1",
		Category:     "support",
		PromptText:   "Test prompt",
		ModelAID:     "m1",
		ModelBID:     "m2",
		ResponseA:    "Answer 1",
		ResponseB:    "Answer 2",
		FinalWinner:  "m1",
		HumanVote:    "model_a",
		CreatedAt:    time.Now(),
	}
	if err := store.SaveMatch(&match); err != nil {
		t.Fatalf("failed to seed match: %v", err)
	}

	srv := server.NewServer(cfg, store, bus, pool, items)
	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()

	// Test GET /api/health
	res, err := http.Get(ts.URL + "/api/health")
	if err != nil {
		t.Fatalf("health check failed: %v", err)
	}
	if res.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", res.StatusCode)
	}

	// Test GET /api/categories
	catRes, err := http.Get(ts.URL + "/api/categories")
	if err != nil {
		t.Fatalf("categories request failed: %v", err)
	}
	if catRes.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", catRes.StatusCode)
	}

	// Test GET /api/calibration
	calRes, err := http.Get(ts.URL + "/api/calibration")
	if err != nil {
		t.Fatalf("calibration request failed: %v", err)
	}
	if calRes.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", calRes.StatusCode)
	}

	// Test POST /api/battle/quick
	quickBody := []byte(`{"prompt":"What is caching?","model_a_id":"m1","model_b_id":"m2"}`)
	quickRes, err := http.Post(ts.URL+"/api/battle/quick", "application/json", bytes.NewReader(quickBody))
	if err != nil {
		t.Fatalf("quick battle request failed: %v", err)
	}
	if quickRes.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", quickRes.StatusCode)
	}
}
