package elo

import (
	"math"
	"sort"
)

// MatchResult represents a single paired match outcome for ELO calculation.
type MatchResult struct {
	ModelA   string
	ModelB   string
	Winner   string // ModelA, ModelB, or "tie"
	LatencyA int64
	LatencyB int64
	CostA    float64
	CostB    float64
}

// ModelRating encapsulates estimated Bradley-Terry ELO and performance metrics.
type ModelRating struct {
	ModelID         string  `json:"model_id"`
	ELO             float64 `json:"elo"`
	CILower         float64 `json:"ci_lower"`
	CIUpper         float64 `json:"ci_upper"`
	Wins            int     `json:"wins"`
	Losses          int     `json:"losses"`
	Ties            int     `json:"ties"`
	TotalBattles    int     `json:"total_battles"`
	WinRate         float64 `json:"win_rate"`
	AvgLatencyMs    float64 `json:"avg_latency_ms"`
	AvgCostPerM     float64 `json:"avg_cost_per_m"`
	IsParetoOptimal bool    `json:"is_pareto_optimal"`
}

// SolveBradleyTerry computes latent skill ratings using Maximum Likelihood Estimation.
func SolveBradleyTerry(models []string, matches []MatchResult, baseELO float64) map[string]float64 {
	K := len(models)
	if K == 0 {
		return nil
	}
	if K == 1 {
		return map[string]float64{models[0]: baseELO}
	}

	modelIdx := make(map[string]int, K)
	for i, m := range models {
		modelIdx[m] = i
	}

	// Pairwise effective win matrix W[i][j] (ties count as 0.5)
	W := make([][]float64, K)
	N := make([][]float64, K)
	for i := range W {
		W[i] = make([]float64, K)
		N[i] = make([]float64, K)
	}

	totalWins := make([]float64, K)

	for _, m := range matches {
		idxA, okA := modelIdx[m.ModelA]
		idxB, okB := modelIdx[m.ModelB]
		if !okA || !okB || idxA == idxB {
			continue
		}

		N[idxA][idxB] += 1.0
		N[idxB][idxA] += 1.0

		if m.Winner == m.ModelA {
			W[idxA][idxB] += 1.0
			totalWins[idxA] += 1.0
		} else if m.Winner == m.ModelB {
			W[idxB][idxA] += 1.0
			totalWins[idxB] += 1.0
		} else {
			// Tie: 0.5 win to each
			W[idxA][idxB] += 0.5
			W[idxB][idxA] += 0.5
			totalWins[idxA] += 0.5
			totalWins[idxB] += 0.5
		}
	}

	// Laplace smoothing / regularization prior to avoid zero division
	prior := 0.01
	for i := 0; i < K; i++ {
		totalWins[i] += prior
		for j := 0; j < K; j++ {
			if i != j {
				N[i][j] += 2 * prior / float64(K-1)
			}
		}
	}

	// gamma_i = exp(beta_i), initialized uniformly
	gamma := make([]float64, K)
	for i := range gamma {
		gamma[i] = 1.0
	}

	// Minorization-Maximization (Hunter 2004) algorithm for global MLE convergence
	maxIter := 200
	tolerance := 1e-7

	for iter := 0; iter < maxIter; iter++ {
		newGamma := make([]float64, K)
		maxDiff := 0.0

		for i := 0; i < K; i++ {
			denom := 0.0
			for j := 0; j < K; j++ {
				if i != j && N[i][j] > 0 {
					denom += N[i][j] / (gamma[i] + gamma[j])
				}
			}

			if denom > 0 {
				newGamma[i] = totalWins[i] / denom
			} else {
				newGamma[i] = gamma[i]
			}

			diff := math.Abs(newGamma[i] - gamma[i])
			if diff > maxDiff {
				maxDiff = diff
			}
		}

		// Geometric mean normalization: prod(gamma) = 1
		logSum := 0.0
		for i := 0; i < K; i++ {
			logSum += math.Log(newGamma[i])
		}
		meanLog := logSum / float64(K)
		for i := 0; i < K; i++ {
			newGamma[i] = math.Exp(math.Log(newGamma[i]) - meanLog)
		}

		gamma = newGamma
		if maxDiff < tolerance {
			break
		}
	}

	// Convert gamma_i to ELO scale: ELO_i = baseELO + (400 / ln(10)) * ln(gamma_i)
	scale := 400.0 / math.Ln10
	eloMap := make(map[string]float64, K)
	for i, m := range models {
		beta := math.Log(gamma[i])
		eloMap[m] = baseELO + scale*beta
	}

	return eloMap
}

// ComputeWinProbability returns P(i beats j) under Bradley-Terry given their ELO scores.
func ComputeWinProbability(eloA, eloB float64) float64 {
	return 1.0 / (1.0 + math.Pow(10.0, (eloB-eloA)/400.0))
}

// SortRatingsByELO sorts rating records descending by ELO score.
func SortRatingsByELO(ratings []ModelRating) {
	sort.Slice(ratings, func(i, j int) bool {
		return ratings[i].ELO > ratings[j].ELO
	})
}
