package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// DefaultPricing provides baseline rates per 1M tokens when omitted.
var DefaultPricing = map[string]PricingConfig{
	"gpt-4o": {
		InputPerMillion:  2.50,
		OutputPerMillion: 10.00,
	},
	"gpt-4o-mini": {
		InputPerMillion:  0.15,
		OutputPerMillion: 0.60,
	},
	"claude-3-5-sonnet": {
		InputPerMillion:  3.00,
		OutputPerMillion: 15.00,
	},
	"claude-3-5-haiku": {
		InputPerMillion:  0.80,
		OutputPerMillion: 4.00,
	},
	"gemini-2.0-flash": {
		InputPerMillion:  0.10,
		OutputPerMillion: 0.40,
	},
	"gemini-1.5-pro": {
		InputPerMillion:  1.25,
		OutputPerMillion: 5.00,
	},
	"llama-3.1-8b": {
		InputPerMillion:  0.05,
		OutputPerMillion: 0.08,
	},
	"llama-3.1-70b": {
		InputPerMillion:  0.35,
		OutputPerMillion: 0.40,
	},
}

// Load reads and parses a YAML configuration file from disk.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %q: %w", path, err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse YAML config: %w", err)
	}

	applyDefaults(&cfg, filepath.Dir(path))

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	// Read rubric file if specified
	if cfg.Judge.RubricFile != "" {
		rubricPath := cfg.Judge.RubricFile
		if _, err := os.Stat(rubricPath); err != nil && !filepath.IsAbs(rubricPath) {
			rubricPath = filepath.Join(filepath.Dir(path), rubricPath)
		}
		content, err := os.ReadFile(rubricPath)
		if err == nil {
			cfg.Judge.RubricContent = string(content)
		}
	}

	return &cfg, nil
}

// applyDefaults populates missing fields with safe operational defaults.
func applyDefaults(cfg *Config, baseDir string) {
	if cfg.Tournament.Name == "" {
		cfg.Tournament.Name = "evalarena Tournament"
	}
	if cfg.Tournament.Schedule == "" {
		cfg.Tournament.Schedule = "round-robin"
	}
	if cfg.Tournament.Rounds <= 0 {
		cfg.Tournament.Rounds = 1
	}
	if cfg.Tournament.Concurrency <= 0 {
		cfg.Tournament.Concurrency = 4
	}
	if cfg.Tournament.Database == "" {
		cfg.Tournament.Database = "evalarena.db"
	} else if !filepath.IsAbs(cfg.Tournament.Database) && baseDir != "" {
		if _, err := os.Stat(cfg.Tournament.Database); err != nil {
			cfg.Tournament.Database = filepath.Join(baseDir, cfg.Tournament.Database)
		}
	}

	if cfg.Judge.Provider == "" {
		cfg.Judge.Provider = "mock"
	}
	if cfg.Judge.Model == "" {
		cfg.Judge.Model = "mock-judge"
	}
	if cfg.Judge.Temperature == 0 {
		cfg.Judge.Temperature = 0.1
	}

	// Adjust dataset path relative to config file if needed
	if cfg.Dataset.Path != "" && !filepath.IsAbs(cfg.Dataset.Path) && baseDir != "" {
		if _, err := os.Stat(cfg.Dataset.Path); err != nil {
			cfg.Dataset.Path = filepath.Join(baseDir, cfg.Dataset.Path)
		}
	}
	if cfg.Dataset.Format == "" {
		cfg.Dataset.Format = "jsonl"
	}

	// Populate model pricing if missing
	for i := range cfg.Models {
		m := &cfg.Models[i]
		if m.Pricing.InputPerMillion == 0 && m.Pricing.OutputPerMillion == 0 {
			lowerModel := strings.ToLower(m.Model)
			for key, price := range DefaultPricing {
				if strings.Contains(lowerModel, key) {
					m.Pricing = price
					break
				}
			}
		}
	}
}

// Validate checks for logical errors and missing mandatory fields.
func (c *Config) Validate() error {
	if len(c.Models) < 2 {
		return fmt.Errorf("at least 2 competing models are required for pairwise tournaments, found %d", len(c.Models))
	}

	ids := make(map[string]bool)
	for _, m := range c.Models {
		if strings.TrimSpace(m.ID) == "" {
			return fmt.Errorf("model id cannot be empty")
		}
		if ids[m.ID] {
			return fmt.Errorf("duplicate model id %q", m.ID)
		}
		ids[m.ID] = true

		if m.Provider == "" {
			return fmt.Errorf("model %q missing provider", m.ID)
		}
	}

	if c.Dataset.Path == "" {
		return fmt.Errorf("dataset path is required")
	}

	validSchedules := map[string]bool{
		"round-robin": true,
		"swiss":       true,
		"random":      true,
	}
	if !validSchedules[c.Tournament.Schedule] {
		return fmt.Errorf("unsupported tournament schedule %q (valid: round-robin, swiss, random)", c.Tournament.Schedule)
	}

	return nil
}
