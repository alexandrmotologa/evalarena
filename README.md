<p align="center">
  <img src="docs/images/logo.png?raw=true" alt="evalarena Logo" width="130" style="border-radius: 24px;" />
</p>

<h1 align="center">evalarena</h1>

<p align="center">
  Self-Hosted Pairwise Blind Model Preference Tournament and Bradley-Terry ELO Benchmarking Platform
</p>

<p align="center">
  <img src="https://img.shields.io/badge/go-1.24-00ADD8?logo=go&logoColor=white" alt="Go Version" />
  <img src="https://img.shields.io/badge/react-19.0-61DAFB?logo=react&logoColor=black" alt="React 19" />
  <img src="https://img.shields.io/badge/sqlite-embedded-003B57?logo=sqlite&logoColor=white" alt="SQLite" />
  <img src="https://img.shields.io/badge/memory-%3C%2025MB%20RAM-success" alt="Memory Footprint" />
  <img src="https://img.shields.io/badge/license-MIT-blue.svg" alt="License" />
</p>

<p align="center">
  <img src="docs/images/evalarena_demo.gif?raw=true" alt="evalarena Interactive Studio Demo" width="100%" style="border-radius: 12px; box-shadow: 0 12px 36px rgba(0,0,0,0.4);" />
</p>

---

## Overview

Public benchmarks such as MMLU, GSM8K, or LMSYS Chatbot Arena measure general model knowledge and consumer chat preference. In enterprise environments, selecting foundation models requires testing against proprietary task datasets, internal compliance rubrics, and concrete latency or budget limits.

**evalarena** runs automated and human-evaluated pairwise tournaments directly within your infrastructure. It eliminates position bias through dual-pass order swapping, solves latent model ratings with the Bradley-Terry Maximum Likelihood Estimation (MLE) algorithm, computes 95% non-parametric bootstrap confidence intervals, and isolates the Pareto Efficiency Frontier balancing quality, latency, and token expenditure.

The system compiles into a single standalone binary with zero external runtime dependencies, an embedded SQLite persistence engine, and an embedded React 19 web studio operating under 25MB RAM.

---

## Visual Tour

### 1. Pareto Efficiency Frontier and Model Optimization
Identifies non-dominated models that maximize ELO rating while minimizing blended token costs ($/1M tokens) and response latency. Generates automated production deployment recommendations.

<p align="center">
  <img src="docs/images/evalarena_dashboard.png?raw=true" alt="evalarena Pareto Frontier and Dashboard" width="100%" style="border-radius: 10px;" />
</p>

### 2. Bradley-Terry Leaderboard and Judge Calibration
Displays latent skill ratings solved via iterative Minorization-Maximization (Hunter 2004) with visual 95% bootstrap confidence interval bar tracks. Features Cohen's Kappa ($\kappa$) inter-annotator scorecard measuring agreement between internal team votes and automated LLM judge verdicts.

<p align="center">
  <img src="docs/images/evalarena_leaderboard.png?raw=true" alt="evalarena Bradley-Terry Leaderboard and Calibration" width="100%" style="border-radius: 10px;" />
</p>

### 3. Collaborative Blind Voting Arena
Presents side-by-side anonymized completions for enterprise prompts. Team members vote on prompt adherence without model identity cues. Features inline word-level diff toggles and post-vote identity reveals.

<p align="center">
  <img src="docs/images/evalarena_blind_arena.png?raw=true" alt="evalarena Blind Voting Arena" width="100%" style="border-radius: 10px;" />
</p>

### 4. Ad-Hoc Quick Battle Playground
Interactive testing suite allowing engineers to pit any two foundation models against arbitrary test prompts. Automatically conducts a dual-pass debiasing audit (Forward A/B and Swapped B/A) and displays pass-by-pass reasoning alongside token metrics.

<p align="center">
  <img src="docs/images/evalarena_quick_battle.png?raw=true" alt="evalarena Ad-Hoc Quick Battle" width="100%" style="border-radius: 10px;" />
</p>

### 5. Match Inspector and Position Bias Audit
Provides full auditing for every pairwise match recorded in SQLite. Engineers can inspect prompt text, latency, token consumption, cost calculations, word diffs, and position bias flags.

<p align="center">
  <img src="docs/images/evalarena_inspector.png?raw=true" alt="evalarena Match Inspector" width="100%" style="border-radius: 10px;" />
</p>

### 6. Pairwise Head-to-Head Win-Rate Matrix
Color-coded $N \times N$ matrix indicating the historical win percentage of every model pair, verifying transitivity and identifying specific model counter-matchups.

<p align="center">
  <img src="docs/images/evalarena_win_matrix.png?raw=true" alt="evalarena Pairwise Win Rate Matrix" width="100%" style="border-radius: 10px;" />
</p>

### 7. Production CLI and CI/CD Gating
Executes batch tournament rounds and enforces regression gates in CI/CD pipelines (`evalarena check`) to prevent deployment of degraded or non-compliant models.

<p align="center">
  <img src="docs/images/evalarena_cli.png?raw=true" alt="evalarena CLI and CI/CD Quality Gate" width="100%" style="border-radius: 10px;" />
</p>

### 8. Full Device Responsiveness
Responsive UI layout adapting cleanly to mobile, tablet, and desktop viewports with horizontal touch navigation, collapsible match cards, and fluid SVG scaling.

<p align="center">
  <img src="docs/images/evalarena_mobile.png?raw=true" alt="evalarena Mobile Responsiveness" width="380" style="border-radius: 12px;" />
</p>

---

## Core Capabilities

- **Position-Bias Mitigated LLM Judge**: Evaluates every pair in forward (A/B) and inverted (B/A) order. Inconsistent judgments caused by first-position preference are detected and flagged.
- **Collaborative Human Arena**: Zero-friction web interface for team members to cast blind votes on anonymized completions, with model reveal upon vote submission.
- **Bradley-Terry ELO Rating Engine**: Iterative Minorization-Maximization solver for maximum likelihood skill ratings, coupled with 1,000-iteration non-parametric bootstrap resampling for empirical 95% confidence intervals.
- **Pareto Efficiency Frontier**: Mathematical multi-objective optimization identifying models that achieve the highest ELO rating for any given cost or latency budget.
- **Universal Provider Client**: Supports OpenAI, Anthropic, Gemini, Mistral, Ollama/vLLM endpoints, and a deterministic offline simulation provider for CI/CD testing.
- **Single-Binary Portability**: Pure Go architecture with zero CGO dependencies, embedded SQLite storage, and embedded React 19 studio, consuming under 25MB RAM.
- **CI/CD Quality Gates**: Automated regression prevention (`evalarena check`) that blocks deployment when a candidate model scores below a baseline threshold.

---

## Architecture

```
                  +-----------------------------------+
                  |   Internal Enterprise Datasets    |
                  |     (JSONL, CSV, Custom Rubric)   |
                  +-----------------------------------+
                                    |
                                    v
+-----------------------------------------------------------------------+
|                         evalarena Core Engine                         |
|                                                                       |
|  +--------------------+   +-------------------+   +----------------+  |
|  | Dataset Ingestion  |-->| Matchmaker        |-->| Worker Pool    |  |
|  | & Validator        |   | (Swiss / RR)      |   | (Rate-limited) |  |
|  +--------------------+   +-------------------+   +----------------+  |
|                                                           |           |
|                                                           v           |
|                                                    Parallel Queries   |
|                                                    [Model A, Model B] |
|                                                           |           |
|                                                           v           |
|  +-----------------------------------------------------------------+  |
|  |                  Dual Evaluation Battleground                   |  |
|  |                                                                 |  |
|  |     [Automated LLM Judge]             [Blind Human Arena]       |  |
|  |     Dual-pass order swap (A/B & B/A)  Anonymized side-by-side   |  |
|  |     Consistency verification          Live voting on :7070      |  |
|  +-----------------------------------------------------------------+  |
|                                   |                                   |
|                                   v                                   |
|  +-----------------------------------------------------------------+  |
|  |                Bradley-Terry Statistical Engine                 |  |
|  |     - Maximum Likelihood Estimation (Hunter MM Algorithm)       |  |
|  |     - Non-Parametric Bootstrap 95% Confidence Intervals         |  |
|  |     - Multi-Objective Pareto Frontier (ELO vs Cost vs Latency)  |  |
|  +-----------------------------------------------------------------+  |
|                                   |                                   |
+-----------------------------------|-----------------------------------+
                                    v
+-----------------------------------------------------------------------+
|                      Delivery and Decision Modes                      |
|                                                                       |
|  1. Embedded React 19 Arena Studio (http://localhost:7070)            |
|  2. Interactive Pareto Frontier Chart and Real-Time Leaderboard       |
|  3. Automated Executive Decision Reports (Markdown, Terminal, JSON)   |
|  4. CI/CD Gate CLI (`evalarena check --max-elo-drop`)                 |
+-----------------------------------------------------------------------+
```

---

## Quick Start

### 1. Installation

Build from source (requires Go 1.24+ and Node.js 20+):

```bash
git clone https://github.com/alexandrmotologa/evalarena.git
cd evalarena
npm --prefix web run build
go build -o bin/evalarena ./cmd/evalarena
```

The compiled binary will be placed at `bin/evalarena`.

### 2. Run a Tournament

Run a pairwise benchmark tournament using the sample configuration and dataset:

```bash
./bin/evalarena run --config examples/evalarena.yaml --rounds 2
```

Sample output:
```
Starting tournament "Enterprise LLM Pairwise Benchmark" with 4 models on 5 prompts (2 rounds)...
Completed 60 pairwise battles in 467ms across universal provider pool.

=== evalarena Leaderboard: Enterprise LLM Pairwise Benchmark ===
┌──────┬───────────────────────┬──────┬───────────────┬───────────┬───────┬────────────┬─────────┬───────────────┐
│ RANK │         MODEL         │ ELO  │   95 %  CI    │ W / L / T │ WIN % │ $/ 1 M TOK │ LATENCY │    PARETO     │
├──────┼───────────────────────┼──────┼───────────────┼───────────┼───────┼────────────┼─────────┼───────────────┤
│ ★ 1  │ gpt-4o                │ 1092 │ [1051 - 1138] │ 10/0/20   │ 66.7% │ $1.42      │ 30ms    │ Yes (Optimal) │
│ 2    │ claude-3-5-sonnet     │ 1000 │ [976 - 1023]  │ 0/0/30    │ 50.0% │ $2.12      │ 35ms    │ No            │
│ 3    │ gemini-2.0-flash      │ 1000 │ [974 - 1025]  │ 0/0/30    │ 50.0% │ $0.06      │ 15ms    │ Yes (Optimal) │
│ 4    │ llama-3.1-8b-instruct │ 908  │ [861 - 947]   │ 0/10/20   │ 33.3% │ $0.01      │ 10ms    │ Yes (Optimal) │
└──────┴───────────────────────┴──────┴───────────────┴───────────┴───────┴────────────┴─────────┴───────────────┘

--- Pareto Frontier Analysis & Recommendations ---
• gpt-4o: Top Quality Baseline (ELO 1092, $1.42/1M tokens, 30ms latency).
• gemini-2.0-flash: High-Efficiency Alternative: Delivers 91.6% of peak ELO while reducing cost by 96.1% (15ms).
• llama-3.1-8b-instruct: High-Efficiency Alternative: Delivers 83.2% of peak ELO while reducing cost by 99.1% (10ms).
```

### 3. Launch the Embedded Arena Studio

To launch the web studio for human blind voting and interactive Pareto exploration:

```bash
./bin/evalarena arena --port 7070 --config examples/evalarena.yaml
```

Open your browser at `http://localhost:7070`.

### 4. CI/CD Regression Quality Gate

Verify whether a candidate model meets your minimum ELO standard relative to production:

```bash
./bin/evalarena check \
  --config examples/evalarena.yaml \
  --candidate claude-3-5-sonnet \
  --baseline gpt-4o \
  --max-elo-drop 100.0
```

If the candidate ELO falls below `baseline_elo - max_elo_drop`, the command exits with code 1, halting pipeline execution.

---

## Mathematical Formulation

### Bradley-Terry Preference Probability
The probability that Model $i$ wins over Model $j$ is defined as:

$$P(i \succ j) = \frac{e^{\beta_i}}{e^{\beta_i} + e^{\beta_j}} = \frac{1}{1 + 10^{(\text{ELO}_j - \text{ELO}_i)/400}}$$

where $\text{ELO}_i = 1000 + \frac{400}{\ln(10)} \cdot \beta_i$.

### Position Bias Mitigation
To counter position bias in LLM evaluators, evalarena evaluates every match in forward ($A/B$) and reversed ($B/A$) order:
- Consistent verdicts yield an authoritative win.
- Order-dependent contradictions (where the first option is always selected) are recorded as ties, mitigating systematic distortion.

---

## Documentation

- [System Architecture](docs/ARCHITECTURE.md)
- [Bradley-Terry Mathematics and Bootstrap Resampling](docs/BRADLEY_TERRY_MATHEMATICS.md)
- [Position Bias Mitigation Protocol](docs/POSITION_BIAS_MITIGATION.md)
- [Configuration Reference](docs/CONFIGURATION_GUIDE.md)
- [REST and SSE API Specification](docs/API_SPECIFICATION.md)

---

## License

MIT License. See [LICENSE](LICENSE) for details.
