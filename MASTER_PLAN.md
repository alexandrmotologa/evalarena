# evalarena: Architecture and Engineering Master Plan
> Self-Hosted Pairwise Blind Model Preference Tournament and Bradley-Terry ELO Benchmarking Platform
> Target: Single-Binary Go Engine (< 25MB RAM) with Embedded React 19 Arena Studio & Pareto Efficiency Frontier

---

## 1. Problem Statement and Motivation

### 1.1 The Pitfalls of Generic Public AI Benchmarks
Engineering leaders and Head of AI teams frequently evaluate foundation models using public leaderboards (such as MMLU, GSM8K, HumanEval, or LMSYS Chatbot Arena). In enterprise environments, this approach leads to flawed architectural decisions:
1. **Benchmark Data Contamination**: Commercial LLMs increasingly memorize public evaluation splits during pre-training, giving inflated impressions of general reasoning.
2. **Enterprise Domain Mismatch**: High performance on academic multiple-choice quizzes has near-zero correlation with specialized enterprise tasks (such as adhering to strict company refund policies, emitting proprietary JSON schemas, or synthesizing complex SQL joins).
3. **Subjective Manual Guesswork**: Without empirical internal head-to-head testing, teams choose expensive default models ($15 to $30 per 1M tokens) when a finely tuned or quantized alternative ($0.15 to $0.60 per 1M tokens) achieves identical task accuracy.
4. **LLM Judge Position Bias**: Naive automated evaluators prefer the first response presented in the prompt (position bias), corrupting benchmark validity.

### 1.2 The evalarena Solution
`evalarena` brings the rigor of blind pairwise competition inside the enterprise firewall. Built as a high-performance, single-binary Go engine with an embedded React 19 tournament studio:
- **Proprietary Dataset Ingestion**: Ingests internal prompts, customer interactions, and domain tasks from JSONL, CSV, and YAML datasets.
- **Blind Pairwise Tournaments**: Dispatches prompt batches simultaneously across competing candidate models (OpenAI, Anthropic, Gemini, Mistral, Ollama, vLLM, and built-in offline simulation engines).
- **Position-Bias Mitigated Automated Judges**: Implements double-blind LLM judge evaluations where candidate positions (A/B vs. B/A) are mirrored and verified for consistency before awarding win, loss, or tie outcomes.
- **Collaborative Human Arena**: Provides an internal web UI where company domain experts cast blind votes on anonymized side-by-side completions.
- **Bradley-Terry ELO Rating Engine**: Continuously updates statistical skill ratings with 95% bootstrap confidence intervals.
- **Pareto Efficiency Frontier**: Plots ELO rating against inference cost (USD per 1M tokens) and latency (TTFT), identifying the exact mathematical trade-off for deployment.
- **CI/CD Quality Gates**: Automated regression prevention (`evalarena check`) blocking pull requests if a new model or prompt drops below baseline ELO.

---

## 2. High-Level System Architecture

```
       +-------------------------------------------------------------+
       |             Internal Enterprise Benchmark Dataset           |
       |     (Proprietary Prompts, Evaluation Rubrics, JSONL/CSV)    |
       +-------------------------------------------------------------+
                                      |
                                      v
       +-------------------------------------------------------------+
       |                     evalarena Core Engine                   |
       |                   (High-Performance Go CLI)                |
       |                                                             |
       |  +---------------------+        +------------------------+  |
       |  | Dataset Ingestion   |        | Pairwise Tournament    |  |
       |  | & Prompt Dispatcher |--+     | Matchmaker (Swiss/RR)  |  |
       |  +---------------------+  |     +------------------------+  |
       |                           |                  |              |
       |                           +------------------+              |
       |                                              |              |
       |                      Parallel Generation     |              |
       |                      v                       v              |
       |               +-------------+         +-------------+       |
       |               |   Model A   |         |   Model B   |       |
       |               +-------------+         +-------------+       |
       |                      \                       /              |
       |                       +---------------------+               |
       |                                  |                          |
       |                                  v                          |
       |  +-------------------------------------------------------+  |
       |  |              Dual Evaluation Battleground             |  |
       |  |                                                       |  |
       |  |   [A: Position-Mitigated Judge]   [B: Human Blind UI] |  |
       |  |   Swap order (A/B vs B/A)         Anonymized Voting   |  |
       |  +-------------------------------------------------------+  |
       |                                  |                          |
       |                                  v                          |
       |  +-------------------------------------------------------+  |
       |  |             Bradley-Terry Statistical Engine          |  |
       |  |   - ELO Rating Calculation & Win-Rate Matrix          |  |
       |  |   - Bootstrap 95% Confidence Intervals (1000 iter)    |  |
       |  |   - Pareto Frontier Analysis (ELO vs Cost & Latency)  |  |
       |  +-------------------------------------------------------+  |
       |                                                             |
       +------------------------------|------------------------------+
                                      |
                                      v
       +-------------------------------------------------------------+
       |                 Output & Decision Interfaces                |
       |  1. Embedded React 19 Arena Studio (Local HTTP :7070)       |
       |  2. Interactive Pareto Frontier Chart & ELO Leaderboard     |
       |  3. Automated Model Selection Decision Report (Markdown/PDF)|
       |  4. CI/CD Gate CLI (`evalarena check`)                       |
       +-------------------------------------------------------------+
```

---

## 3. Mathematical Formulation: Bradley-Terry ELO Model

### 3.1 Probability of Model i Defeating Model j
Under the Bradley-Terry preference model, each model $i$ possesses a latent skill rating $\beta_i \in \mathbb{R}$. The probability that Model $i$ wins over Model $j$ in a pairwise comparison is:

$$P(i \succ j) = \frac{e^{\beta_i}}{e^{\beta_i} + e^{\beta_j}} = \frac{1}{1 + 10^{(\text{ELO}_j - \text{ELO}_i)/400}}$$

Where $\text{ELO}_i = 1000 + 400 \cdot \frac{\beta_i}{\ln(10)}$.

### 3.2 Maximum Likelihood Estimation (MLE)
Given win counts $W_{ij}$ representing the number of times Model $i$ defeated Model $j$, and ties counted as 0.5 wins for each, the log-likelihood function maximized over all comparisons is:

$$\ln \mathcal{L}(\boldsymbol{\beta}) = \sum_{i < j} \left[ W_{ij} \ln P(i \succ j) + W_{ji} \ln P(j \succ i) \right]$$

The solver applies iterative Newton-Raphson optimization with gradient and Hessian evaluation. Regularization (Laplace smoothing or small Ridge prior) prevents divergence when a model has zero losses.

### 3.3 Bootstrap Resampling for Confidence Intervals
To avoid relying on asymptotic assumptions with small sample sizes, `evalarena` applies non-parametric bootstrap resampling ($N = 1,000$ iterations). For each iteration:
1. Resample matches uniformly with replacement from the observed match dataset.
2. Re-fit the Bradley-Terry MLE parameters $\boldsymbol{\beta}^{(b)}$.
3. Record the resulting ELO ratings.
4. Extract empirical 2.5th and 97.5th percentiles to establish the 95% confidence interval $[ELO_{2.5\%}, ELO_{97.5\%}]$.

### 3.4 Pareto Efficiency Frontier
A model configuration $M_k = (\text{ELO}_k, \text{Cost}_k, \text{Latency}_k)$ is strictly Pareto-optimal if no other configuration exists that achieves equal or higher ELO at strictly lower cost and latency:

$$\nexists M_m \text{ such that } (\text{ELO}_m \ge \text{ELO}_k \land \text{Cost}_m \le \text{Cost}_k \land \text{Latency}_m \le \text{Latency}_k) \land (\text{ELO}_m > \text{ELO}_k \lor \text{Cost}_m < \text{Cost}_k \lor \text{Latency}_m < \text{Latency}_k)$$

---

## 4. Position Bias Mitigation Algorithm

When an LLM evaluates two candidate completions, position bias causes it to systematically prefer candidate A (or candidate B) regardless of quality.

`evalarena` implements a dual-pass debiasing protocol:
1. **Pass 1 (Direct Order)**: Present candidate A as `Option A` and candidate B as `Option B`. Prompt the judge for structured JSON verdict $V_1 \in \{\text{Option A}, \text{Option B}, \text{Tie}\}$.
2. **Pass 2 (Swapped Order)**: Present candidate B as `Option A` and candidate A as `Option B`. Prompt the judge for structured JSON verdict $V_2 \in \{\text{Option A}, \text{Option B}, \text{Tie}\}$.
3. **Consistency Evaluation**:
   - If Pass 1 selects Model A and Pass 2 selects Model A (which was labeled Option B in Pass 2): **Model A Wins**.
   - If Pass 1 selects Model B and Pass 2 selects Model B: **Model B Wins**.
   - If Pass 1 selects Tie and Pass 2 selects Tie: **Tie**.
   - If Pass 1 and Pass 2 contradict (e.g., whichever model was presented in the first slot was declared winner): **Position Bias Inconsistency Detected**. The match is recorded as a Tie with a position-bias flag, or optionally re-evaluated with a neutral rubric.

---

## 5. Repository Structure

```
evalarena/
├── cmd/
│   └── evalarena/
│       └── main.go                 # CLI entry point (run, arena, serve, report, check, init)
├── internal/
│   ├── config/
│   │   ├── config.go               # Configuration parser for models, providers, rubrics
│   │   └── schema.go               # Types for matchmaking, providers, budget thresholds
│   ├── provider/
│   │   ├── client.go               # Universal LLM provider interface
│   │   ├── openai.go               # OpenAI API provider
│   │   ├── anthropic.go            # Anthropic Claude API provider
│   │   ├── gemini.go               # Google Gemini API provider
│   │   ├── ollama.go               # Ollama / vLLM local API provider
│   │   ├── mock.go                 # Deterministic simulation provider for testing & CI
│   │   ├── pricing.go              # Pricing catalog per 1M tokens (prompt and completion)
│   │   ├── metrics.go              # TTFT, generation latency, and token cost tracking
│   │   └── pool.go                 # Worker pool with per-model rate limiting & concurrency
│   ├── dataset/
│   │   ├── loader.go               # JSONL, CSV, and YAML dataset reader
│   │   └── validator.go            # Schema validation for test cases and evaluation criteria
│   ├── tournament/
│   │   ├── matchmaker.go           # Round-Robin, Swiss, and Random pairing schedulers
│   │   ├── runner.go               # Execution coordinator for parallel pairwise battles
│   │   ├── session.go              # Tournament state management
│   │   └── store.go                # Pure Go SQLite database layer (modernc.org/sqlite)
│   ├── judge/
│   │   ├── llm_judge.go            # LLM-as-a-judge execution and prompt templating
│   │   ├── debias.go               # Position-bias mitigation with order-swapping
│   │   └── rubric.go               # Domain rubric evaluation (Accuracy, Style, Formatting)
│   ├── elo/
│   │   ├── bradley_terry.go        # Bradley-Terry MLE solver using Newton-Raphson
│   │   ├── bootstrap.go            # Non-parametric bootstrap resampling (1,000 iterations)
│   │   ├── pareto.go               # Pareto efficiency frontier computation
│   │   └── matrix.go               # Pairwise head-to-head win-rate matrix
│   ├── report/
│   │   ├── terminal.go             # Terminal UI with ASCII tables and Pareto summary
│   │   ├── markdown.go             # Executive decision report generator
│   │   └── export.go               # JSON/CSV exporter for CI/CD metrics dashboards
│   └── server/
│       ├── api.go                  # REST API for human voting, leaderboard, tournament controls
│       ├── embed.go                # Embedded React 19 arena studio (go:embed)
│       └── sse.go                  # Server-Sent Events broker for live tournament rounds
├── web/
│   ├── src/
│   │   ├── App.tsx                 # React 19 tournament studio application
│   │   ├── api/
│   │   │   └── client.ts           # Typed API client for evalarena backend & SSE
│   │   ├── components/
│   │   │   ├── LeaderboardTable.tsx # Live ELO rankings with 95% bootstrap CI & Pareto tag
│   │   │   ├── ParetoFrontier.tsx  # Interactive ELO vs Cost ($/1M tokens) scatter plot
│   │   │   ├── WinRateHeatmap.tsx  # Pairwise head-to-head win percentage matrix
│   │   │   ├── BlindArenaVote.tsx  # Side-by-side anonymized human voting interface
│   │   │   ├── MatchInspector.tsx  # Detailed prompt-by-prompt comparative inspector
│   │   │   └── TournamentRunner.tsx# Direct tournament execution control panel
│   │   ├── pages/
│   │   │   ├── Dashboard.tsx       # Executive overview & Pareto efficiency view
│   │   │   ├── ArenaPage.tsx       # Human blind voting portal
│   │   │   ├── MatchesPage.tsx     # Full match inspection list
│   │   │   └── SettingsPage.tsx    # Config and model management
│   │   ├── index.css               # Tailwind CSS styles
│   │   └── main.tsx                # React entry point
│   ├── package.json
│   ├── tsconfig.json
│   └── vite.config.ts
├── examples/
│   ├── evalarena.yaml              # Complete tournament configuration example
│   ├── dataset.jsonl               # Realistic enterprise customer support evaluation dataset
│   └── custom_rubric.md            # Domain-specific evaluation rubric
├── docs/
│   ├── ARCHITECTURE.md             # System architecture, memory model, and concurrency
│   ├── BRADLEY_TERRY_MATHEMATICS.md# Mathematical derivation, MLE solver, and bootstrap CI
│   ├── POSITION_BIAS_MITIGATION.md # Bias mitigation protocol & empirical findings
│   ├── CONFIGURATION_GUIDE.md      # Configuration parameters reference
│   └── API_SPECIFICATION.md        # REST and SSE API endpoints
├── .github/
│   └── workflows/
│       └── ci.yml                  # GitHub Actions CI workflow (lint, test, build)
├── Makefile                        # Build, test, and run automation
├── go.mod
├── go.sum
├── LICENSE                         # MIT License
└── README.md                       # Humanized project documentation
```

---

## 6. Implementation Progress Tracker

### Phase 1: Project Setup, Tooling & Documentation Foundation
- [x] 1.1 Initialize Git repository, configure GitHub remote, set topics and description
- [x] 1.2 Setup Go module (`github.com/alexandrmotologa/evalarena`) with `go 1.24+`
- [x] 1.3 Add `.gitignore`, `LICENSE`, and `Makefile`
- [x] 1.4 Write comprehensive documentation in `docs/` (Architecture, Bradley-Terry Math, Position Bias, Config Guide, API Spec)
- [x] 1.5 Write clean humanized `README.md` (no fake images, factual, clear CLI guide)

### Phase 2: Configuration, Dataset Ingestion & SQLite Storage
- [x] 2.1 Implement `internal/config`: YAML configuration parser, validation, pricing registry
- [x] 2.2 Implement `internal/dataset`: JSONL, CSV, and YAML prompt loader with schema validation
- [x] 2.3 Implement `internal/tournament/store.go`: Pure Go embedded SQLite (`modernc.org/sqlite`) with automatic schema creation and migrations

### Phase 3: Universal Provider Engine with Metrics & Simulation
- [x] 3.1 Implement `internal/provider/client.go`: Provider interface, completion requests, response schemas
- [x] 3.2 Implement `internal/provider/pricing.go`: Per-token cost calculator with default pricing for OpenAI, Anthropic, Gemini, Mistral, and local models
- [x] 3.3 Implement `internal/provider/mock.go`: Deterministic simulation provider for offline tests, CI/CD, and demonstrations
- [x] 3.4 Implement real providers: OpenAI (`openai.go`), Anthropic (`anthropic.go`), Gemini (`gemini.go`), and Ollama/vLLM (`ollama.go`)
- [x] 3.5 Implement `internal/provider/pool.go`: Concurrency manager with rate limiting and retry backoff

### Phase 4: Position-Debiased LLM-as-a-Judge Engine
- [x] 4.1 Implement `internal/judge/rubric.go`: Scoring rubrics, evaluation criteria, and system prompt formatting
- [x] 4.2 Implement `internal/judge/llm_judge.go`: LLM evaluation harness with JSON structured outputs
- [x] 4.3 Implement `internal/judge/debias.go`: Order-swapping protocol (A/B and B/A), consistency verification, and tie classification

### Phase 5: Tournament Matchmaker & Execution Pipeline
- [x] 5.1 Implement `internal/tournament/matchmaker.go`: Round-Robin, Swiss, and Random pairing algorithms
- [x] 5.2 Implement `internal/tournament/runner.go`: Multi-model async execution, response collection, dual-pass judge submission, database persistence
- [x] 5.3 Implement SSE event dispatcher for real-time progress broadcast

### Phase 6: Bradley-Terry ELO Solver & Pareto Efficiency Engine
- [x] 6.1 Implement `internal/elo/bradley_terry.go`: Maximum Likelihood Estimation solver with Newton-Raphson iteration
- [x] 6.2 Implement `internal/elo/bootstrap.go`: Non-parametric bootstrap resampling (1,000 iterations) for 95% confidence intervals
- [x] 6.3 Implement `internal/elo/matrix.go`: Pairwise head-to-head win-loss-tie matrix calculation
- [x] 6.4 Implement `internal/elo/pareto.go`: Pareto efficiency frontier calculation identifying cost-efficient models

### Phase 7: REST API, Server & Embedded React 19 Studio
- [x] 7.1 Implement `internal/server/api.go`: REST API endpoints for models, battles, leaderboard, human voting, and tournament triggers
- [x] 7.2 Implement `internal/server/sse.go`: Real-time SSE streaming endpoint for live tournament rounds
- [x] 7.3 Build React 19 Arena Studio in `web/` (Tailwind CSS, Lucide icons, Leaderboard with 95% CI, Pareto Frontier plot, Blind Arena Vote, Win-Rate Matrix, Match Inspector)
- [x] 7.4 Embed static build into Go binary (`//go:embed all:dist`) with single-binary HTTP server

### Phase 8: CLI Commands, CI/CD Quality Gate & Executive Reports
- [x] 8.1 Implement CLI commands in `cmd/evalarena/main.go`:
  - `evalarena run`: Run tournament from config file
  - `evalarena arena`: Start embedded web studio
  - `evalarena report`: Generate terminal, markdown, JSON, or CSV report
  - `evalarena check`: Automated CI/CD regression quality gate (`--fail-if-below`)
  - `evalarena init`: Scaffold starter config and dataset
- [x] 8.2 Implement `internal/report`: Terminal colored output and Markdown executive report generator
- [x] 8.3 Setup GitHub Actions CI workflow (`.github/workflows/ci.yml`)

### Phase 9: Verification, Test Suite & Mathematical Validation
- [x] 9.1 Unit tests for Bradley-Terry math solver against known reference datasets
- [x] 9.2 Unit tests for bootstrap confidence intervals and Pareto dominance calculations
- [x] 9.3 Unit tests for position bias mitigation and order swapping
- [x] 9.4 End-to-end tournament run verification with simulation models
- [x] 9.5 Memory benchmark verification confirming < 25MB RAM footprint (measured 15.87 MB)
- [x] 9.6 Commit and push to GitHub repository
