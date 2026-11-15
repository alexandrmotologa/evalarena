package report

import (
	"fmt"
	"io"
	"os"

	"github.com/alexandrmotologa/evalarena/internal/elo"
	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
)

// PrintTerminalLeaderboard prints an ASCII table with colored ELO rankings and Pareto indicators.
func PrintTerminalLeaderboard(w io.Writer, tournamentName string, ratings []elo.ModelRating, recs []elo.Recommendation) {
	if w == nil {
		w = os.Stdout
	}

	headerColor := color.New(color.FgCyan, color.Bold)
	title := fmt.Sprintf("\n=== evalarena Leaderboard: %s ===\n", tournamentName)
	headerColor.Fprint(w, title)

	table := tablewriter.NewWriter(w)
	table.Header("Rank", "Model", "ELO", "95% CI", "W/L/T", "Win%", "$/1M Tok", "Latency", "Pareto")

	green := color.New(color.FgGreen).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()
	white := color.New(color.FgWhite).SprintFunc()

	for i, r := range ratings {
		paretoTag := "No"
		if r.IsParetoOptimal {
			paretoTag = green("Yes (Optimal)")
		}

		ciStr := fmt.Sprintf("[%.0f - %.0f]", r.CILower, r.CIUpper)
		wltStr := fmt.Sprintf("%d/%d/%d", r.Wins, r.Losses, r.Ties)
		winPctStr := fmt.Sprintf("%.1f%%", r.WinRate*100.0)
		costStr := fmt.Sprintf("$%.2f", r.AvgCostPerM)
		latencyStr := fmt.Sprintf("%.0fms", r.AvgLatencyMs)

		rankStr := fmt.Sprintf("%d", i+1)
		if i == 0 {
			rankStr = yellow("★ 1")
		}

		_ = table.Append(
			rankStr,
			white(r.ModelID),
			fmt.Sprintf("%.0f", r.ELO),
			ciStr,
			wltStr,
			winPctStr,
			costStr,
			latencyStr,
			paretoTag,
		)
	}

	_ = table.Render()

	// Print Pareto recommendations
	if len(recs) > 0 {
		fmt.Fprintln(w, "\n--- Pareto Frontier Analysis & Recommendations ---")
		for _, rec := range recs {
			fmt.Fprintf(w, "• %s: %s\n", rec.ModelID, rec.Summary)
		}
		fmt.Fprintln(w)
	}
}
