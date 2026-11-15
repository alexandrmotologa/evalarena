package elo_test

import (
	"testing"

	"github.com/alexandrmotologa/evalarena/internal/elo"
)

func TestComputeJudgeCalibration_PerfectAgreement(t *testing.T) {
	pairs := []elo.ComparisonPair{
		{HumanVote: "model_a", JudgeWinner: "m1", ModelAID: "m1", ModelBID: "m2"},
		{HumanVote: "model_b", JudgeWinner: "m2", ModelAID: "m1", ModelBID: "m2"},
		{HumanVote: "tie", JudgeWinner: "tie", ModelAID: "m1", ModelBID: "m2"},
	}

	report := elo.ComputeJudgeCalibration(pairs)

	if report.TotalComparisons != 3 {
		t.Errorf("expected 3 comparisons, got %d", report.TotalComparisons)
	}
	if report.ObservedAgreementPct != 100.0 {
		t.Errorf("expected 100%% agreement, got %f", report.ObservedAgreementPct)
	}
	if report.CohensKappa < 0.99 {
		t.Errorf("expected kappa close to 1.0, got %f", report.CohensKappa)
	}
	if report.Interpretation != "Almost Perfect Agreement" {
		t.Errorf("expected Almost Perfect Agreement, got %s", report.Interpretation)
	}
}

func TestComputeJudgeCalibration_ModerateAgreement(t *testing.T) {
	pairs := []elo.ComparisonPair{
		{HumanVote: "model_a", JudgeWinner: "m1", ModelAID: "m1", ModelBID: "m2"},
		{HumanVote: "model_a", JudgeWinner: "m1", ModelAID: "m1", ModelBID: "m2"},
		{HumanVote: "model_b", JudgeWinner: "m2", ModelAID: "m1", ModelBID: "m2"},
		{HumanVote: "tie", JudgeWinner: "m1", ModelAID: "m1", ModelBID: "m2"}, // disagreement
	}

	report := elo.ComputeJudgeCalibration(pairs)

	if report.TotalComparisons != 4 {
		t.Errorf("expected 4 comparisons, got %d", report.TotalComparisons)
	}
	if report.Agreements != 3 {
		t.Errorf("expected 3 agreements, got %d", report.Agreements)
	}
	if report.CohensKappa <= 0 {
		t.Errorf("expected positive kappa, got %f", report.CohensKappa)
	}
}
