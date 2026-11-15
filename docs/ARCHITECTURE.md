# evalarena Architecture

evalarena provides self-hosted blind pairwise evaluation tournaments for large language models. The platform runs as a single Go binary with an embedded React 19 interface, keeping operational memory consumption below 25MB RAM under standard evaluation workloads.

## System Design Overview

The architecture consists of seven core subsystems:

```
+-----------------------------------------------------------------+
|                         evalarena CLI                           |
|       (cmd/evalarena: run, arena, report, check, init)          |
+-----------------------------------------------------------------+
                                 |
                                 v
+-----------------------------------------------------------------+
|                     Tournament Orchestration                    |
|   - Matchmaker: Round-Robin, Swiss, Random Pairing              |
|   - Runner: Parallel generation across competing models         |
|   - State and Session Management                                |
+-----------------------------------------------------------------+
           |                                     |
           v                                     v
+-----------------------+             +---------------------------+
|    Provider Engine    |             |    Evaluation Engine      |
| - Universal Client    |             | - Double-blind LLM Judge  |
| - OpenAI / Anthropic  |             | - Order swapping (A/B,B/A)|
| - Gemini / Ollama     |             | - Consistency verification|
| - Deterministic Mock  |             | - Human Blind Arena UI    |
| - Token & Cost Meters |             +---------------------------+
+-----------------------+                           |
           |                                        v
           |                          +---------------------------+
           |                          | Bradley-Terry ELO Engine  |
           |                          | - MLE via Newton-Raphson  |
           |                          | - Bootstrap 95% CI (1000) |
           |                          | - Pareto Frontier Solver  |
           +------------------------->+---------------------------+
                                                    |
                                                    v
+-----------------------------------------------------------------+
|                    Storage & Output Layer                       |
|   - Pure Go SQLite Database (modernc.org/sqlite, zero CGO)      |
|   - Embedded React 19 Studio (:7070) with SSE Live Updates      |
|   - Terminal ASCII reports, Markdown executive summaries, CSV   |
+-----------------------------------------------------------------+
```

## Subsystem Details

### 1. Provider Engine (`internal/provider`)
The provider layer abstracts model API differences behind a uniform interface:
- Supports streaming and non-streaming completion requests.
- Measures time-to-first-token (TTFT) and total generation latency.
- Calculates exact pricing based on token consumption (prompt and completion rates).
- Provides a deterministic mock provider with configurable latency and quality distributions for offline testing and CI workflows.

### 2. Tournament Engine (`internal/tournament`)
Coordinates match scheduling across model pairs:
- **Round-Robin**: Evaluates all combinations $(M_i, M_j)$ for exhaustive evaluation.
- **Swiss Scheduling**: Pairs models with similar ELO ratings to maximize informative pairwise comparisons with fewer rounds.
- **Random Pairing**: Uniform sampling for quick checks over large model catalogs.

### 3. Evaluation and Debiasing Engine (`internal/judge`)
Position bias causes language models to favor whichever response appears first in the evaluation prompt:
- evalarena executes two symmetric evaluation passes per pair:
  - Pass 1: Candidate A in position 1, Candidate B in position 2.
  - Pass 2: Candidate B in position 1, Candidate A in position 2.
- A candidate wins only when selected as superior across both presentations.
- Inconsistent outcomes are classified as ties or flagged for human review.

### 4. Bradley-Terry ELO Solver (`internal/elo`)
Calculates latent skill ratings from pairwise win-loss matrices using Maximum Likelihood Estimation:
- Newton-Raphson iteration converges on optimal parameters.
- Non-parametric bootstrap resampling ($N = 1,000$ iterations) generates 95% empirical confidence intervals without assuming asymptotic normality.
- Multi-objective Pareto frontier calculation isolates models that optimize the balance of ELO rating, token cost, and response latency.

### 5. Storage Layer (`internal/tournament/store.go`)
Uses `modernc.org/sqlite`, a pure Go implementation requiring no CGO compiler dependencies:
- Stores datasets, test cases, model definitions, match rounds, judge verdicts, and human votes.
- Enables single-file portability with WAL mode for concurrent reading during live tournament runs.

### 6. Embedded React 19 Studio (`internal/server` & `web/`)
Compiles to static assets embedded inside the Go executable via `go:embed`:
- Server-Sent Events (SSE) broadcast round completions in real time.
- Blind voting interface displays anonymized candidate completions with reveal-on-vote mechanisms.
- Interactive visualizations display ELO rankings, 95% confidence intervals, win-rate heatmaps, and Pareto frontier scatter charts.
