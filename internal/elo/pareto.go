package elo

import (
	"fmt"
	"sort"
)

// ParetoPoint represents a candidate configuration in multi-objective space.
type ParetoPoint struct {
	ModelID   string  `json:"model_id"`
	ELO       float64 `json:"elo"`
	CostPerM  float64 `json:"cost_per_m"`
	LatencyMs float64 `json:"latency_ms"`
	IsOptimal bool    `json:"is_optimal"`
}

// ComputeParetoFrontier identifies non-dominated models in (ELO, Cost, Latency) space.
func ComputeParetoFrontier(points []ParetoPoint) []ParetoPoint {
	n := len(points)
	result := make([]ParetoPoint, n)
	copy(result, points)

	for i := 0; i < n; i++ {
		result[i].IsOptimal = true
		for j := 0; j < n; j++ {
			if i == j {
				continue
			}
			// Check if point j strictly dominates point i
			// Dominates if j has >= ELO, <= Cost, <= Latency, with at least one strictly better
			jDominatesI := result[j].ELO >= result[i].ELO &&
				result[j].CostPerM <= result[i].CostPerM &&
				result[j].LatencyMs <= result[i].LatencyMs &&
				(result[j].ELO > result[i].ELO || result[j].CostPerM < result[i].CostPerM || result[j].LatencyMs < result[i].LatencyMs)

			if jDominatesI {
				result[i].IsOptimal = false
				break
			}
		}
	}

	return result
}

// Recommendation highlights high-efficiency models for enterprise deployment.
type Recommendation struct {
	ModelID          string  `json:"model_id"`
	ELORating        float64 `json:"elo_rating"`
	CostPerM         float64 `json:"cost_per_m"`
	LatencyMs        float64 `json:"latency_ms"`
	ELORetentionPct  float64 `json:"elo_retention_pct"` // % of top model ELO
	CostReductionPct float64 `json:"cost_reduction_pct"` // % saved compared to top model
	Summary          string  `json:"summary"`
}

// GenerateRecommendations analyzes Pareto points and produces concrete business choices.
func GenerateRecommendations(points []ParetoPoint) []Recommendation {
	if len(points) == 0 {
		return nil
	}

	sorted := make([]ParetoPoint, len(points))
	copy(sorted, points)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].ELO > sorted[j].ELO
	})

	topModel := sorted[0]
	var recs []Recommendation

	for _, p := range sorted {
		if !p.IsOptimal {
			continue
		}

		eloPct := (p.ELO / topModel.ELO) * 100.0
		costReduction := 0.0
		if topModel.CostPerM > 0 {
			costReduction = ((topModel.CostPerM - p.CostPerM) / topModel.CostPerM) * 100.0
		}

		var summary string
		if p.ModelID == topModel.ModelID {
			summary = fmt.Sprintf("Top Quality Baseline (ELO %.0f, $%.2f/1M tokens, %.0fms latency).", p.ELO, p.CostPerM, p.LatencyMs)
		} else if costReduction > 0 {
			summary = fmt.Sprintf("High-Efficiency Alternative: Delivers %.1f%% of peak ELO while reducing inference cost by %.1f%% (%.0fms latency).", eloPct, costReduction, p.LatencyMs)
		} else {
			summary = fmt.Sprintf("Pareto Optimal: ELO %.0f, $%.2f/1M tokens, %.0fms latency.", p.ELO, p.CostPerM, p.LatencyMs)
		}

		recs = append(recs, Recommendation{
			ModelID:          p.ModelID,
			ELORating:        p.ELO,
			CostPerM:         p.CostPerM,
			LatencyMs:        p.LatencyMs,
			ELORetentionPct:  eloPct,
			CostReductionPct: costReduction,
			Summary:          summary,
		})
	}

	return recs
}
