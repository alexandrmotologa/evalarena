package server

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/alexandrmotologa/evalarena/internal/config"
	"github.com/alexandrmotologa/evalarena/internal/dataset"
	"github.com/alexandrmotologa/evalarena/internal/elo"
	"github.com/alexandrmotologa/evalarena/internal/provider"
	"github.com/alexandrmotologa/evalarena/internal/tournament"
)

// Server coordinates HTTP API endpoints and tournament execution.
type Server struct {
	cfg       *config.Config
	store     *tournament.Store
	bus       *tournament.EventBus
	pool      *provider.Pool
	dataset   []dataset.Item
	mux       *http.ServeMux
	runningMu sync.Mutex
	isRunning bool
}

// NewServer initializes the HTTP API router and handlers.
func NewServer(cfg *config.Config, store *tournament.Store, bus *tournament.EventBus, pool *provider.Pool, items []dataset.Item) *Server {
	if bus == nil {
		bus = tournament.NewEventBus()
	}

	s := &Server{
		cfg:     cfg,
		store:   store,
		bus:     bus,
		pool:    pool,
		dataset: items,
		mux:     http.NewServeMux(),
	}

	s.routes()
	return s
}

// Handler returns the HTTP request handler with CORS enabled.
func (s *Server) Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		s.mux.ServeHTTP(w, r)
	})
}

func (s *Server) routes() {
	s.mux.HandleFunc("GET /api/health", s.handleHealth)
	s.mux.HandleFunc("GET /api/leaderboard", s.handleLeaderboard)
	s.mux.HandleFunc("GET /api/matrix", s.handleMatrix)
	s.mux.HandleFunc("GET /api/pareto", s.handlePareto)
	s.mux.HandleFunc("GET /api/matches", s.handleMatches)
	s.mux.HandleFunc("GET /api/arena/next", s.handleArenaNext)
	s.mux.HandleFunc("POST /api/arena/vote", s.handleArenaVote)
	s.mux.HandleFunc("POST /api/tournament/start", s.handleTournamentStart)
	s.mux.HandleFunc("GET /api/events", HandleSSE(s.bus))
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status":     "healthy",
		"service":    "evalarena",
		"tournament": s.cfg.Tournament.Name,
		"timestamp":  time.Now(),
	})
}

func (s *Server) handleLeaderboard(w http.ResponseWriter, r *http.Request) {
	ratings, err := s.computeCurrentRatings()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, ratings)
}

func (s *Server) handleMatrix(w http.ResponseWriter, r *http.Request) {
	matches, err := s.store.GetMatches(s.cfg.Tournament.Name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var modelIDs []string
	for _, m := range s.cfg.Models {
		modelIDs = append(modelIDs, m.ID)
	}

	var eloMatches []elo.MatchResult
	for _, m := range matches {
		eloMatches = append(eloMatches, elo.MatchResult{
			ModelA: m.ModelAID,
			ModelB: m.ModelBID,
			Winner: m.FinalWinner,
		})
	}

	matrix := elo.ComputeWinRateMatrix(modelIDs, eloMatches)
	writeJSON(w, http.StatusOK, matrix)
}

func (s *Server) handlePareto(w http.ResponseWriter, r *http.Request) {
	ratings, err := s.computeCurrentRatings()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var points []elo.ParetoPoint
	for _, r := range ratings {
		points = append(points, elo.ParetoPoint{
			ModelID:   r.ModelID,
			ELO:       r.ELO,
			CostPerM:  r.AvgCostPerM,
			LatencyMs: r.AvgLatencyMs,
		})
	}

	frontier := elo.ComputeParetoFrontier(points)
	recs := elo.GenerateRecommendations(frontier)

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"points":          frontier,
		"recommendations": recs,
	})
}

func (s *Server) handleMatches(w http.ResponseWriter, r *http.Request) {
	matches, err := s.store.GetMatches(s.cfg.Tournament.Name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, matches)
}

func (s *Server) handleArenaNext(w http.ResponseWriter, r *http.Request) {
	match, err := s.store.GetUnvotedMatch(s.cfg.Tournament.Name)
	if err != nil {
		http.Error(w, "no matches available in database, run a tournament first", http.StatusNotFound)
		return
	}

	// Anonymize model identities for blind evaluation
	blindView := map[string]interface{}{
		"match_id":   match.ID,
		"prompt":     match.PromptText,
		"response_a": match.ResponseA,
		"response_b": match.ResponseB,
		"round_num":  match.RoundNum,
	}

	writeJSON(w, http.StatusOK, blindView)
}

type voteRequest struct {
	MatchID string `json:"match_id"`
	Vote    string `json:"vote"` // "model_a", "model_b", "tie", "both_bad"
}

func (s *Server) handleArenaVote(w http.ResponseWriter, r *http.Request) {
	var req voteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	record, err := s.store.RecordHumanVote(req.MatchID, req.Vote)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Reveal identities and stats upon receiving human vote
	reveal := map[string]interface{}{
		"match_id":           record.ID,
		"vote":               record.HumanVote,
		"model_a_id":         record.ModelAID,
		"model_b_id":         record.ModelBID,
		"model_a_latency_ms": record.LatencyAMs,
		"model_b_latency_ms": record.LatencyBMs,
		"judge_verdict":      record.FinalWinner,
		"bias_detected":      record.PositionBiasDetected,
	}

	s.bus.Publish("human_vote", reveal)
	writeJSON(w, http.StatusOK, reveal)
}

func (s *Server) handleTournamentStart(w http.ResponseWriter, r *http.Request) {
	s.runningMu.Lock()
	if s.isRunning {
		s.runningMu.Unlock()
		http.Error(w, "tournament is already running", http.StatusConflict)
		return
	}
	s.isRunning = true
	s.runningMu.Unlock()

	go func() {
		defer func() {
			s.runningMu.Lock()
			s.isRunning = false
			s.runningMu.Unlock()
		}()

		runner := tournament.NewRunner(s.cfg, s.dataset, s.pool, s.store, s.bus)
		_, err := runner.Run(context.Background())
		if err != nil {
			s.bus.Publish("tournament_error", map[string]string{"error": err.Error()})
			return
		}

		ratings, _ := s.computeCurrentRatings()
		s.bus.Publish("tournament_complete", ratings)
	}()

	writeJSON(w, http.StatusAccepted, map[string]string{
		"status":  "started",
		"message": "Tournament execution started in background",
	})
}

func (s *Server) computeCurrentRatings() ([]elo.ModelRating, error) {
	matches, err := s.store.GetMatches(s.cfg.Tournament.Name)
	if err != nil {
		return nil, err
	}

	var modelIDs []string
	for _, m := range s.cfg.Models {
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
	cis := elo.ComputeBootstrapCIs(modelIDs, eloMatches, 1000.0, 300)

	var points []elo.ParetoPoint
	for _, id := range modelIDs {
		st := statMap[id]
		avgLat := 0.0
		avgCost := 0.0
		if st.count > 0 {
			avgLat = float64(st.totalLatency) / float64(st.count)
			avgCost = (st.totalCost / float64(st.count)) * 1000.0 // scaled to per 1M estimate
		}
		points = append(points, elo.ParetoPoint{
			ModelID:   id,
			ELO:       eloMap[id],
			CostPerM:  avgCost,
			LatencyMs: avgLat,
		})
	}

	frontier := elo.ComputeParetoFrontier(points)
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
	return ratings, nil
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}
