package report

import (
	"fmt"
	"strings"
	"time"

	"github.com/alexandrmotologa/evalarena/internal/elo"
)

// GenerateMarkdownReport produces an executive summary formatted for pull requests and documentation.
func GenerateMarkdownReport(tournamentName string, ratings []elo.ModelRating, matrix *elo.WinRateMatrix, recs []elo.Recommendation) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("# evalarena Benchmark Report: %s\n\n", tournamentName))
	sb.WriteString(fmt.Sprintf("*Generated at: %s*\n\n", time.Now().UTC().Format(time.RFC1123)))

	sb.WriteString("## 1. Bradley-Terry ELO Leaderboard\n\n")
	sb.WriteString("| Rank | Model | ELO Rating | 95% Bootstrap CI | W / L / T | Win Rate | Blended Cost / 1M | Latency | Pareto Status |\n")
	sb.WriteString("|:----:|:------|:----------:|:----------------:|:---------:|:--------:|:-----------------:|:-------:|:-------------:|\n")

	for i, r := range ratings {
		paretoTag := "Dominated"
		if r.IsParetoOptimal {
			paretoTag = "**Pareto Optimal**"
		}
		sb.WriteString(fmt.Sprintf("| %d | `%s` | %.1f | [%.1f - %.1f] | %d/%d/%d | %.1f%% | $%.2f | %.0fms | %s |\n",
			i+1, r.ModelID, r.ELO, r.CILower, r.CIUpper, r.Wins, r.Losses, r.Ties, r.WinRate*100.0, r.AvgCostPerM, r.AvgLatencyMs, paretoTag))
	}
	sb.WriteString("\n")

	if len(recs) > 0 {
		sb.WriteString("## 2. Pareto Efficiency Recommendations\n\n")
		for _, rec := range recs {
			sb.WriteString(fmt.Sprintf("- **`%s`**: %s\n", rec.ModelID, rec.Summary))
		}
		sb.WriteString("\n")
	}

	if matrix != nil && len(matrix.Models) > 1 {
		sb.WriteString("## 3. Pairwise Win-Rate Matrix\n\n")
		sb.WriteString("| Model vs Opponent |")
		for _, m := range matrix.Models {
			sb.WriteString(fmt.Sprintf(" `%s` |", m))
		}
		sb.WriteString("\n|:---|")
		for range matrix.Models {
			sb.WriteString(":---:|")
		}
		sb.WriteString("\n")

		for _, mA := range matrix.Models {
			sb.WriteString(fmt.Sprintf("| `%s` |", mA))
			for _, mB := range matrix.Models {
				if mA == mB {
					sb.WriteString(" - |")
				} else {
					rec := matrix.Data[mA][mB]
					sb.WriteString(fmt.Sprintf(" %.1f%% (%d-%d-%d) |", rec.WinRate*100.0, rec.Wins, rec.Losses, rec.Ties))
				}
			}
			sb.WriteString("\n")
		}
		sb.WriteString("\n")
	}

	sb.WriteString("## 4. Evaluation Methodology\n\n")
	sb.WriteString("- **Position-Bias Mitigation**: Every comparison was evaluated in both direct ($A/B$) and swapped ($B/A$) presentation order to eliminate first-option preference bias.\n")
	sb.WriteString("- **Rating Solver**: Latent skills calculated via Minorization-Maximization Bradley-Terry Maximum Likelihood Estimation.\n")
	sb.WriteString("- **Statistical Rigor**: 95% confidence intervals generated via 1,000 non-parametric bootstrap resampling iterations.\n")

	return sb.String()
}
