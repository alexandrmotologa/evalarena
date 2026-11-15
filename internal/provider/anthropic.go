package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/alexandrmotologa/evalarena/internal/config"
)

// AnthropicClient implements Client for Anthropic Messages API.
type AnthropicClient struct {
	id         string
	model      string
	endpoint   string
	apiKey     string
	calculator *Calculator
	httpClient *http.Client
}

// NewAnthropicClient creates an Anthropic API client.
func NewAnthropicClient(cfg config.ModelConfig) *AnthropicClient {
	endpoint := cfg.Endpoint
	if endpoint == "" {
		endpoint = "https://api.anthropic.com/v1"
	}

	apiKey := cfg.APIKey
	if apiKey == "" {
		apiKey = os.Getenv("ANTHROPIC_API_KEY")
	}

	return &AnthropicClient{
		id:         cfg.ID,
		model:      cfg.Model,
		endpoint:   endpoint,
		apiKey:     apiKey,
		calculator: NewCalculator(cfg.Pricing),
		httpClient: &http.Client{Timeout: 60 * time.Second},
	}
}

func (c *AnthropicClient) ID() string {
	return c.id
}

type anthropicMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type anthropicRequest struct {
	Model       string             `json:"model"`
	Messages    []anthropicMessage `json:"messages"`
	System      string             `json:"system,omitempty"`
	MaxTokens   int                `json:"max_tokens"`
	Temperature float64            `json:"temperature"`
}

type anthropicResponse struct {
	Content []struct {
		Text string `json:"text"`
	} `json:"content"`
	Usage struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func (c *AnthropicClient) Generate(ctx context.Context, req CompletionRequest) (*CompletionResponse, error) {
	url := fmt.Sprintf("%s/messages", c.endpoint)

	maxTokens := req.MaxTokens
	if maxTokens <= 0 {
		maxTokens = 2048
	}

	var messages []anthropicMessage
	systemPrompt := req.SystemPrompt

	if len(req.Messages) > 0 {
		for _, m := range req.Messages {
			if m.Role == "system" {
				if systemPrompt == "" {
					systemPrompt = m.Content
				}
				continue
			}
			role := m.Role
			if role != "assistant" {
				role = "user"
			}
			messages = append(messages, anthropicMessage{Role: role, Content: m.Content})
		}
	} else {
		messages = []anthropicMessage{
			{Role: "user", Content: req.Prompt},
		}
	}

	payload := anthropicRequest{
		Model:       c.model,
		Messages:    messages,
		System:      systemPrompt,
		MaxTokens:   maxTokens,
		Temperature: req.Temperature,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to encode anthropic request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create http request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", c.apiKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")

	start := time.Now()
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("anthropic request failed: %w", err)
	}
	defer resp.Body.Close()

	latency := time.Since(start)

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("anthropic api returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var parsed anthropicResponse
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return nil, fmt.Errorf("failed to decode anthropic response: %w", err)
	}

	if parsed.Error != nil {
		return nil, fmt.Errorf("anthropic error: %s", parsed.Error.Message)
	}

	if len(parsed.Content) == 0 {
		return nil, fmt.Errorf("anthropic returned 0 content blocks")
	}

	content := parsed.Content[0].Text
	cost := c.calculator.Calculate(parsed.Usage.InputTokens, parsed.Usage.OutputTokens)

	return &CompletionResponse{
		Text:         content,
		ModelID:      c.id,
		ProviderName: "anthropic",
		Latency:      latency,
		LatencyMs:    latency.Milliseconds(),
		TTFTMs:       latency.Milliseconds() / 3,
		InputTokens:  parsed.Usage.InputTokens,
		OutputTokens: parsed.Usage.OutputTokens,
		TotalCost:    cost,
	}, nil
}
