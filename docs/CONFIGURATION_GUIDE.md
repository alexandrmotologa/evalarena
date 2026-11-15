# evalarena Configuration Guide

evalarena tournaments are configured through YAML files. This document details all available configuration blocks.

## Sample Configuration File

```yaml
version: "1.0"
tournament:
  name: "Customer Support Policy Adherence"
  schedule: "round-robin" # options: "round-robin", "swiss", "random"
  rounds: 3
  concurrency: 4
  database: "evalarena.db"

judge:
  provider: "openai"
  model: "gpt-4o"
  temperature: 0.1
  debias_position: true # enables dual-pass order swapping
  rubric_file: "examples/custom_rubric.md"

models:
  - id: "gpt-4o"
    provider: "openai"
    model: "gpt-4o"
    pricing:
      input_per_million: 2.50
      output_per_million: 10.00

  - id: "claude-3-5-sonnet"
    provider: "anthropic"
    model: "claude-3-5-sonnet-20241022"
    pricing:
      input_per_million: 3.00
      output_per_million: 15.00

  - id: "gemini-2.0-flash"
    provider: "gemini"
    model: "gemini-2.0-flash"
    pricing:
      input_per_million: 0.10
      output_per_million: 0.40

  - id: "llama-3.1-8b-local"
    provider: "ollama"
    model: "llama3.1:8b"
    endpoint: "http://localhost:11434"
    pricing:
      input_per_million: 0.05
      output_per_million: 0.05

dataset:
  path: "examples/dataset.jsonl"
  format: "jsonl" # options: "jsonl", "csv"

gate:
  baseline_model: "gpt-4o"
  candidate_model: "gemini-2.0-flash"
  max_elo_drop: 30.0
  max_cost_ratio: 0.50
```

## Field Descriptions

### `tournament`
- `name`: Display title for the evaluation run.
- `schedule`: Tournament match scheduling algorithm.
  - `round-robin`: Every model faces every other model on all dataset test cases.
  - `swiss`: Models with similar running scores face each other to refine rankings.
  - `random`: Uniformly selects pairs for broad sampling.
- `rounds`: Repetitions per pair when running Swiss or random pairings.
- `concurrency`: Maximum concurrent API requests across providers.
- `database`: SQLite file path for persistence.

### `judge`
- `provider`: Model provider for the automated judge (`openai`, `anthropic`, `gemini`, `ollama`, or `mock`).
- `model`: Model identifier used for judging.
- `temperature`: Sampling temperature (recommended 0.0 to 0.2 for reproducible scoring).
- `debias_position`: When true, runs two passes with reversed completion order to cancel position bias.
- `rubric_file`: Path to Markdown evaluation guidelines.

### `models`
A list of model definitions participating in the tournament. Each entry contains:
- `id`: Unique identifier used in leaderboards and reports.
- `provider`: One of `openai`, `anthropic`, `gemini`, `ollama`, or `mock`.
- `model`: Target provider model string.
- `endpoint`: Optional base URL (for Ollama or vLLM instances).
- `pricing`: Token pricing in USD per 1,000,000 tokens for input (prompt) and output (completion).
