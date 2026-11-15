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

// OpenAIClient implements Client for OpenAI compatible endpoints.
type OpenAIClient struct {
	id         string
	model      string
	endpoint   string
	apiKey     string
	calculator *Calculator
	httpClient *http.Client
}

// NewOpenAIClient creates an OpenAI API client.
func NewOpenAIClient(cfg config.ModelConfig) *OpenAIClient {
	endpoint := cfg.Endpoint
	if endpoint == "" {
		endpoint = "https://api.openai.com/v1"
	}

	apiKey := cfg.APIKey
	if apiKey == "" {
		apiKey = os.Getenv("OPENAI_API_KEY")
	}

	return &OpenAIClient{
		id:         cfg.ID,
		model:      cfg.Model,
		endpoint:   endpoint,
		apiKey:     apiKey,
		calculator: NewCalculator(cfg.Pricing),
		httpClient: &http.Client{Timeout: 60 * time.Second},
	}
}

func (c *OpenAIClient) ID() string {
	return c.id
}

type openAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openAIRequest struct {
	Model       string          `json:"model"`
	Messages    []openAIMessage `json:"messages"`
	Temperature float64         `json:"temperature"`
	MaxTokens   int             `json:"max_tokens,omitempty"`
}

type openAIResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func (c *OpenAIClient) Generate(ctx context.Context, req CompletionRequest) (*CompletionResponse, error) {
	url := fmt.Sprintf("%s/chat/completions", c.endpoint)

	var messages []openAIMessage
	if req.SystemPrompt != "" {
		messages = append(messages, openAIMessage{Role: "system", Content: req.SystemPrompt})
	}
	if len(req.Messages) > 0 {
		for _, m := range req.Messages {
			role := m.Role
			if role == "" {
				role = "user"
			}
			messages = append(messages, openAIMessage{Role: role, Content: m.Content})
		}
	} else {
		messages = append(messages, openAIMessage{Role: "user", Content: req.Prompt})
	}

	payload := openAIRequest{
		Model:       c.model,
		Messages:    messages,
		Temperature: req.Temperature,
		MaxTokens:   req.MaxTokens,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to encode openai request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create http request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))
	}

	start := time.Now()
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("openai request failed: %w", err)
	}
	defer resp.Body.Close()

	latency := time.Since(start)

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("openai api returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var parsed openAIResponse
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return nil, fmt.Errorf("failed to decode openai response: %w", err)
	}

	if parsed.Error != nil {
		return nil, fmt.Errorf("openai error: %s", parsed.Error.Message)
	}

	if len(parsed.Choices) == 0 {
		return nil, fmt.Errorf("openai returned 0 choices")
	}

	content := parsed.Choices[0].Message.Content
	cost := c.calculator.Calculate(parsed.Usage.PromptTokens, parsed.Usage.CompletionTokens)

	return &CompletionResponse{
		Text:         content,
		ModelID:      c.id,
		ProviderName: "openai",
		Latency:      latency,
		LatencyMs:    latency.Milliseconds(),
		TTFTMs:       latency.Milliseconds() / 3, // approximated when non-streaming
		InputTokens:  parsed.Usage.PromptTokens,
		OutputTokens: parsed.Usage.CompletionTokens,
		TotalCost:    cost,
	}, nil
}
