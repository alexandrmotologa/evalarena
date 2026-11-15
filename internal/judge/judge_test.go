package judge_test

import (
	"context"
	"testing"
	"time"

	"github.com/alexandrmotologa/evalarena/internal/judge"
	"github.com/alexandrmotologa/evalarena/internal/provider"
)

// mockJudgeClient simulates predefined judge verdicts
type testJudgeClient struct {
	verdicts []string
	callIdx  int
}

func (t *testJudgeClient) ID() string { return "test-judge" }

func (t *testJudgeClient) Generate(ctx context.Context, req provider.CompletionRequest) (*provider.CompletionResponse, error) {
	v := t.verdicts[t.callIdx%len(t.verdicts)]
	t.callIdx++
	return &provider.CompletionResponse{
		Text:      v,
		Latency:   10 * time.Millisecond,
		LatencyMs: 10,
	}, nil
}

func TestEvaluateDebiased_ConsistentWin(t *testing.T) {
	// Pass 1: A wins (slot A is model-1) -> model-1 wins
	// Pass 2: B wins (slot B is model-1) -> model-1 wins
	client := &testJudgeClient{
		verdicts: []string{
			`{"winner": "A", "reasoning": "Pass 1: Model 1 was better"}`,
			`{"winner": "B", "reasoning": "Pass 2: Model 1 (in slot B) was better"}`,
		},
	}

	ctx := context.Background()
	res, err := judge.EvaluateDebiased(ctx, client, "model-1", "model-2", "Test prompt", "Resp 1", "Resp 2", "", 0.1, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res.FinalWinner != "model-1" {
		t.Errorf("expected final winner model-1, got %s", res.FinalWinner)
	}
	if res.PositionBiasDetected {
		t.Errorf("expected no position bias detected")
	}
	if res.Inconsistent {
		t.Errorf("expected consistent evaluation")
	}
}

func TestEvaluateDebiased_PositionBiasDetected(t *testing.T) {
	// Judge suffers from first-position bias: always chooses slot A!
	// Pass 1: selects slot A (model-1)
	// Pass 2: selects slot A (model-2)
	client := &testJudgeClient{
		verdicts: []string{
			`{"winner": "A", "reasoning": "Pass 1 favors slot A"}`,
			`{"winner": "A", "reasoning": "Pass 2 favors slot A"}`,
		},
	}

	ctx := context.Background()
	res, err := judge.EvaluateDebiased(ctx, client, "model-1", "model-2", "Test prompt", "Resp 1", "Resp 2", "", 0.1, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res.FinalWinner != "tie" {
		t.Errorf("expected tie due to position bias, got %s", res.FinalWinner)
	}
	if !res.PositionBiasDetected {
		t.Errorf("expected PositionBiasDetected to be true")
	}
	if !res.Inconsistent {
		t.Errorf("expected Inconsistent to be true")
	}
}

func TestEvaluateSinglePass_MarkdownJSONParsing(t *testing.T) {
	markdownWrapped := "Here is my judgment:\n```json\n{\"winner\": \"B\", \"reasoning\": \"Response B is superior\"}\n```\nHope this helps!"
	client := &testJudgeClient{
		verdicts: []string{markdownWrapped},
	}

	ctx := context.Background()
	v, err := judge.EvaluateSinglePass(ctx, client, "prompt", "respA", "respB", "", 0.1)
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	if v.Winner != "B" {
		t.Errorf("expected winner B, got %s", v.Winner)
	}
}
