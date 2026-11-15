package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/alexandrmotologa/evalarena/internal/config"
)

// OllamaClient implements Client for local Ollama and vLLM servers.
type OllamaClient struct {
	id         string
	model      string
	endpoint   string
	calculator *Calculator
	httpClient *http.Client
}

// NewOllamaClient creates an Ollama client.
func NewOllamaClient(cfg config.ModelConfig) *OllamaClient {
	endpoint := cfg.Endpoint
	if endpoint == "" {
		endpoint = "http://localhost:11434"
	}
	endpoint = strings.TrimSuffix(endpoint, "/")

	return &OllamaClient{
		id:         cfg.ID,
		model:      cfg.Model,
		endpoint:   endpoint,
		calculator: NewCalculator(cfg.Pricing),
		httpClient: &http.Client{Timeout: 120 * time.Second},
	}
}

func (c *OllamaClient) ID() string {
	return c.id
}

type ollamaRequest struct {
	Model       string                 `json:"model"`
	Prompt      string                 `json:"prompt"`
	System      string                 `json:"system,omitempty"`
	Stream      bool                   `json:"stream"`
	Options     map[string]interface{} `json:"options,omitempty"`
}

type ollamaResponse struct {
	Response        string `json:"response"`
	PromptEvalCount int    `json:"prompt_eval_count"`
	EvalCount       int    `json:"eval_count"`
	TotalDuration   int64  `json:"total_duration"` // nanoseconds
	LoadDuration    int64  `json:"load_duration"`
	PromptDuration  int64  `json:"prompt_eval_duration"`
	Error           string `json:"error,omitempty"`
}

func (c *OllamaClient) Generate(ctx context.Context, req CompletionRequest) (*CompletionResponse, error) {
	url := fmt.Sprintf("%s/api/generate", c.endpoint)

	payload := ollamaRequest{
		Model:  c.model,
		Prompt: req.Prompt,
		System: req.SystemPrompt,
		Stream: false,
		Options: map[string]interface{}{
			"temperature": req.Temperature,
		},
	}
	if req.MaxTokens > 0 {
		payload.Options["num_predict"] = req.MaxTokens
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to encode ollama request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create http request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	start := time.Now()
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("ollama request failed: %w", err)
	}
	defer resp.Body.Close()

	latency := time.Since(start)

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ollama api returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var parsed ollamaResponse
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return nil, fmt.Errorf("failed to decode ollama response: %w", err)
	}

	if parsed.Error != "" {
		return nil, fmt.Errorf("ollama error: %s", parsed.Error)
	}

	inputTokens := parsed.PromptEvalCount
	outputTokens := parsed.EvalCount
	if inputTokens == 0 {
		inputTokens = len(strings.Fields(req.Prompt))
	}
	if outputTokens == 0 {
		outputTokens = len(strings.Fields(parsed.Response))
	}

	cost := c.calculator.Calculate(inputTokens, outputTokens)

	ttftMs := int64(0)
	if parsed.PromptDuration > 0 {
		ttftMs = parsed.PromptDuration / 1_000_000
	} else {
		ttftMs = latency.Milliseconds() / 4
	}

	return &CompletionResponse{
		Text:         parsed.Response,
		ModelID:      c.id,
		ProviderName: "ollama",
		Latency:      latency,
		LatencyMs:    latency.Milliseconds(),
		TTFTMs:       ttftMs,
		InputTokens:  inputTokens,
		OutputTokens: outputTokens,
		TotalCost:    cost,
	}, nil
}
