package tournament

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

// MatchRecord represents a persisted pairwise battle.
type MatchRecord struct {
	ID                   string     `json:"id"`
	TournamentID         string     `json:"tournament_id"`
	PromptID             string     `json:"prompt_id"`
	Category             string     `json:"category"`
	PromptText           string     `json:"prompt_text"`
	RoundNum             int        `json:"round_num"`
	ModelAID             string     `json:"model_a_id"`
	ModelBID             string     `json:"model_b_id"`
	ResponseA            string     `json:"response_a"`
	ResponseB            string     `json:"response_b"`
	LatencyAMs           int64      `json:"latency_a_ms"`
	LatencyBMs           int64      `json:"latency_b_ms"`
	CostA                float64    `json:"cost_a"`
	CostB                float64    `json:"cost_b"`
	TokensA              int        `json:"tokens_a"`
	TokensB              int        `json:"tokens_b"`
	Pass1Winner          string     `json:"pass1_winner"`
	Pass1Reasoning       string     `json:"pass1_reasoning"`
	Pass2Winner          string     `json:"pass2_winner"`
	Pass2Reasoning       string     `json:"pass2_reasoning"`
	FinalWinner          string     `json:"final_winner"` // model_a_id, model_b_id, or "tie"
	PositionBiasDetected bool       `json:"position_bias_detected"`
	HumanVote            string     `json:"human_vote,omitempty"`
	HumanVotedAt         *time.Time `json:"human_voted_at,omitempty"`
	CreatedAt            time.Time  `json:"created_at"`
}

// LeaderboardRecord represents a persisted ELO entry.
type LeaderboardRecord struct {
	ModelID         string  `json:"model_id"`
	ELO             float64 `json:"elo"`
	CILower         float64 `json:"ci_lower"`
	CIUpper         float64 `json:"ci_upper"`
	Wins            int     `json:"wins"`
	Losses          int     `json:"losses"`
	Ties            int     `json:"ties"`
	WinRate         float64 `json:"win_rate"`
	AvgLatencyMs    float64 `json:"avg_latency_ms"`
	AvgCostPerM     float64 `json:"avg_cost_per_m"`
	IsParetoOptimal bool    `json:"is_pareto_optimal"`
}

// Store handles SQLite database operations with zero CGO dependencies.
type Store struct {
	db *sql.DB
}

// NewStore opens or creates an embedded SQLite database and initializes tables.
func NewStore(dbPath string) (*Store, error) {
	if dir := filepath.Dir(dbPath); dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create database directory: %w", err)
		}
	}

	dsn := fmt.Sprintf("%s?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)", dbPath)
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite database %q: %w", dbPath, err)
	}

	s := &Store{db: db}
	if err := s.migrate(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to run database migrations: %w", err)
	}

	return s, nil
}

func (s *Store) migrate() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS tournaments (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			created_at DATETIME NOT NULL,
			status TEXT NOT NULL,
			config_json TEXT
		);`,
		`CREATE TABLE IF NOT EXISTS matches (
			id TEXT PRIMARY KEY,
			tournament_id TEXT NOT NULL,
			prompt_id TEXT NOT NULL,
			category TEXT NOT NULL DEFAULT '',
			prompt_text TEXT NOT NULL,
			round_num INTEGER NOT NULL,
			model_a_id TEXT NOT NULL,
			model_b_id TEXT NOT NULL,
			response_a TEXT NOT NULL,
			response_b TEXT NOT NULL,
			latency_a_ms INTEGER NOT NULL,
			latency_b_ms INTEGER NOT NULL,
			cost_a REAL NOT NULL,
			cost_b REAL NOT NULL,
			tokens_a INTEGER NOT NULL,
			tokens_b INTEGER NOT NULL,
			pass1_winner TEXT NOT NULL,
			pass1_reasoning TEXT,
			pass2_winner TEXT NOT NULL,
			pass2_reasoning TEXT,
			final_winner TEXT NOT NULL,
			position_bias_detected BOOLEAN NOT NULL DEFAULT 0,
			human_vote TEXT,
			human_voted_at DATETIME,
			created_at DATETIME NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS leaderboards (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			tournament_id TEXT NOT NULL,
			model_id TEXT NOT NULL,
			elo REAL NOT NULL,
			ci_lower REAL NOT NULL,
			ci_upper REAL NOT NULL,
			wins INTEGER NOT NULL,
			losses INTEGER NOT NULL,
			ties INTEGER NOT NULL,
			win_rate REAL NOT NULL,
			avg_latency_ms REAL NOT NULL,
			avg_cost_per_m REAL NOT NULL,
			is_pareto_optimal BOOLEAN NOT NULL DEFAULT 0,
			updated_at DATETIME NOT NULL,
			UNIQUE(tournament_id, model_id)
		);`,
		`CREATE INDEX IF NOT EXISTS idx_matches_tourney ON matches(tournament_id);`,
		`CREATE INDEX IF NOT EXISTS idx_matches_models ON matches(model_a_id, model_b_id);`,
	}

	for _, q := range queries {
		if _, err := s.db.Exec(q); err != nil {
			return fmt.Errorf("migration query error: %w", err)
		}
	}

	// Graceful column addition if table existed without category
	_, _ = s.db.Exec(`ALTER TABLE matches ADD COLUMN category TEXT NOT NULL DEFAULT ''`)

	return nil
}

// Close closes the database handle.
func (s *Store) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

// SaveMatch persists or updates a pairwise match record.
func (s *Store) SaveMatch(m *MatchRecord) error {
	query := `INSERT INTO matches (
		id, tournament_id, prompt_id, category, prompt_text, round_num,
		model_a_id, model_b_id, response_a, response_b,
		latency_a_ms, latency_b_ms, cost_a, cost_b, tokens_a, tokens_b,
		pass1_winner, pass1_reasoning, pass2_winner, pass2_reasoning,
		final_winner, position_bias_detected, human_vote, human_voted_at, created_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	ON CONFLICT(id) DO UPDATE SET
		category=excluded.category,
		human_vote=excluded.human_vote,
		human_voted_at=excluded.human_voted_at,
		final_winner=excluded.final_winner`

	_, err := s.db.Exec(query,
		m.ID, m.TournamentID, m.PromptID, m.Category, m.PromptText, m.RoundNum,
		m.ModelAID, m.ModelBID, m.ResponseA, m.ResponseB,
		m.LatencyAMs, m.LatencyBMs, m.CostA, m.CostB, m.TokensA, m.TokensB,
		m.Pass1Winner, m.Pass1Reasoning, m.Pass2Winner, m.Pass2Reasoning,
		m.FinalWinner, m.PositionBiasDetected, m.HumanVote, m.HumanVotedAt, m.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to save match %s: %w", m.ID, err)
	}

	return nil
}

// GetMatches returns all recorded matches for a tournament, optionally filtered by category.
func (s *Store) GetMatches(tournamentID string, category string) ([]MatchRecord, error) {
	query := `SELECT 
		id, tournament_id, prompt_id, category, prompt_text, round_num,
		model_a_id, model_b_id, response_a, response_b,
		latency_a_ms, latency_b_ms, cost_a, cost_b, tokens_a, tokens_b,
		pass1_winner, pass1_reasoning, pass2_winner, pass2_reasoning,
		final_winner, position_bias_detected, human_vote, human_voted_at, created_at
	FROM matches 
	WHERE (tournament_id = ? OR ? = '') AND (category = ? OR ? = '')
	ORDER BY round_num ASC, created_at ASC`

	rows, err := s.db.Query(query, tournamentID, tournamentID, category, category)
	if err != nil {
		return nil, fmt.Errorf("failed to query matches: %w", err)
	}
	defer rows.Close()

	var matches []MatchRecord
	for rows.Next() {
		var m MatchRecord
		var humanVote sql.NullString
		var humanVotedAt sql.NullTime

		err := rows.Scan(
			&m.ID, &m.TournamentID, &m.PromptID, &m.Category, &m.PromptText, &m.RoundNum,
			&m.ModelAID, &m.ModelBID, &m.ResponseA, &m.ResponseB,
			&m.LatencyAMs, &m.LatencyBMs, &m.CostA, &m.CostB, &m.TokensA, &m.TokensB,
			&m.Pass1Winner, &m.Pass1Reasoning, &m.Pass2Winner, &m.Pass2Reasoning,
			&m.FinalWinner, &m.PositionBiasDetected, &humanVote, &humanVotedAt, &m.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan match row: %w", err)
		}

		if humanVote.Valid {
			m.HumanVote = humanVote.String
		}
		if humanVotedAt.Valid {
			m.HumanVotedAt = &humanVotedAt.Time
		}

		matches = append(matches, m)
	}

	return matches, nil
}

// GetCategories returns all distinct categories present in recorded matches.
func (s *Store) GetCategories(tournamentID string) ([]string, error) {
	query := `SELECT DISTINCT category FROM matches WHERE (tournament_id = ? OR ? = '') AND category != '' ORDER BY category ASC`
	rows, err := s.db.Query(query, tournamentID, tournamentID)
	if err != nil {
		return nil, fmt.Errorf("failed to query categories: %w", err)
	}
	defer rows.Close()

	var categories []string
	for rows.Next() {
		var cat string
		if err := rows.Scan(&cat); err != nil {
			return nil, err
		}
		categories = append(categories, cat)
	}
	return categories, nil
}

// GetUnvotedMatch returns a random match waiting for human blind evaluation.
func (s *Store) GetUnvotedMatch(tournamentID string) (*MatchRecord, error) {
	query := `SELECT 
		id, tournament_id, prompt_id, category, prompt_text, round_num,
		model_a_id, model_b_id, response_a, response_b,
		latency_a_ms, latency_b_ms, cost_a, cost_b, tokens_a, tokens_b,
		pass1_winner, pass1_reasoning, pass2_winner, pass2_reasoning,
		final_winner, position_bias_detected, created_at
	FROM matches 
	WHERE (tournament_id = ? OR ? = '') AND human_vote IS NULL
	ORDER BY RANDOM() LIMIT 1`

	row := s.db.QueryRow(query, tournamentID, tournamentID)
	var m MatchRecord
	err := row.Scan(
		&m.ID, &m.TournamentID, &m.PromptID, &m.Category, &m.PromptText, &m.RoundNum,
		&m.ModelAID, &m.ModelBID, &m.ResponseA, &m.ResponseB,
		&m.LatencyAMs, &m.LatencyBMs, &m.CostA, &m.CostB, &m.TokensA, &m.TokensB,
		&m.Pass1Winner, &m.Pass1Reasoning, &m.Pass2Winner, &m.Pass2Reasoning,
		&m.FinalWinner, &m.PositionBiasDetected, &m.CreatedAt,
	)
	if err == sql.ErrNoRows {
		// If all have votes, return any random match for arena practice
		fallbackQuery := `SELECT 
			id, tournament_id, prompt_id, category, prompt_text, round_num,
			model_a_id, model_b_id, response_a, response_b,
			latency_a_ms, latency_b_ms, cost_a, cost_b, tokens_a, tokens_b,
			pass1_winner, pass1_reasoning, pass2_winner, pass2_reasoning,
			final_winner, position_bias_detected, created_at
		FROM matches 
		WHERE (tournament_id = ? OR ? = '')
		ORDER BY RANDOM() LIMIT 1`

		fbRow := s.db.QueryRow(fallbackQuery, tournamentID, tournamentID)
		if fbErr := fbRow.Scan(
			&m.ID, &m.TournamentID, &m.PromptID, &m.Category, &m.PromptText, &m.RoundNum,
			&m.ModelAID, &m.ModelBID, &m.ResponseA, &m.ResponseB,
			&m.LatencyAMs, &m.LatencyBMs, &m.CostA, &m.CostB, &m.TokensA, &m.TokensB,
			&m.Pass1Winner, &m.Pass1Reasoning, &m.Pass2Winner, &m.Pass2Reasoning,
			&m.FinalWinner, &m.PositionBiasDetected, &m.CreatedAt,
		); fbErr != nil {
			return nil, fbErr
		}
		return &m, nil
	} else if err != nil {
		return nil, fmt.Errorf("failed to query unvoted match: %w", err)
	}

	return &m, nil
}

// RecordHumanVote updates a match with human preference.
func (s *Store) RecordHumanVote(matchID string, vote string) (*MatchRecord, error) {
	now := time.Now()
	query := `UPDATE matches SET human_vote = ?, human_voted_at = ? WHERE id = ?`
	res, err := s.db.Exec(query, vote, now, matchID)
	if err != nil {
		return nil, fmt.Errorf("failed to record human vote: %w", err)
	}
	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		return nil, fmt.Errorf("match %q not found", matchID)
	}

	// Fetch updated record
	var m MatchRecord
	selQuery := `SELECT 
		id, tournament_id, prompt_id, category, prompt_text, round_num,
		model_a_id, model_b_id, response_a, response_b,
		latency_a_ms, latency_b_ms, cost_a, cost_b, tokens_a, tokens_b,
		pass1_winner, pass1_reasoning, pass2_winner, pass2_reasoning,
		final_winner, position_bias_detected, human_vote, created_at
	FROM matches WHERE id = ?`
	row := s.db.QueryRow(selQuery, matchID)
	var hVote sql.NullString
	if err := row.Scan(
		&m.ID, &m.TournamentID, &m.PromptID, &m.Category, &m.PromptText, &m.RoundNum,
		&m.ModelAID, &m.ModelBID, &m.ResponseA, &m.ResponseB,
		&m.LatencyAMs, &m.LatencyBMs, &m.CostA, &m.CostB, &m.TokensA, &m.TokensB,
		&m.Pass1Winner, &m.Pass1Reasoning, &m.Pass2Winner, &m.Pass2Reasoning,
		&m.FinalWinner, &m.PositionBiasDetected, &hVote, &m.CreatedAt,
	); err != nil {
		return nil, err
	}
	if hVote.Valid {
		m.HumanVote = hVote.String
	}
	m.HumanVotedAt = &now

	return &m, nil
}

// SaveLeaderboard stores the updated ELO standings for a tournament.
func (s *Store) SaveLeaderboard(tournamentID string, records []LeaderboardRecord) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `INSERT INTO leaderboards (
		tournament_id, model_id, elo, ci_lower, ci_upper,
		wins, losses, ties, win_rate, avg_latency_ms, avg_cost_per_m,
		is_pareto_optimal, updated_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	ON CONFLICT(tournament_id, model_id) DO UPDATE SET
		elo=excluded.elo,
		ci_lower=excluded.ci_lower,
		ci_upper=excluded.ci_upper,
		wins=excluded.wins,
		losses=excluded.losses,
		ties=excluded.ties,
		win_rate=excluded.win_rate,
		avg_latency_ms=excluded.avg_latency_ms,
		avg_cost_per_m=excluded.avg_cost_per_m,
		is_pareto_optimal=excluded.is_pareto_optimal,
		updated_at=excluded.updated_at`

	now := time.Now()
	for _, r := range records {
		_, err := tx.Exec(query,
			tournamentID, r.ModelID, r.ELO, r.CILower, r.CIUpper,
			r.Wins, r.Losses, r.Ties, r.WinRate, r.AvgLatencyMs, r.AvgCostPerM,
			r.IsParetoOptimal, now,
		)
		if err != nil {
			return fmt.Errorf("failed to insert leaderboard record: %w", err)
		}
	}

	return tx.Commit()
}

// GetLeaderboard retrieves saved ELO rankings for a tournament.
func (s *Store) GetLeaderboard(tournamentID string) ([]LeaderboardRecord, error) {
	query := `SELECT 
		model_id, elo, ci_lower, ci_upper,
		wins, losses, ties, win_rate, avg_latency_ms, avg_cost_per_m, is_pareto_optimal
	FROM leaderboards WHERE tournament_id = ? ORDER BY elo DESC`

	rows, err := s.db.Query(query, tournamentID)
	if err != nil {
		return nil, fmt.Errorf("failed to query leaderboard: %w", err)
	}
	defer rows.Close()

	var records []LeaderboardRecord
	for rows.Next() {
		var r LeaderboardRecord
		err := rows.Scan(
			&r.ModelID, &r.ELO, &r.CILower, &r.CIUpper,
			&r.Wins, &r.Losses, &r.Ties, &r.WinRate, &r.AvgLatencyMs, &r.AvgCostPerM, &r.IsParetoOptimal,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan leaderboard row: %w", err)
		}
		records = append(records, r)
	}

	return records, nil
}
