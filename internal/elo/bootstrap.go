package elo

import (
	"math/rand"
	"sort"
	"time"
)

// ConfidenceInterval contains lower and upper bounds of ELO rating.
type ConfidenceInterval struct {
	Lower float64
	Upper float64
}

// ComputeBootstrapCIs calculates 95% confidence intervals via non-parametric bootstrap resampling.
func ComputeBootstrapCIs(models []string, matches []MatchResult, baseELO float64, iterations int) map[string]ConfidenceInterval {
	if iterations <= 0 {
		iterations = 1000
	}

	M := len(matches)
	if M == 0 {
		res := make(map[string]ConfidenceInterval, len(models))
		for _, m := range models {
			res[m] = ConfidenceInterval{Lower: baseELO, Upper: baseELO}
		}
		return res
	}

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	samples := make(map[string][]float64, len(models))
	for _, m := range models {
		samples[m] = make([]float64, 0, iterations)
	}

	resampled := make([]MatchResult, M)

	for b := 0; b < iterations; b++ {
		// Draw M matches with replacement
		for i := 0; i < M; i++ {
			resampled[i] = matches[r.Intn(M)]
		}

		eloMap := SolveBradleyTerry(models, resampled, baseELO)
		for _, m := range models {
			samples[m] = append(samples[m], eloMap[m])
		}
	}

	cis := make(map[string]ConfidenceInterval, len(models))
	lowerIdx := int(float64(iterations) * 0.025)
	upperIdx := int(float64(iterations) * 0.975)
	if upperIdx >= iterations {
		upperIdx = iterations - 1
	}

	for _, m := range models {
		vals := samples[m]
		sort.Float64s(vals)
		cis[m] = ConfidenceInterval{
			Lower: vals[lowerIdx],
			Upper: vals[upperIdx],
		}
	}

	return cis
}
