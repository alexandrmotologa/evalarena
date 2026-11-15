package elo

import (
	"strings"
)

// CalibrationReport contains inter-rater agreement statistics between human votes and LLM judge verdicts.
type CalibrationReport struct {
	TotalComparisons     int     `json:"total_comparisons"`
	Agreements           int     `json:"agreements"`
	ObservedAgreementPct float64 `json:"observed_agreement_pct"`
	ExpectedAgreementPct float64 `json:"expected_agreement_pct"`
	CohensKappa          float64 `json:"cohens_kappa"`
	Interpretation       string  `json:"interpretation"`
}

// ComparisonPair holds one pairwise human vote and one LLM judge verdict.
type ComparisonPair struct {
	HumanVote   string // "model_a", "model_b", "tie", "both_bad"
	JudgeWinner string // modelA ID, modelB ID, or "tie"
	ModelAID    string
	ModelBID    string
}

// ComputeJudgeCalibration calculates observed agreement and Cohen's Kappa (κ).
func ComputeJudgeCalibration(pairs []ComparisonPair) CalibrationReport {
	validCount := 0
	agreedCount := 0

	// 3x3 confusion matrix: 0: A, 1: B, 2: Tie
	var C [3][3]int

	for _, p := range pairs {
		if strings.TrimSpace(p.HumanVote) == "" {
			continue
		}

		hClass := mapHumanVoteToClass(p.HumanVote)
		jClass := mapJudgeVerdictToClass(p.JudgeWinner, p.ModelAID, p.ModelBID)

		if hClass < 0 || jClass < 0 {
			continue
		}

		C[hClass][jClass]++
		validCount++
		if hClass == jClass {
			agreedCount++
		}
	}

	if validCount == 0 {
		return CalibrationReport{
			TotalComparisons:     0,
			Agreements:           0,
			ObservedAgreementPct: 0.0,
			ExpectedAgreementPct: 0.0,
			CohensKappa:          0.0,
			Interpretation:       "No human votes recorded yet",
		}
	}

	// Observed agreement P_o
	Po := float64(agreedCount) / float64(validCount)

	// Expected agreement by chance P_e
	Pe := 0.0
	for k := 0; k < 3; k++ {
		rowSum := 0
		colSum := 0
		for j := 0; j < 3; j++ {
			rowSum += C[k][j]
			colSum += C[j][k]
		}
		pRow := float64(rowSum) / float64(validCount)
		pCol := float64(colSum) / float64(validCount)
		Pe += pRow * pCol
	}

	kappa := 0.0
	if 1.0-Pe > 1e-9 {
		kappa = (Po - Pe) / (1.0 - Pe)
	} else if Po == 1.0 {
		kappa = 1.0
	}

	interpretation := interpretKappa(kappa)

	return CalibrationReport{
		TotalComparisons:     validCount,
		Agreements:           agreedCount,
		ObservedAgreementPct: Po * 100.0,
		ExpectedAgreementPct: Pe * 100.0,
		CohensKappa:          kappa,
		Interpretation:       interpretation,
	}
}

func mapHumanVoteToClass(vote string) int {
	v := strings.ToLower(strings.TrimSpace(vote))
	switch v {
	case "model_a", "a":
		return 0
	case "model_b", "b":
		return 1
	case "tie", "both_bad", "equal":
		return 2
	default:
		return -1
	}
}

func mapJudgeVerdictToClass(verdict, modelAID, modelBID string) int {
	v := strings.ToLower(strings.TrimSpace(verdict))
	if v == strings.ToLower(modelAID) || v == "a" || v == "model_a" {
		return 0
	}
	if v == strings.ToLower(modelBID) || v == "b" || v == "model_b" {
		return 1
	}
	if v == "tie" || v == "equal" || v == "both_bad" {
		return 2
	}
	return -1
}

func interpretKappa(k float64) string {
	if k < 0.0 {
		return "Poor (Disagreement)"
	} else if k <= 0.20 {
		return "Slight Agreement"
	} else if k <= 0.40 {
		return "Fair Agreement"
	} else if k <= 0.60 {
		return "Moderate Agreement"
	} else if k <= 0.80 {
		return "Substantial Agreement"
	} else {
		return "Almost Perfect Agreement"
	}
}
