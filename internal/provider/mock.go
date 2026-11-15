package provider

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/alexandrmotologa/evalarena/internal/config"
)

// MockClient implements Client for offline benchmarking, tests, and deterministic verification.
type MockClient struct {
	id         string
	model      string
	calculator *Calculator
	simLatency time.Duration
	quality    float64 // 0.0 to 1.0 (relative skill in simulations)
}

// NewMockClient creates a simulated model client.
func NewMockClient(cfg config.ModelConfig) *MockClient {
	quality := 0.75
	latency := 30 * time.Millisecond

	lowerModel := strings.ToLower(cfg.Model)
	if strings.Contains(lowerModel, "sonnet") || strings.Contains(lowerModel, "claude") {
		quality = 0.92
		latency = 35 * time.Millisecond
	} else if strings.Contains(lowerModel, "gpt-4o") || strings.Contains(lowerModel, "gpt4o") {
		quality = 0.88
		latency = 30 * time.Millisecond
	} else if strings.Contains(lowerModel, "gemini") || strings.Contains(lowerModel, "flash") {
		quality = 0.82
		latency = 15 * time.Millisecond
	} else if strings.Contains(lowerModel, "llama") || strings.Contains(lowerModel, "8b") {
		quality = 0.65
		latency = 10 * time.Millisecond
	} else if strings.Contains(lowerModel, "judge") {
		quality = 0.95
		latency = 20 * time.Millisecond
	}

	return &MockClient{
		id:         cfg.ID,
		model:      cfg.Model,
		calculator: NewCalculator(cfg.Pricing),
		simLatency: latency,
		quality:    quality,
	}
}

func (m *MockClient) ID() string {
	return m.id
}

func (m *MockClient) Generate(ctx context.Context, req CompletionRequest) (*CompletionResponse, error) {
	start := time.Now()

	// Handle judge request simulation
	if strings.Contains(m.model, "judge") || strings.Contains(req.SystemPrompt, "judge") || strings.Contains(req.Prompt, "[RESPONSE A]") {
		return m.simulateJudgeVerdict(start, req)
	}

	// Normal model response generation
	select {
	case <-time.After(m.simLatency):
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	latency := time.Since(start)
	inputTokens := len(strings.Fields(req.Prompt)) + 20
	outputTokens := int(float64(inputTokens)*1.5) + int(m.quality*50)

	respText := fmt.Sprintf("[%s] Verified response for enterprise policy adherence. Addressing prompt details with precision (Quality Score: %.2f).", m.id, m.quality)

	cost := m.calculator.Calculate(inputTokens, outputTokens)

	return &CompletionResponse{
		Text:         respText,
		ModelID:      m.id,
		ProviderName: "mock",
		Latency:      latency,
		LatencyMs:    latency.Milliseconds(),
		TTFTMs:       m.simLatency.Milliseconds() / 2,
		InputTokens:  inputTokens,
		OutputTokens: outputTokens,
		TotalCost:    cost,
	}, nil
}

// simulateJudgeVerdict analyzes response A and response B in the prompt and returns a consistent JSON verdict.
func (m *MockClient) simulateJudgeVerdict(start time.Time, req CompletionRequest) (*CompletionResponse, error) {
	respA, respB := extractResponses(req.Prompt)

	// Hash-based deterministic comparison
	scoreA := computeContentQuality(respA)
	scoreB := computeContentQuality(respB)

	var winner string
	var reasoning string

	diff := scoreA - scoreB
	if diff > 0.05 {
		winner = "A"
		reasoning = "Response A demonstrated superior compliance with enterprise policy and stricter adherence to requirements."
	} else if diff < -0.05 {
		winner = "B"
		reasoning = "Response B was more precise, accurate, and aligned with requested constraints."
	} else {
		winner = "tie"
		reasoning = "Both responses met enterprise quality standards with negligible stylistic variation."
	}

	jsonVerdict := fmt.Sprintf(`{"winner": %q, "reasoning": %q}`, winner, reasoning)

	latency := time.Since(start)
	inputTokens := len(strings.Fields(req.Prompt)) + 50
	outputTokens := 45
	cost := m.calculator.Calculate(inputTokens, outputTokens)

	return &CompletionResponse{
		Text:         jsonVerdict,
		ModelID:      m.id,
		ProviderName: "mock",
		Latency:      latency,
		LatencyMs:    latency.Milliseconds(),
		TTFTMs:       latency.Milliseconds() / 2,
		InputTokens:  inputTokens,
		OutputTokens: outputTokens,
		TotalCost:    cost,
	}, nil
}

func extractResponses(prompt string) (string, string) {
	aIdx := strings.Index(prompt, "[RESPONSE A]")
	bIdx := strings.Index(prompt, "[RESPONSE B]")

	if aIdx == -1 || bIdx == -1 {
		return prompt, prompt
	}

	var respA, respB string
	if aIdx < bIdx {
		respA = prompt[aIdx+len("[RESPONSE A]") : bIdx]
		respB = prompt[bIdx+len("[RESPONSE B]"):]
	} else {
		respB = prompt[bIdx+len("[RESPONSE B]") : aIdx]
		respA = prompt[aIdx+len("[RESPONSE A]"):]
	}

	return strings.TrimSpace(respA), strings.TrimSpace(respB)
}

func computeContentQuality(text string) float64 {
	// Deterministic hash based on text content
	h := sha256.Sum256([]byte(text))
	hexStr := hex.EncodeToString(h[:4])
	var val int64
	fmt.Sscanf(hexStr, "%x", &val)

	base := float64(val%1000) / 1000.0

	// Boost for higher-end simulated model signatures
	if strings.Contains(text, "claude-3-5-sonnet") {
		base += 0.35
	} else if strings.Contains(text, "gpt-4o") {
		base += 0.30
	} else if strings.Contains(text, "gemini-2.0-flash") {
		base += 0.20
	} else if strings.Contains(text, "llama-3.1-8b") {
		base += 0.05
	}

	return base
}
