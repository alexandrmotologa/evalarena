# evalarena REST and SSE API Specification

When running `evalarena arena --port 7070`, the embedded HTTP server exposes a REST and Server-Sent Events (SSE) API to power the Arena Studio and external automation.

## Endpoints

### 1. Leaderboard and Rankings
- **GET `/api/leaderboard`**
  Returns models ranked by Bradley-Terry ELO rating, with 95% bootstrap confidence intervals, win/loss records, average latency, and Pareto frontier membership.
  ```json
  [
    {
      "model_id": "claude-3-5-sonnet",
      "elo": 1142.5,
      "ci_lower": 1118.2,
      "ci_upper": 1168.0,
      "wins": 42,
      "losses": 14,
      "ties": 4,
      "win_rate": 0.733,
      "avg_latency_ms": 780,
      "avg_cost_per_m": 4.80,
      "is_pareto_optimal": true
    }
  ]
  ```

### 2. Pairwise Win-Rate Matrix
- **GET `/api/matrix`**
  Returns pairwise head-to-head win, loss, and tie counts between every pair of evaluated models.
  ```json
  {
    "models": ["gpt-4o", "claude-3-5-sonnet", "gemini-2.0-flash"],
    "matrix": {
      "gpt-4o": {
        "claude-3-5-sonnet": { "wins": 12, "losses": 14, "ties": 4 },
        "gemini-2.0-flash": { "wins": 18, "losses": 10, "ties": 2 }
      }
    }
  }
  ```

### 3. Pareto Efficiency Frontier
- **GET `/api/pareto`**
  Returns the computed Pareto efficiency frontier data points (ELO rating, blended cost per 1M tokens, latency in milliseconds, and dominance status).

### 4. Human Blind Arena
- **GET `/api/arena/next`**
  Fetches an unvoted or random pairwise completion pair with anonymized model identifiers (`model_a` and `model_b`).
  ```json
  {
    "match_id": "m_9831f2a",
    "prompt": "Summarize the customer refund policy for digital goods.",
    "response_a": "Digital goods are eligible for refund within 14 days...",
    "response_b": "Customers may request refunds for digital downloads within 14 days..."
  }
  ```

- **POST `/api/arena/vote`**
  Submits a human vote on an anonymized match.
  Request:
  ```json
  {
    "match_id": "m_9831f2a",
    "vote": "model_a" // "model_a" | "model_b" | "tie" | "both_bad"
  }
  ```
  Response reveals model identities, latency, token counts, and updated ELO delta:
  ```json
  {
    "match_id": "m_9831f2a",
    "model_a": "gpt-4o",
    "model_b": "gemini-2.0-flash",
    "vote": "model_a",
    "model_a_latency_ms": 620,
    "model_b_latency_ms": 310
  }
  ```

### 5. Live Tournament Stream
- **GET `/api/events`**
  Server-Sent Events (SSE) stream broadcasting real-time progress:
  - `round_start`: Dispatched when a new pair battle begins.
  - `match_complete`: Dispatched with completion outputs and judge verdicts.
  - `elo_update`: Dispatched with recomputed ELO ratings after each round.
