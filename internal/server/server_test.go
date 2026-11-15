package server_test

import (
	"encoding/json"
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
		{ID: "p1", Prompt: "Test prompt"},
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
		PromptText:   "Test prompt",
		ModelAID:     "m1",
		ModelBID:     "m2",
		ResponseA:    "Answer 1",
		ResponseB:    "Answer 2",
		FinalWinner:  "m1",
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

	// Test GET /api/leaderboard
	lbRes, err := http.Get(ts.URL + "/api/leaderboard")
	if err != nil {
		t.Fatalf("leaderboard request failed: %v", err)
	}
	if lbRes.StatusCode != http.StatusOK {
		t.Errorf("expected status 200 from leaderboard, got %d", lbRes.StatusCode)
	}

	// Test GET /api/arena/next
	arenaRes, err := http.Get(ts.URL + "/api/arena/next")
	if err != nil {
		t.Fatalf("arena next failed: %v", err)
	}
	if arenaRes.StatusCode != http.StatusOK {
		t.Errorf("expected status 200 from arena next, got %d", arenaRes.StatusCode)
	}

	var blindView map[string]interface{}
	if err := json.NewDecoder(arenaRes.Body).Decode(&blindView); err != nil {
		t.Fatalf("failed to decode arena next JSON: %v", err)
	}
	if blindView["match_id"] != "test_match_1" {
		t.Errorf("expected match_id test_match_1, got %v", blindView["match_id"])
	}
}
