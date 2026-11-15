package tournament_test

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/alexandrmotologa/evalarena/internal/tournament"
)

func TestStore_MatchesAndLeaderboard(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test_evalarena.db")

	store, err := tournament.NewStore(dbPath)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer store.Close()

	match := tournament.MatchRecord{
		ID:                   "m_001",
		TournamentID:         "tourney_1",
		PromptID:             "p_001",
		Category:             "support_policy",
		PromptText:           "Test prompt",
		RoundNum:             1,
		ModelAID:             "gpt-4o",
		ModelBID:             "claude-3-5-sonnet",
		ResponseA:            "Answer A",
		ResponseB:            "Answer B",
		LatencyAMs:           450,
		LatencyBMs:           520,
		CostA:                0.002,
		CostB:                0.003,
		TokensA:              150,
		TokensB:              160,
		Pass1Winner:          "gpt-4o",
		Pass1Reasoning:       "A was clearer",
		Pass2Winner:          "gpt-4o",
		Pass2Reasoning:       "A remained clearer",
		FinalWinner:          "gpt-4o",
		PositionBiasDetected: false,
		CreatedAt:            time.Now(),
	}

	if err := store.SaveMatch(&match); err != nil {
		t.Fatalf("failed to save match: %v", err)
	}

	matches, err := store.GetMatches("tourney_1", "")
	if err != nil {
		t.Fatalf("failed to get matches: %v", err)
	}
	if len(matches) != 1 {
		t.Fatalf("expected 1 match, got %d", len(matches))
	}
	if matches[0].FinalWinner != "gpt-4o" {
		t.Errorf("expected final winner gpt-4o, got %s", matches[0].FinalWinner)
	}

	cats, err := store.GetCategories("tourney_1")
	if err != nil {
		t.Fatalf("failed to get categories: %v", err)
	}
	if len(cats) != 1 || cats[0] != "support_policy" {
		t.Errorf("unexpected categories: %+v", cats)
	}

	// Human voting
	unvoted, err := store.GetUnvotedMatch("tourney_1")
	if err != nil {
		t.Fatalf("failed to get unvoted match: %v", err)
	}
	if unvoted.ID != "m_001" {
		t.Errorf("expected match ID m_001, got %s", unvoted.ID)
	}

	updated, err := store.RecordHumanVote("m_001", "model_b")
	if err != nil {
		t.Fatalf("failed to record human vote: %v", err)
	}
	if updated.HumanVote != "model_b" {
		t.Errorf("expected human vote model_b, got %s", updated.HumanVote)
	}

	// Leaderboard test
	lbRecords := []tournament.LeaderboardRecord{
		{
			ModelID:         "gpt-4o",
			ELO:             1050.0,
			CILower:         1020.0,
			CIUpper:         1080.0,
			Wins:            1,
			Losses:          0,
			Ties:            0,
			WinRate:         1.0,
			AvgLatencyMs:    450.0,
			AvgCostPerM:     2.5,
			IsParetoOptimal: true,
		},
	}

	if err := store.SaveLeaderboard("tourney_1", lbRecords); err != nil {
		t.Fatalf("failed to save leaderboard: %v", err)
	}

	loadedLB, err := store.GetLeaderboard("tourney_1")
	if err != nil {
		t.Fatalf("failed to get leaderboard: %v", err)
	}
	if len(loadedLB) != 1 {
		t.Fatalf("expected 1 leaderboard entry, got %d", len(loadedLB))
	}
	if loadedLB[0].ModelID != "gpt-4o" || loadedLB[0].ELO != 1050.0 {
		t.Errorf("unexpected leaderboard record: %+v", loadedLB[0])
	}
}
