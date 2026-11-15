package report

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"

	"github.com/alexandrmotologa/evalarena/internal/elo"
)

// ExportJSON writes ratings and Pareto recommendations as structured JSON.
func ExportJSON(w io.Writer, ratings []elo.ModelRating, recs []elo.Recommendation) error {
	data := map[string]interface{}{
		"ratings":         ratings,
		"recommendations": recs,
	}

	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

// ExportCSV writes rating metrics as a flat CSV table for spreadsheet inspection.
func ExportCSV(w io.Writer, ratings []elo.ModelRating) error {
	writer := csv.NewWriter(w)
	defer writer.Flush()

	header := []string{
		"rank", "model_id", "elo", "ci_lower", "ci_upper",
		"wins", "losses", "ties", "win_rate", "avg_latency_ms", "avg_cost_per_m", "is_pareto_optimal",
	}
	if err := writer.Write(header); err != nil {
		return err
	}

	for i, r := range ratings {
		row := []string{
			fmt.Sprintf("%d", i+1),
			r.ModelID,
			fmt.Sprintf("%.2f", r.ELO),
			fmt.Sprintf("%.2f", r.CILower),
			fmt.Sprintf("%.2f", r.CIUpper),
			fmt.Sprintf("%d", r.Wins),
			fmt.Sprintf("%d", r.Losses),
			fmt.Sprintf("%d", r.Ties),
			fmt.Sprintf("%.4f", r.WinRate),
			fmt.Sprintf("%.1f", r.AvgLatencyMs),
			fmt.Sprintf("%.4f", r.AvgCostPerM),
			fmt.Sprintf("%t", r.IsParetoOptimal),
		}
		if err := writer.Write(row); err != nil {
			return err
		}
	}

	return nil
}
