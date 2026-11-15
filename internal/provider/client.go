package provider

import (
	"context"
	"time"
)

// Message represents a single chat turn in multi-turn completions.
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// CompletionRequest defines prompt input parameters for a model.
type CompletionRequest struct {
	Prompt        string    `json:"prompt"`
	SystemPrompt  string    `json:"system_prompt,omitempty"`
	Messages      []Message `json:"messages,omitempty"`
	Temperature   float64   `json:"temperature"`
	MaxTokens     int       `json:"max_tokens,omitempty"`
	StopSequences []string  `json:"stop_sequences,omitempty"`
}

// CompletionResponse encapsulates generation outputs, timing, and token metrics.
type CompletionResponse struct {
	Text         string        `json:"text"`
	ModelID      string        `json:"model_id"`
	ProviderName string        `json:"provider_name"`
	Latency      time.Duration `json:"latency"`
	LatencyMs    int64         `json:"latency_ms"`
	TTFTMs       int64         `json:"ttft_ms"`
	InputTokens  int           `json:"input_tokens"`
	OutputTokens int           `json:"output_tokens"`
	TotalCost    float64       `json:"total_cost"`
}

// Client defines the common interface for all foundation model providers.
type Client interface {
	ID() string
	Generate(ctx context.Context, req CompletionRequest) (*CompletionResponse, error)
}
