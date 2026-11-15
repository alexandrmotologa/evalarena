package provider

import "github.com/alexandrmotologa/evalarena/internal/config"

// Calculator computes cost based on input/output token counts and pricing rates.
type Calculator struct {
	InputPerMillion  float64
	OutputPerMillion float64
}

// NewCalculator initializes a pricing calculator from config.
func NewCalculator(cfg config.PricingConfig) *Calculator {
	return &Calculator{
		InputPerMillion:  cfg.InputPerMillion,
		OutputPerMillion: cfg.OutputPerMillion,
	}
}

// Calculate returns total cost in USD for the given token quantities.
func (c *Calculator) Calculate(inputTokens, outputTokens int) float64 {
	inputCost := (float64(inputTokens) / 1_000_000.0) * c.InputPerMillion
	outputCost := (float64(outputTokens) / 1_000_000.0) * c.OutputPerMillion
	return inputCost + outputCost
}
