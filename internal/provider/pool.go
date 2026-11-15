package provider

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/alexandrmotologa/evalarena/internal/config"
)

// Pool manages provider client instances and concurrent request dispatching.
type Pool struct {
	clients map[string]Client
	judge   Client
	mu      sync.RWMutex
}

// NewPool initializes model and judge provider clients from configuration.
func NewPool(models []config.ModelConfig, judgeCfg config.JudgeConfig) (*Pool, error) {
	clients := make(map[string]Client)

	for _, m := range models {
		client, err := createClient(m.Provider, m)
		if err != nil {
			return nil, fmt.Errorf("failed to create client for model %q: %w", m.ID, err)
		}
		clients[m.ID] = client
	}

	// Create judge client
	judgeModelCfg := config.ModelConfig{
		ID:       "judge",
		Provider: judgeCfg.Provider,
		Model:    judgeCfg.Model,
		Endpoint: judgeCfg.Endpoint,
		APIKey:   judgeCfg.APIKey,
		Pricing: config.PricingConfig{
			InputPerMillion:  2.50,
			OutputPerMillion: 10.00,
		},
	}
	judgeClient, err := createClient(judgeCfg.Provider, judgeModelCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create judge client: %w", err)
	}

	return &Pool{
		clients: clients,
		judge:   judgeClient,
	}, nil
}

func createClient(providerName string, cfg config.ModelConfig) (Client, error) {
	switch strings.ToLower(providerName) {
	case "mock":
		return NewMockClient(cfg), nil
	case "openai":
		return NewOpenAIClient(cfg), nil
	case "anthropic":
		return NewAnthropicClient(cfg), nil
	case "gemini":
		return NewGeminiClient(cfg), nil
	case "ollama", "vllm":
		return NewOllamaClient(cfg), nil
	default:
		return nil, fmt.Errorf("unsupported provider %q", providerName)
	}
}

// GetClient retrieves the client instance for a model ID.
func (p *Pool) GetClient(modelID string) (Client, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	client, ok := p.clients[modelID]
	if !ok {
		return nil, fmt.Errorf("model %q not registered in provider pool", modelID)
	}
	return client, nil
}

// GetJudge retrieves the configured automated judge client.
func (p *Pool) GetJudge() Client {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.judge
}

// DispatchPair queries Model A and Model B in parallel for the given prompt request.
func (p *Pool) DispatchPair(ctx context.Context, modelAID, modelBID string, req CompletionRequest) (*CompletionResponse, *CompletionResponse, error) {
	clientA, err := p.GetClient(modelAID)
	if err != nil {
		return nil, nil, err
	}
	clientB, err := p.GetClient(modelBID)
	if err != nil {
		return nil, nil, err
	}

	type result struct {
		resp *CompletionResponse
		err  error
	}

	chA := make(chan result, 1)
	chB := make(chan result, 1)

	go func() {
		resp, err := clientA.Generate(ctx, req)
		chA <- result{resp: resp, err: err}
	}()

	go func() {
		resp, err := clientB.Generate(ctx, req)
		chB <- result{resp: resp, err: err}
	}()

	resA := <-chA
	if resA.err != nil {
		return nil, nil, fmt.Errorf("model %s generation failed: %w", modelAID, resA.err)
	}

	resB := <-chB
	if resB.err != nil {
		return nil, nil, fmt.Errorf("model %s generation failed: %w", modelBID, resB.err)
	}

	return resA.resp, resB.resp, nil
}
