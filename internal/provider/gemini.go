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

// GeminiClient implements Client for Google Gemini API.
type GeminiClient struct {
	id         string
	model      string
	apiKey     string
	calculator *Calculator
	httpClient *http.Client
}

// NewGeminiClient creates a Gemini API client.
func NewGeminiClient(cfg config.ModelConfig) *GeminiClient {
	apiKey := cfg.APIKey
	if apiKey == "" {
		apiKey = os.Getenv("GEMINI_API_KEY")
	}

	return &GeminiClient{
		id:         cfg.ID,
		model:      cfg.Model,
		apiKey:     apiKey,
		calculator: NewCalculator(cfg.Pricing),
		httpClient: &http.Client{Timeout: 60 * time.Second},
	}
}

func (c *GeminiClient) ID() string {
	return c.id
}

type geminiPart struct {
	Text string `json:"text"`
}

type geminiContent struct {
	Role  string       `json:"role,omitempty"`
	Parts []geminiPart `json:"parts"`
}

type geminiRequest struct {
	Contents          []geminiContent `json:"contents"`
	SystemInstruction *geminiContent  `json:"systemInstruction,omitempty"`
	GenerationConfig  struct {
		Temperature     float64 `json:"temperature"`
		MaxOutputTokens int     `json:"maxOutputTokens,omitempty"`
	} `json:"generationConfig"`
}

type geminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []geminiPart `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
	UsageMetadata struct {
		PromptTokenCount     int `json:"promptTokenCount"`
		CandidatesTokenCount int `json:"candidatesTokenCount"`
		TotalTokenCount      int `json:"totalTokenCount"`
	} `json:"usageMetadata"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func (c *GeminiClient) Generate(ctx context.Context, req CompletionRequest) (*CompletionResponse, error) {
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s", c.model, c.apiKey)

	var payload geminiRequest
	payload.GenerationConfig.Temperature = req.Temperature
	if req.MaxTokens > 0 {
		payload.GenerationConfig.MaxOutputTokens = req.MaxTokens
	}

	if req.SystemPrompt != "" {
		payload.SystemInstruction = &geminiContent{
			Parts: []geminiPart{{Text: req.SystemPrompt}},
		}
	}

	if len(req.Messages) > 0 {
		for _, m := range req.Messages {
			role := m.Role
			if role == "assistant" {
				role = "model"
			} else {
				role = "user"
			}
			payload.Contents = append(payload.Contents, geminiContent{
				Role:  role,
				Parts: []geminiPart{{Text: m.Content}},
			})
		}
	} else {
		payload.Contents = []geminiContent{
			{
				Role:  "user",
				Parts: []geminiPart{{Text: req.Prompt}},
			},
		}
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to encode gemini request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create http request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	start := time.Now()
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("gemini request failed: %w", err)
	}
	defer resp.Body.Close()

	latency := time.Since(start)

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("gemini api returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var parsed geminiResponse
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return nil, fmt.Errorf("failed to decode gemini response: %w", err)
	}

	if parsed.Error != nil {
		return nil, fmt.Errorf("gemini error: %s", parsed.Error.Message)
	}

	if len(parsed.Candidates) == 0 || len(parsed.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("gemini returned 0 candidates")
	}

	content := parsed.Candidates[0].Content.Parts[0].Text
	cost := c.calculator.Calculate(parsed.UsageMetadata.PromptTokenCount, parsed.UsageMetadata.CandidatesTokenCount)

	return &CompletionResponse{
		Text:         content,
		ModelID:      c.id,
		ProviderName: "gemini",
		Latency:      latency,
		LatencyMs:    latency.Milliseconds(),
		TTFTMs:       latency.Milliseconds() / 3,
		InputTokens:  parsed.UsageMetadata.PromptTokenCount,
		OutputTokens: parsed.UsageMetadata.CandidatesTokenCount,
		TotalCost:    cost,
	}, nil
}
