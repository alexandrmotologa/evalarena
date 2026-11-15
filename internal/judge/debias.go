package judge

import (
	"context"
	"fmt"

	"github.com/alexandrmotologa/evalarena/internal/provider"
)

// DualPassResult contains the full outcomes of symmetric forward and swapped passes.
type DualPassResult struct {
	Pass1Winner          string `json:"pass1_winner"`
	Pass1Reasoning       string `json:"pass1_reasoning"`
	Pass2Winner          string `json:"pass2_winner"`
	Pass2Reasoning       string `json:"pass2_reasoning"`
	FinalWinner          string `json:"final_winner"` // ModelAID, ModelBID, or "tie"
	PositionBiasDetected bool   `json:"position_bias_detected"`
	Inconsistent         bool   `json:"inconsistent"`
}

// EvaluateDebiased executes a dual-pass evaluation to eliminate position bias.
func EvaluateDebiased(
	ctx context.Context,
	client provider.Client,
	modelAID, modelBID string,
	userPrompt string,
	responseA, responseB string,
	rubric string,
	temperature float64,
	debiasEnabled bool,
) (*DualPassResult, error) {
	// Pass 1: Candidate A in Option A, Candidate B in Option B
	v1, err := EvaluateSinglePass(ctx, client, userPrompt, responseA, responseB, rubric, temperature)
	if err != nil {
		return nil, fmt.Errorf("pass 1 failed: %w", err)
	}

	pass1WinnerID := resolveWinnerID(v1.Winner, modelAID, modelBID)

	// If position debiasing is disabled, accept single pass result directly
	if !debiasEnabled {
		return &DualPassResult{
			Pass1Winner:          pass1WinnerID,
			Pass1Reasoning:       v1.Reasoning,
			Pass2Winner:          pass1WinnerID,
			Pass2Reasoning:       "Debiasing pass skipped (disabled in config).",
			FinalWinner:          pass1WinnerID,
			PositionBiasDetected: false,
			Inconsistent:         false,
		}, nil
	}

	// Pass 2: Inverted order - Candidate B in Option A, Candidate A in Option B
	v2, err := EvaluateSinglePass(ctx, client, userPrompt, responseB, responseA, rubric, temperature)
	if err != nil {
		return nil, fmt.Errorf("pass 2 failed: %w", err)
	}

	// In Pass 2, Option A is Model B, and Option B is Model A
	pass2WinnerID := resolveWinnerID(v2.Winner, modelBID, modelAID)

	result := &DualPassResult{
		Pass1Winner:    pass1WinnerID,
		Pass1Reasoning: v1.Reasoning,
		Pass2Winner:    pass2WinnerID,
		Pass2Reasoning: v2.Reasoning,
	}

	// Evaluate consistency between Pass 1 and Pass 2
	if pass1WinnerID == pass2WinnerID {
		// Both passes agree
		result.FinalWinner = pass1WinnerID
		result.PositionBiasDetected = false
		result.Inconsistent = false
		return result, nil
	}

	// Check if the judge systematically favored whichever option was in position A (first slot)
	// In Pass 1, slot A was modelAID. In Pass 2, slot A was modelBID.
	if v1.Winner == "A" && v2.Winner == "A" {
		result.FinalWinner = "tie"
		result.PositionBiasDetected = true
		result.Inconsistent = true
		return result, nil
	}

	// Check if the judge systematically favored slot B (second slot)
	if v1.Winner == "B" && v2.Winner == "B" {
		result.FinalWinner = "tie"
		result.PositionBiasDetected = true
		result.Inconsistent = true
		return result, nil
	}

	// Mixed inconsistency (e.g., win on one pass, tie on the other)
	result.FinalWinner = "tie"
	result.PositionBiasDetected = false
	result.Inconsistent = true
	return result, nil
}

func resolveWinnerID(verdictWinner, optionAID, optionBID string) string {
	switch verdictWinner {
	case "A":
		return optionAID
	case "B":
		return optionBID
	default:
		return "tie"
	}
}
