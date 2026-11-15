export interface ModelRating {
  model_id: string;
  elo: number;
  ci_lower: number;
  ci_upper: number;
  wins: number;
  losses: number;
  ties: number;
  total_battles: number;
  win_rate: number;
  avg_latency_ms: number;
  avg_cost_per_m: number;
  is_pareto_optimal: boolean;
}

export interface HeadToHeadRecord {
  wins: number;
  losses: number;
  ties: number;
  total: number;
  win_rate: number;
}

export interface WinRateMatrix {
  models: string[];
  data: Record<string, Record<string, HeadToHeadRecord>>;
}

export interface ParetoPoint {
  model_id: string;
  elo: number;
  cost_per_m: number;
  latency_ms: number;
  is_optimal: boolean;
}

export interface Recommendation {
  model_id: string;
  elo_rating: number;
  cost_per_m: number;
  latency_ms: number;
  elo_retention_pct: number;
  cost_reduction_pct: number;
  summary: string;
}

export interface MatchRecord {
  id: string;
  tournament_id: string;
  prompt_id: string;
  category: string;
  prompt_text: string;
  round_num: number;
  model_a_id: string;
  model_b_id: string;
  response_a: string;
  response_b: string;
  latency_a_ms: number;
  latency_b_ms: number;
  cost_a: number;
  cost_b: number;
  tokens_a: number;
  tokens_b: number;
  pass1_winner: string;
  pass1_reasoning: string;
  pass2_winner: string;
  pass2_reasoning: string;
  final_winner: string;
  position_bias_detected: boolean;
  human_vote?: string;
  created_at: string;
}

export interface BlindMatch {
  match_id: string;
  prompt: string;
  category?: string;
  response_a: string;
  response_b: string;
  round_num: number;
}

export interface VoteReveal {
  match_id: string;
  vote: string;
  model_a_id: string;
  model_b_id: string;
  model_a_latency_ms: number;
  model_b_latency_ms: number;
  judge_verdict: string;
  bias_detected: boolean;
}

export interface CalibrationReport {
  total_comparisons: number;
  agreements: number;
  observed_agreement_pct: number;
  expected_agreement_pct: number;
  cohens_kappa: number;
  interpretation: string;
}

export interface QuickBattleRequest {
  prompt: string;
  system_prompt?: string;
  model_a_id: string;
  model_b_id: string;
  rubric?: string;
}

export interface QuickBattleResponse {
  match_id: string;
  model_a_id: string;
  model_b_id: string;
  response_a: string;
  response_b: string;
  latency_a_ms: number;
  latency_b_ms: number;
  cost_a: number;
  cost_b: number;
  tokens_a: number;
  tokens_b: number;
  pass1_winner: string;
  pass1_reasoning: string;
  pass2_winner: string;
  pass2_reasoning: string;
  final_winner: string;
  position_bias_detected: boolean;
}

const API_BASE = '/api';

export async function fetchCategories(): Promise<string[]> {
  const res = await fetch(`${API_BASE}/categories`);
  if (!res.ok) return [];
  return res.json();
}

export async function fetchCalibration(): Promise<CalibrationReport> {
  const res = await fetch(`${API_BASE}/calibration`);
  if (!res.ok) throw new Error(`Failed to fetch calibration: ${res.statusText}`);
  return res.json();
}

export async function fetchLeaderboard(category?: string): Promise<ModelRating[]> {
  const url = category ? `${API_BASE}/leaderboard?category=${encodeURIComponent(category)}` : `${API_BASE}/leaderboard`;
  const res = await fetch(url);
  if (!res.ok) throw new Error(`Failed to fetch leaderboard: ${res.statusText}`);
  return res.json();
}

export async function fetchMatrix(category?: string): Promise<WinRateMatrix> {
  const url = category ? `${API_BASE}/matrix?category=${encodeURIComponent(category)}` : `${API_BASE}/matrix`;
  const res = await fetch(url);
  if (!res.ok) throw new Error(`Failed to fetch matrix: ${res.statusText}`);
  return res.json();
}

export async function fetchPareto(category?: string): Promise<{ points: ParetoPoint[]; recommendations: Recommendation[] }> {
  const url = category ? `${API_BASE}/pareto?category=${encodeURIComponent(category)}` : `${API_BASE}/pareto`;
  const res = await fetch(url);
  if (!res.ok) throw new Error(`Failed to fetch pareto: ${res.statusText}`);
  return res.json();
}

export async function fetchMatches(category?: string): Promise<MatchRecord[]> {
  const url = category ? `${API_BASE}/matches?category=${encodeURIComponent(category)}` : `${API_BASE}/matches`;
  const res = await fetch(url);
  if (!res.ok) throw new Error(`Failed to fetch matches: ${res.statusText}`);
  return res.json();
}

export async function fetchNextArenaMatch(): Promise<BlindMatch> {
  const res = await fetch(`${API_BASE}/arena/next`);
  if (!res.ok) throw new Error(`Failed to fetch next arena match: ${res.statusText}`);
  return res.json();
}

export async function submitArenaVote(matchId: string, vote: string): Promise<VoteReveal> {
  const res = await fetch(`${API_BASE}/arena/vote`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ match_id: matchId, vote }),
  });
  if (!res.ok) throw new Error(`Failed to submit vote: ${res.statusText}`);
  return res.json();
}

export async function quickBattle(req: QuickBattleRequest): Promise<QuickBattleResponse> {
  const res = await fetch(`${API_BASE}/battle/quick`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(req),
  });
  if (!res.ok) throw new Error(`Quick battle failed: ${res.statusText}`);
  return res.json();
}

export async function startTournament(): Promise<{ status: string; message: string }> {
  const res = await fetch(`${API_BASE}/tournament/start`, { method: 'POST' });
  if (!res.ok) throw new Error(`Failed to start tournament: ${res.statusText}`);
  return res.json();
}

export function subscribeToEvents(onEvent: (event: { type: string; payload: any }) => void): () => void {
  const eventSource = new EventSource(`${API_BASE}/events`);

  eventSource.addEventListener('match_start', (e) => {
    try { onEvent({ type: 'match_start', payload: JSON.parse(e.data) }); } catch {}
  });

  eventSource.addEventListener('match_finish', (e) => {
    try { onEvent({ type: 'match_finish', payload: JSON.parse(e.data) }); } catch {}
  });

  eventSource.addEventListener('tournament_complete', (e) => {
    try { onEvent({ type: 'tournament_complete', payload: JSON.parse(e.data) }); } catch {}
  });

  eventSource.addEventListener('human_vote', (e) => {
    try { onEvent({ type: 'human_vote', payload: JSON.parse(e.data) }); } catch {}
  });

  return () => {
    eventSource.close();
  };
}
