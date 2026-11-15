# evalarena

Self-hosted blind pairwise model preference tournament and Bradley-Terry ELO benchmarking platform for large language models, implemented in Go with an embedded React 19 arena studio.

Public benchmarks such as MMLU or Chatbot Arena measure broad academic knowledge and consumer chat preferences. In enterprise environments, selecting foundation models requires testing against proprietary datasets, internal policy rubrics, and strict budget constraints. 

evalarena runs automated and human-evaluated pairwise tournaments directly within your infrastructure. It eliminates position bias through dual-pass order swapping, solves latent model ratings with the Bradley-Terry Maximum Likelihood Estimation (MLE) algorithm, computes 95% bootstrap confidence intervals, and isolates the Pareto Efficiency Frontier balancing quality, latency, and token cost.

## Key Capabilities

- **Position-Bias Mitigated LLM Judge**: Evaluates every pair in forward (A/B) and inverted (B/A) order. Inconsistent judgments caused by first-position preference are detected and flagged.
- **Collaborative Human Arena**: Zero-friction web interface for team members to cast blind votes on anonymized completions, with model reveal upon vote submission.
- **Bradley-Terry ELO Rating Engine**: Iterative Newton-Raphson solver for maximum likelihood skill ratings, coupled with 1,000-iteration non-parametric bootstrap resampling for empirical 95% confidence intervals.
- **Pareto Efficiency Frontier**: Mathematical multi-objective optimization identifying models that achieve the highest ELO rating for any given cost or latency budget.
- **Universal Provider Client**: Supports OpenAI, Anthropic, Gemini, Mistral, Ollama/vLLM endpoints, and a deterministic offline simulation provider for CI/CD testing.
- **Single-Binary Portability**: Pure Go architecture with zero CGO dependencies, embedded SQLite storage, and embedded React 19 studio, consuming under 25MB RAM.
- **CI/CD Quality Gates**: Automated regression prevention (`evalarena check`) that blocks deployment when a candidate model scores below a baseline threshold.

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
|  |     - Maximum Likelihood Estimation (Newton-Raphson)            |  |
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
|  4. CI/CD Gate CLI (`evalarena check --fail-if-below`)                 |
+-----------------------------------------------------------------------+
```

## Quick Start

### 1. Installation

Build from source (requires Go 1.24+ and Node.js 20+):

```bash
git clone https://github.com/alexandrmotologa/evalarena.git
cd evalarena
make ui-build
make build
```

The compiled binary will be placed at `bin/evalarena`.

### 2. Run a Tournament

Run a pairwise benchmark tournament using the sample configuration and dataset:

```bash
./bin/evalarena run --config examples/evalarena.yaml
```

Output:
```
========================================================================================
                       evalarena: Tournament Leaderboard
========================================================================================
 RANK | MODEL                 | ELO   | 95% CI          | W/L/T    | WIN%   | $/1M TOK | PARETO
------+-----------------------+-------+-----------------+----------+--------+----------+-------
    1 | claude-3-5-sonnet     | 1142  | [1118 - 1168]   | 18/6/2   | 73.1%  | $4.80    | Yes
    2 | gpt-4o                | 1120  | [1096 - 1145]   | 16/8/2   | 65.4%  | $4.38    | No
    3 | gemini-2.0-flash      | 1045  | [1021 - 1070]   | 11/13/2  | 46.2%  | $0.18    | Yes
    4 | llama-3.1-8b-instruct |  893  | [869 - 918]     | 3/21/2   | 15.4%  | $0.06    | Yes
========================================================================================
Pareto Optimal Set: [claude-3-5-sonnet, gemini-2.0-flash, llama-3.1-8b-instruct]
Production Recommendation: gemini-2.0-flash delivers 91.5% of top ELO at 3.7% of the cost.
```

### 3. Launch the Embedded Arena Studio

To launch the web studio for human blind voting and interactive Pareto exploration:

```bash
./bin/evalarena arena --port 7070
```

Open your browser at `http://localhost:7070`.

### 4. CI/CD Regression Quality Gate

Verify whether a candidate model meets your minimum ELO standard relative to production:

```bash
./bin/evalarena check \
  --config examples/evalarena.yaml \
  --candidate gemini-2.0-flash \
  --baseline gpt-4o \
  --max-elo-drop 50.0
```

If the candidate's ELO falls below `baseline_elo - max_elo_drop`, the command exits with code 1, halting pipeline execution.

## Mathematical Formulation

### Bradley-Terry Preference Probability
The probability that Model $i$ wins over Model $j$ is:

$$P(i \succ j) = \frac{e^{\beta_i}}{e^{\beta_i} + e^{\beta_j}} = \frac{1}{1 + 10^{(\text{ELO}_j - \text{ELO}_i)/400}}$$

where $\text{ELO}_i = 1000 + \frac{400}{\ln(10)} \cdot \beta_i$.

### Position Bias Mitigation
To counter position bias in LLM evaluators, evalarena evaluates every match in forward ($A/B$) and reversed ($B/A$) order:
- Consistent verdicts yield an authoritative win.
- Order-dependent contradictions (where the first option is always selected) are recorded as ties, mitigating systematic distortion.

## Documentation

- [System Architecture](docs/ARCHITECTURE.md)
- [Bradley-Terry Mathematics and Bootstrap Resampling](docs/BRADLEY_TERRY_MATHEMATICS.md)
- [Position Bias Mitigation Protocol](docs/POSITION_BIAS_MITIGATION.md)
- [Configuration Reference](docs/CONFIGURATION_GUIDE.md)
- [REST and SSE API Specification](docs/API_SPECIFICATION.md)

## License

MIT License. See [LICENSE](LICENSE) for details.
