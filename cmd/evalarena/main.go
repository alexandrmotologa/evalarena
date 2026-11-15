package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/alexandrmotologa/evalarena/internal/config"
	"github.com/alexandrmotologa/evalarena/internal/dataset"
	"github.com/alexandrmotologa/evalarena/internal/elo"
	"github.com/alexandrmotologa/evalarena/internal/provider"
	"github.com/alexandrmotologa/evalarena/internal/report"
	"github.com/alexandrmotologa/evalarena/internal/server"
	"github.com/alexandrmotologa/evalarena/internal/tournament"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "evalarena",
		Short: "evalarena — Self-Hosted Blind Pairwise Model Preference Tournament Engine",
		Long: `evalarena is an enterprise platform for blind pairwise LLM evaluation,
featuring position-bias mitigated LLM judges, Bradley-Terry ELO rating estimation
with 95% bootstrap confidence intervals, and Pareto efficiency frontier analysis.`,
	}

	rootCmd.AddCommand(newRunCmd())
	rootCmd.AddCommand(newArenaCmd())
	rootCmd.AddCommand(newReportCmd())
	rootCmd.AddCommand(newCheckCmd())
	rootCmd.AddCommand(newInitCmd())

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func newRunCmd() *cobra.Command {
	var cfgPath string
	var rounds int

	cmd := &cobra.Command{
		Use:   "run",
		Short: "Execute a pairwise benchmark tournament across configured models",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load(cfgPath)
			if err != nil {
				return err
			}

			if rounds > 0 {
				cfg.Tournament.Rounds = rounds
			}

			items, err := dataset.LoadDataset(cfg.Dataset.Path, cfg.Dataset.Format)
			if err != nil {
				return fmt.Errorf("failed to load dataset: %w", err)
			}

			store, err := tournament.NewStore(cfg.Tournament.Database)
			if err != nil {
				return fmt.Errorf("failed to open database: %w", err)
			}
			defer store.Close()

			pool, err := provider.NewPool(cfg.Models, cfg.Judge)
			if err != nil {
				return fmt.Errorf("failed to initialize provider pool: %w", err)
			}

			bus := tournament.NewEventBus()
			runner := tournament.NewRunner(cfg, items, pool, store, bus)

			fmt.Printf("Starting tournament %q with %d models on %d prompts (%d rounds)...\n",
				cfg.Tournament.Name, len(cfg.Models), len(items), cfg.Tournament.Rounds)

			ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
			defer cancel()

			start := time.Now()
			matches, err := runner.Run(ctx)
			if err != nil {
				return fmt.Errorf("tournament failed: %w", err)
			}

			fmt.Printf("Completed %d pairwise battles in %s.\n", len(matches), time.Since(start).Round(time.Millisecond))

			// Compute final ratings and display terminal summary
			ratings, matrix, recs, err := calculateTournamentResults(cfg, store, "")
			if err != nil {
				return err
			}

			report.PrintTerminalLeaderboard(os.Stdout, cfg.Tournament.Name, ratings, recs)
			_ = matrix

			return nil
		},
	}

	cmd.Flags().StringVarP(&cfgPath, "config", "c", "examples/evalarena.yaml", "Path to YAML configuration file")
	cmd.Flags().IntVarP(&rounds, "rounds", "r", 0, "Override number of tournament rounds")

	return cmd
}

func newArenaCmd() *cobra.Command {
	var cfgPath string
	var port int

	cmd := &cobra.Command{
		Use:   "arena",
		Short: "Launch the embedded React 19 Arena Studio for human voting and Pareto exploration",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load(cfgPath)
			if err != nil {
				return err
			}

			items, err := dataset.LoadDataset(cfg.Dataset.Path, cfg.Dataset.Format)
			if err != nil {
				return fmt.Errorf("failed to load dataset: %w", err)
			}

			store, err := tournament.NewStore(cfg.Tournament.Database)
			if err != nil {
				return fmt.Errorf("failed to open database: %w", err)
			}
			defer store.Close()

			pool, err := provider.NewPool(cfg.Models, cfg.Judge)
			if err != nil {
				return fmt.Errorf("failed to initialize provider pool: %w", err)
			}

			bus := tournament.NewEventBus()
			apiServer := server.NewServer(cfg, store, bus, pool, items)

			// Combine API router and embedded static asset file server
			mux := http.NewServeMux()
			mux.Handle("/api/", apiServer.Handler())
			mux.Handle("/", server.FileServer())

			addr := fmt.Sprintf(":%d", port)
			httpServer := &http.Server{
				Addr:         addr,
				Handler:      mux,
				ReadTimeout:  30 * time.Second,
				WriteTimeout: 60 * time.Second,
			}

			color.Cyan("\n╔════════════════════════════════════════════════════════════╗")
			color.Cyan("║            evalarena Studio — Embedded Server              ║")
			color.Cyan("╚════════════════════════════════════════════════════════════╝")
			fmt.Printf("• Web Studio:       http://localhost:%d\n", port)
			fmt.Printf("• REST API:         http://localhost:%d/api/leaderboard\n", port)
			fmt.Printf("• Database:         %s\n", cfg.Tournament.Database)
			fmt.Printf("• Active Models:    %d\n\n", len(cfg.Models))
			color.Green("Ready for blind voting, live SSE streaming, and Pareto exploration. Press Ctrl+C to stop.\n")

			go func() {
				if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
				}
			}()

			stop := make(chan os.Signal, 1)
			signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
			<-stop

			color.Yellow("\nShutting down server gracefully...")
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			return httpServer.Shutdown(ctx)
		},
	}

	cmd.Flags().StringVarP(&cfgPath, "config", "c", "examples/evalarena.yaml", "Path to YAML configuration file")
	cmd.Flags().IntVarP(&port, "port", "p", 7070, "HTTP server port")

	return cmd
}

func newReportCmd() *cobra.Command {
	var cfgPath string
	var format string
	var outputPath string
	var category string

	cmd := &cobra.Command{
		Use:   "report",
		Short: "Generate terminal, markdown, JSON, or CSV benchmark reports",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load(cfgPath)
			if err != nil {
				return err
			}

			store, err := tournament.NewStore(cfg.Tournament.Database)
			if err != nil {
				return fmt.Errorf("failed to open database: %w", err)
			}
			defer store.Close()

			ratings, matrix, recs, err := calculateTournamentResults(cfg, store, category)
			if err != nil {
				return err
			}

			var out *os.File = os.Stdout
			if outputPath != "" {
				f, err := os.Create(outputPath)
				if err != nil {
					return fmt.Errorf("failed to create output file: %w", err)
				}
				defer f.Close()
				out = f
			}

			tourneyTitle := cfg.Tournament.Name
			if category != "" {
				tourneyTitle = fmt.Sprintf("%s [Category: %s]", tourneyTitle, category)
			}

			switch format {
			case "markdown", "md":
				md := report.GenerateMarkdownReport(tourneyTitle, ratings, matrix, recs)
				_, err = fmt.Fprint(out, md)
				return err
			case "json":
				return report.ExportJSON(out, ratings, recs)
			case "csv":
				return report.ExportCSV(out, ratings)
			case "terminal":
				report.PrintTerminalLeaderboard(out, tourneyTitle, ratings, recs)
				return nil
			default:
				return fmt.Errorf("unsupported format %q (valid: terminal, markdown, json, csv)", format)
			}
		},
	}

	cmd.Flags().StringVarP(&cfgPath, "config", "c", "examples/evalarena.yaml", "Path to YAML configuration file")
	cmd.Flags().StringVarP(&format, "format", "f", "terminal", "Output format: terminal, markdown, json, csv")
	cmd.Flags().StringVarP(&outputPath, "output", "o", "", "File path to save report (defaults to stdout)")
	cmd.Flags().StringVar(&category, "category", "", "Filter report by prompt category (e.g. sql_synthesis)")

	return cmd
}

func newCheckCmd() *cobra.Command {
	var cfgPath string
	var candidateModel string
	var baselineModel string
	var maxELODrop float64
	var category string

	cmd := &cobra.Command{
		Use:   "check",
		Short: "CI/CD quality gate: verify candidate model ELO against production baseline",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load(cfgPath)
			if err != nil {
				return err
			}

			if candidateModel == "" {
				candidateModel = cfg.Gate.CandidateModel
			}
			if baselineModel == "" {
				baselineModel = cfg.Gate.BaselineModel
			}
			if maxELODrop <= 0 {
				maxELODrop = cfg.Gate.MaxELODrop
			}
			if maxELODrop <= 0 {
				maxELODrop = 30.0 // default 30 ELO points max drop
			}

			if candidateModel == "" || baselineModel == "" {
				return fmt.Errorf("both --candidate and --baseline models are required")
			}

			store, err := tournament.NewStore(cfg.Tournament.Database)
			if err != nil {
				return fmt.Errorf("failed to open database: %w", err)
			}
			defer store.Close()

			ratings, _, _, err := calculateTournamentResults(cfg, store, category)
			if err != nil {
				return err
			}

			ratingMap := make(map[string]elo.ModelRating)
			for _, r := range ratings {
				ratingMap[r.ModelID] = r
			}

			baseRating, okBase := ratingMap[baselineModel]
			if !okBase {
				return fmt.Errorf("baseline model %q has no recorded tournament matches", baselineModel)
			}

			candRating, okCand := ratingMap[candidateModel]
			if !okCand {
				return fmt.Errorf("candidate model %q has no recorded tournament matches", candidateModel)
			}

			eloDiff := candRating.ELO - baseRating.ELO
			allowedMin := -maxELODrop

			fmt.Printf("evalarena CI/CD Quality Gate Check\n")
			if category != "" {
				fmt.Printf("• Category Filter: %s\n", category)
			}
			fmt.Printf("• Baseline (%s):   ELO %.1f\n", baselineModel, baseRating.ELO)
			fmt.Printf("• Candidate (%s):  ELO %.1f\n", candidateModel, candRating.ELO)
			fmt.Printf("• ELO Difference:       %+.1f (Allowed max drop: -%.1f)\n", eloDiff, maxELODrop)

			if eloDiff < allowedMin {
				color.Red("\n[GATE FAILED] Regression detected! Candidate model ELO dropped by %.1f points (threshold: %.1f points).\n", -eloDiff, maxELODrop)
				os.Exit(1)
			}

			color.Green("\n[GATE PASSED] Candidate model meets quality threshold.\n")
			return nil
		},
	}

	cmd.Flags().StringVarP(&cfgPath, "config", "c", "examples/evalarena.yaml", "Path to YAML configuration file")
	cmd.Flags().StringVar(&candidateModel, "candidate", "", "Candidate model ID to verify")
	cmd.Flags().StringVar(&baselineModel, "baseline", "", "Production baseline model ID to compare against")
	cmd.Flags().Float64Var(&maxELODrop, "max-elo-drop", 0, "Maximum permissible ELO rating drop")
	cmd.Flags().StringVar(&category, "category", "", "Filter quality gate check by prompt category")

	return cmd
}

func newInitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Scaffold an evalarena.yaml configuration and sample dataset in the current directory",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfgFile := "evalarena.yaml"
			datasetFile := "dataset.jsonl"

			if _, err := os.Stat(cfgFile); err == nil {
				return fmt.Errorf("%s already exists", cfgFile)
			}

			defaultConfig := `version: "1.0"
tournament:
  name: "Internal Model Tournament"
  schedule: "round-robin"
  rounds: 2
  concurrency: 4
  database: "evalarena.db"

judge:
  provider: "mock"
  model: "mock-judge"
  temperature: 0.1
  debias_position: true

models:
  - id: "gpt-4o"
    provider: "mock"
    model: "gpt-4o"
    pricing:
      input_per_million: 2.50
      output_per_million: 10.00

  - id: "gemini-2.0-flash"
    provider: "mock"
    model: "gemini-2.0-flash"
    pricing:
      input_per_million: 0.10
      output_per_million: 0.40

dataset:
  path: "dataset.jsonl"
  format: "jsonl"
`
			defaultDataset := `{"id": "p001", "category": "general", "prompt": "Summarize key principles of zero trust architecture."}
{"id": "p002", "category": "general", "prompt": "Write a Python function to validate and format RFC 3339 timestamps."}
`

			if err := os.WriteFile(cfgFile, []byte(defaultConfig), 0644); err != nil {
				return err
			}
			if err := os.WriteFile(datasetFile, []byte(defaultDataset), 0644); err != nil {
				return err
			}

			fmt.Printf("Created %s and %s. Run `evalarena run` to start your first tournament.\n", cfgFile, datasetFile)
			return nil
		},
	}
	return cmd
}

func calculateTournamentResults(cfg *config.Config, store *tournament.Store, category string) ([]elo.ModelRating, *elo.WinRateMatrix, []elo.Recommendation, error) {
	matches, err := store.GetMatches(cfg.Tournament.Name, category)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to query matches: %w", err)
	}

	var modelIDs []string
	for _, m := range cfg.Models {
		modelIDs = append(modelIDs, m.ID)
	}

	var eloMatches []elo.MatchResult
	type stats struct {
		wins         int
		losses       int
		ties         int
		totalLatency int64
		totalCost    float64
		count        int
	}
	statMap := make(map[string]*stats)
	for _, id := range modelIDs {
		statMap[id] = &stats{}
	}

	for _, m := range matches {
		eloMatches = append(eloMatches, elo.MatchResult{
			ModelA: m.ModelAID,
			ModelB: m.ModelBID,
			Winner: m.FinalWinner,
		})

		sA := statMap[m.ModelAID]
		sB := statMap[m.ModelBID]

		if sA != nil {
			sA.totalLatency += m.LatencyAMs
			sA.totalCost += m.CostA
			sA.count++
		}
		if sB != nil {
			sB.totalLatency += m.LatencyBMs
			sB.totalCost += m.CostB
			sB.count++
		}

		if m.FinalWinner == m.ModelAID {
			if sA != nil {
				sA.wins++
			}
			if sB != nil {
				sB.losses++
			}
		} else if m.FinalWinner == m.ModelBID {
			if sB != nil {
				sB.wins++
			}
			if sA != nil {
				sA.losses++
			}
		} else {
			if sA != nil {
				sA.ties++
			}
			if sB != nil {
				sB.ties++
			}
		}
	}

	eloMap := elo.SolveBradleyTerry(modelIDs, eloMatches, 1000.0)
	cis := elo.ComputeBootstrapCIs(modelIDs, eloMatches, 1000.0, 500)
	matrix := elo.ComputeWinRateMatrix(modelIDs, eloMatches)

	var points []elo.ParetoPoint
	for _, id := range modelIDs {
		st := statMap[id]
		avgLat := 0.0
		avgCost := 0.0
		if st.count > 0 {
			avgLat = float64(st.totalLatency) / float64(st.count)
			avgCost = (st.totalCost / float64(st.count)) * 1000.0
		}
		points = append(points, elo.ParetoPoint{
			ModelID:   id,
			ELO:       eloMap[id],
			CostPerM:  avgCost,
			LatencyMs: avgLat,
		})
	}

	frontier := elo.ComputeParetoFrontier(points)
	recs := elo.GenerateRecommendations(frontier)
	paretoStatus := make(map[string]bool)
	for _, p := range frontier {
		paretoStatus[p.ModelID] = p.IsOptimal
	}

	var ratings []elo.ModelRating
	for _, id := range modelIDs {
		st := statMap[id]
		totalBattles := st.wins + st.losses + st.ties
		winRate := 0.0
		if totalBattles > 0 {
			winRate = (float64(st.wins) + 0.5*float64(st.ties)) / float64(totalBattles)
		}

		avgLat := 0.0
		avgCost := 0.0
		if st.count > 0 {
			avgLat = float64(st.totalLatency) / float64(st.count)
			avgCost = (st.totalCost / float64(st.count)) * 1000.0
		}

		ratings = append(ratings, elo.ModelRating{
			ModelID:         id,
			ELO:             eloMap[id],
			CILower:         cis[id].Lower,
			CIUpper:         cis[id].Upper,
			Wins:            st.wins,
			Losses:          st.losses,
			Ties:            st.ties,
			TotalBattles:    totalBattles,
			WinRate:         winRate,
			AvgLatencyMs:    avgLat,
			AvgCostPerM:     avgCost,
			IsParetoOptimal: paretoStatus[id],
		})
	}

	elo.SortRatingsByELO(ratings)
	return ratings, matrix, recs, nil
}
