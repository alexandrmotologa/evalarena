package elo

// HeadToHeadRecord tracks pairwise battle outcomes between two specific models.
type HeadToHeadRecord struct {
	Wins    int     `json:"wins"`
	Losses  int     `json:"losses"`
	Ties    int     `json:"ties"`
	Total   int     `json:"total"`
	WinRate float64 `json:"win_rate"` // (wins + 0.5*ties) / total
}

// WinRateMatrix encapsulates the full pairwise head-to-head table.
type WinRateMatrix struct {
	Models []string                               `json:"models"`
	Data   map[string]map[string]HeadToHeadRecord `json:"data"`
}

// ComputeWinRateMatrix builds the N x N head-to-head grid from tournament matches.
func ComputeWinRateMatrix(models []string, matches []MatchResult) *WinRateMatrix {
	matrix := &WinRateMatrix{
		Models: models,
		Data:   make(map[string]map[string]HeadToHeadRecord, len(models)),
	}

	for _, m := range models {
		matrix.Data[m] = make(map[string]HeadToHeadRecord, len(models))
		for _, opponent := range models {
			matrix.Data[m][opponent] = HeadToHeadRecord{}
		}
	}

	for _, match := range matches {
		mA := match.ModelA
		mB := match.ModelB

		recA := matrix.Data[mA][mB]
		recB := matrix.Data[mB][mA]

		recA.Total++
		recB.Total++

		if match.Winner == mA {
			recA.Wins++
			recB.Losses++
		} else if match.Winner == mB {
			recA.Losses++
			recB.Wins++
		} else {
			recA.Ties++
			recB.Ties++
		}

		if recA.Total > 0 {
			recA.WinRate = (float64(recA.Wins) + 0.5*float64(recA.Ties)) / float64(recA.Total)
		}
		if recB.Total > 0 {
			recB.WinRate = (float64(recB.Wins) + 0.5*float64(recB.Ties)) / float64(recB.Total)
		}

		matrix.Data[mA][mB] = recA
		matrix.Data[mB][mA] = recB
	}

	return matrix
}
