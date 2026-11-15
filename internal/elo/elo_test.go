package elo_test

import (
	"testing"

	"github.com/alexandrmotologa/evalarena/internal/elo"
)

func TestSolveBradleyTerry_Monotonicity(t *testing.T) {
	models := []string{"strong", "medium", "weak"}
	matches := []elo.MatchResult{
		// strong beats medium 4 times
		{ModelA: "strong", ModelB: "medium", Winner: "strong"},
		{ModelA: "strong", ModelB: "medium", Winner: "strong"},
		{ModelA: "strong", ModelB: "medium", Winner: "strong"},
		{ModelA: "strong", ModelB: "medium", Winner: "strong"},
		// strong beats weak 5 times
		{ModelA: "strong", ModelB: "weak", Winner: "strong"},
		{ModelA: "strong", ModelB: "weak", Winner: "strong"},
		{ModelA: "strong", ModelB: "weak", Winner: "strong"},
		{ModelA: "strong", ModelB: "weak", Winner: "strong"},
		{ModelA: "strong", ModelB: "weak", Winner: "strong"},
		// medium beats weak 3 times
		{ModelA: "medium", ModelB: "weak", Winner: "medium"},
		{ModelA: "medium", ModelB: "weak", Winner: "medium"},
		{ModelA: "medium", ModelB: "weak", Winner: "medium"},
	}

	ratings := elo.SolveBradleyTerry(models, matches, 1000.0)

	if ratings["strong"] <= ratings["medium"] {
		t.Errorf("expected strong ELO (%f) > medium ELO (%f)", ratings["strong"], ratings["medium"])
	}
	if ratings["medium"] <= ratings["weak"] {
		t.Errorf("expected medium ELO (%f) > weak ELO (%f)", ratings["medium"], ratings["weak"])
	}
}

func TestBootstrapCIs(t *testing.T) {
	models := []string{"m1", "m2"}
	matches := []elo.MatchResult{
		{ModelA: "m1", ModelB: "m2", Winner: "m1"},
		{ModelA: "m1", ModelB: "m2", Winner: "m1"},
		{ModelA: "m1", ModelB: "m2", Winner: "m2"},
	}

	cis := elo.ComputeBootstrapCIs(models, matches, 1000.0, 100)
	ci1 := cis["m1"]
	if ci1.Lower > ci1.Upper {
		t.Errorf("invalid CI bounds for m1: lower=%f > upper=%f", ci1.Lower, ci1.Upper)
	}
}

func TestParetoFrontier(t *testing.T) {
	points := []elo.ParetoPoint{
		{ModelID: "flagship", ELO: 1200, CostPerM: 10.0, LatencyMs: 800},
		{ModelID: "balanced", ELO: 1100, CostPerM: 1.0, LatencyMs: 200},
		{ModelID: "budget", ELO: 950, CostPerM: 0.1, LatencyMs: 50},
		{ModelID: "dominated", ELO: 1050, CostPerM: 2.0, LatencyMs: 300}, // dominated by 'balanced' (balanced has higher ELO, lower cost, lower latency)
	}

	frontier := elo.ComputeParetoFrontier(points)

	status := make(map[string]bool)
	for _, p := range frontier {
		status[p.ModelID] = p.IsOptimal
	}

	if !status["flagship"] {
		t.Errorf("expected flagship to be pareto optimal")
	}
	if !status["balanced"] {
		t.Errorf("expected balanced to be pareto optimal")
	}
	if !status["budget"] {
		t.Errorf("expected budget to be pareto optimal")
	}
	if status["dominated"] {
		t.Errorf("expected dominated model to NOT be pareto optimal")
	}
}

func TestComputeWinRateMatrix(t *testing.T) {
	models := []string{"A", "B"}
	matches := []elo.MatchResult{
		{ModelA: "A", ModelB: "B", Winner: "A"},
		{ModelA: "A", ModelB: "B", Winner: "tie"},
	}

	matrix := elo.ComputeWinRateMatrix(models, matches)
	recA := matrix.Data["A"]["B"]

	if recA.Wins != 1 || recA.Ties != 1 || recA.Total != 2 {
		t.Errorf("unexpected record for A vs B: %+v", recA)
	}
	if recA.WinRate != 0.75 { // (1 + 0.5*1)/2 = 0.75
		t.Errorf("expected win rate 0.75, got %f", recA.WinRate)
	}
}
