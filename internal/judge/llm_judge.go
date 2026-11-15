package judge

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/alexandrmotologa/evalarena/internal/provider"
)

// SingleVerdict contains the output of one evaluation pass.
type SingleVerdict struct {
	Winner    string `json:"winner"` // "A", "B", or "tie"
	Reasoning string `json:"reasoning"`
}

// EvaluateSinglePass executes one judging step given Option A and Option B.
func EvaluateSinglePass(ctx context.Context, client provider.Client, userPrompt, optionA, optionB, rubric string, temp float64) (*SingleVerdict, error) {
	judgePrompt := BuildJudgePrompt(userPrompt, optionA, optionB)
	systemPrompt := BuildJudgeSystemPrompt(rubric)

	req := provider.CompletionRequest{
		Prompt:       judgePrompt,
		SystemPrompt: systemPrompt,
		Temperature:  temp,
		MaxTokens:    1024,
	}

	resp, err := client.Generate(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("judge completion failed: %w", err)
	}

	verdict, err := parseJudgeJSON(resp.Text)
	if err != nil {
		return nil, fmt.Errorf("failed to parse judge verdict: %w (raw response: %s)", err, resp.Text)
	}

	return verdict, nil
}

func parseJudgeJSON(raw string) (*SingleVerdict, error) {
	cleaned := strings.TrimSpace(raw)

	// Strip markdown code block wrappers if present (e.g. ```json ... ```)
	if strings.HasPrefix(cleaned, "```") {
		lines := strings.Split(cleaned, "\n")
		if len(lines) >= 2 {
			if strings.HasPrefix(lines[0], "```") {
				lines = lines[1:]
			}
			if len(lines) > 0 && strings.HasPrefix(lines[len(lines)-1], "```") {
				lines = lines[:len(lines)-1]
			}
			cleaned = strings.TrimSpace(strings.Join(lines, "\n"))
		}
	}

	// First attempt direct JSON unmarshal
	var v SingleVerdict
	if err := json.Unmarshal([]byte(cleaned), &v); err == nil {
		v.Winner = normalizeWinner(v.Winner)
		return &v, nil
	}

	// Regex fallback to extract JSON object
	re := regexp.MustCompile(`\{[^{}]*"winner"[^{}]*\}`)
	match := re.FindString(cleaned)
	if match != "" {
		if err := json.Unmarshal([]byte(match), &v); err == nil {
			v.Winner = normalizeWinner(v.Winner)
			return &v, nil
		}
	}

	// Heuristic fallback if json structure was broken
	lower := strings.ToLower(cleaned)
	if strings.Contains(lower, `"winner": "a"`) || strings.Contains(lower, `"winner":"a"`) || strings.Contains(lower, "winner: a") {
		return &SingleVerdict{Winner: "A", Reasoning: cleaned}, nil
	}
	if strings.Contains(lower, `"winner": "b"`) || strings.Contains(lower, `"winner":"b"`) || strings.Contains(lower, "winner: b") {
		return &SingleVerdict{Winner: "B", Reasoning: cleaned}, nil
	}
	if strings.Contains(lower, "tie") || strings.Contains(lower, "equal") {
		return &SingleVerdict{Winner: "tie", Reasoning: cleaned}, nil
	}

	return nil, fmt.Errorf("could not extract structured winner from response: %q", raw)
}

func normalizeWinner(w string) string {
	w = strings.ToUpper(strings.TrimSpace(w))
	switch w {
	case "A", "OPTION A", "RESPONSE A", "MODEL A":
		return "A"
	case "B", "OPTION B", "RESPONSE B", "MODEL B":
		return "B"
	default:
		return "tie"
	}
}
