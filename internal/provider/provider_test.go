package provider_test

import (
	"context"
	"testing"
	"time"

	"github.com/alexandrmotologa/evalarena/internal/config"
	"github.com/alexandrmotologa/evalarena/internal/provider"
)

func TestPricingCalculator(t *testing.T) {
	calc := provider.NewCalculator(config.PricingConfig{
		InputPerMillion:  2.0,
		OutputPerMillion: 10.0,
	})

	// 1,000 input tokens = $0.002, 1,000 output tokens = $0.010 -> total $0.012
	cost := calc.Calculate(1000, 1000)
	expected := 0.012
	diff := cost - expected
	if diff < -1e-6 || diff > 1e-6 {
		t.Errorf("expected cost %.6f, got %.6f", expected, cost)
	}
}

func TestMockClient_Generate(t *testing.T) {
	cfg := config.ModelConfig{
		ID:       "sim-gpt4o",
		Provider: "mock",
		Model:    "gpt-4o",
		Pricing: config.PricingConfig{
			InputPerMillion:  2.5,
			OutputPerMillion: 10.0,
		},
	}

	client := provider.NewMockClient(cfg)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	resp, err := client.Generate(ctx, provider.CompletionRequest{
		Prompt: "Test enterprise prompt",
	})
	if err != nil {
		t.Fatalf("unexpected mock generate error: %v", err)
	}

	if resp.Text == "" {
		t.Errorf("expected non-empty completion text")
	}
	if resp.ModelID != "sim-gpt4o" {
		t.Errorf("expected model id sim-gpt4o, got %s", resp.ModelID)
	}
	if resp.InputTokens <= 0 || resp.OutputTokens <= 0 {
		t.Errorf("expected positive token counts: in=%d, out=%d", resp.InputTokens, resp.OutputTokens)
	}
	if resp.TotalCost <= 0 {
		t.Errorf("expected positive total cost, got %f", resp.TotalCost)
	}
}

func TestPool_DispatchPair(t *testing.T) {
	models := []config.ModelConfig{
		{ID: "m1", Provider: "mock", Model: "gpt-4o"},
		{ID: "m2", Provider: "mock", Model: "claude-3-5-sonnet"},
	}
	judgeCfg := config.JudgeConfig{
		Provider: "mock",
		Model:    "mock-judge",
	}

	pool, err := provider.NewPool(models, judgeCfg)
	if err != nil {
		t.Fatalf("failed to create pool: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	respA, respB, err := pool.DispatchPair(ctx, "m1", "m2", provider.CompletionRequest{
		Prompt: "Compare these responses",
	})
	if err != nil {
		t.Fatalf("failed to dispatch pair: %v", err)
	}

	if respA.ModelID != "m1" || respB.ModelID != "m2" {
		t.Errorf("unexpected model IDs: %s, %s", respA.ModelID, respB.ModelID)
	}
}
