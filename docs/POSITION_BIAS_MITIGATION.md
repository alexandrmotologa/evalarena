# Position Bias Mitigation Protocol

Automated model-based evaluators (LLM-as-a-judge) frequently display position bias: when presented with two completions, the judge systematically favors the candidate shown in the first position (Option A), regardless of substance.

evalarena eliminates this distortion through a deterministic dual-pass evaluation protocol.

## Protocol Mechanics

For every pairwise comparison between Candidate $M_1$ and Candidate $M_2$ answering prompt $Q$:

### Step 1: Forward Pass (Order A/B)
The prompt and completions are structured as:
```
[PROMPT]
{Q}

[RESPONSE A]
{Completion from M_1}

[RESPONSE B]
{Completion from M_2}
```
The judge evaluates both completions against the configured rubric and outputs a structured JSON judgment:
```json
{
  "reasoning": "Detailed justification...",
  "winner": "A" | "B" | "tie"
}
```
Let $V_1$ denote the winner mapped back to the candidate model ID:
- If `winner == "A"`: $V_1 = M_1$
- If `winner == "B"`: $V_1 = M_2$
- If `winner == "tie"`: $V_1 = \text{TIE}$

### Step 2: Inverted Pass (Order B/A)
The same completions are presented with reversed ordering:
```
[PROMPT]
{Q}

[RESPONSE A]
{Completion from M_2}

[RESPONSE B]
{Completion from M_1}
```
The judge re-evaluates independently. Let $V_2$ denote the resolved winner:
- If `winner == "A"`: $V_2 = M_2$
- If `winner == "B"`: $V_2 = M_1$
- If `winner == "tie"`: $V_2 = \text{TIE}$

### Step 3: Consistency Resolution

The final match outcome is resolved by comparing $V_1$ and $V_2$:

| Forward Verdict $V_1$ | Inverted Verdict $V_2$ | Final Tournament Outcome | Position Bias Detected |
|:----------------------|:-----------------------|:-------------------------|:-----------------------|
| $M_1$                 | $M_1$                  | **$M_1$ Wins**           | No                     |
| $M_2$                 | $M_2$                  | **$M_2$ Wins**           | No                     |
| $\text{TIE}$          | $\text{TIE}$           | **Tie**                  | No                     |
| $M_1$ (Option A)      | $M_2$ (Option A)       | **Tie (Position Bias)**  | **Yes (First-Order)**  |
| $M_2$ (Option B)      | $M_1$ (Option B)       | **Tie (Position Bias)**  | **Yes (Second-Order)** |
| $M_1$                 | $\text{TIE}$           | **Tie (Inconclusive)**   | Soft Inconsistency     |
| $\text{TIE}$          | $M_2$                  | **Tie (Inconclusive)**   | Soft Inconsistency     |

When an inconsistency is detected, evalarena logs the event with the judge reasoning from both passes, records the match as a tie to protect rating integrity, and increments the judge's position bias diagnostic metric.
