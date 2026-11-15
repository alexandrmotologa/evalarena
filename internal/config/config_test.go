package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/alexandrmotologa/evalarena/internal/config"
)

func TestLoadConfig(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "test.yaml")
	datasetPath := filepath.Join(tempDir, "dataset.jsonl")

	if err := os.WriteFile(datasetPath, []byte(`{"id":"1","prompt":"hello"}`), 0644); err != nil {
		t.Fatalf("failed to create dummy dataset: %v", err)
	}

	yamlContent := `
tournament:
  name: "Unit Test Arena"
  schedule: "round-robin"
  rounds: 2
  concurrency: 2
  database: "test.db"

judge:
  provider: "mock"
  model: "mock-judge"
  debias_position: true

models:
  - id: "m1"
    provider: "mock"
    model: "mock-m1"
  - id: "m2"
    provider: "mock"
    model: "mock-m2"

dataset:
  path: "dataset.jsonl"
  format: "jsonl"
`
	if err := os.WriteFile(configPath, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		t.Fatalf("expected clean load, got error: %v", err)
	}

	if cfg.Tournament.Name != "Unit Test Arena" {
		t.Errorf("expected tournament name 'Unit Test Arena', got %q", cfg.Tournament.Name)
	}
	if len(cfg.Models) != 2 {
		t.Errorf("expected 2 models, got %d", len(cfg.Models))
	}
	if !cfg.Judge.DebiasPosition {
		t.Errorf("expected debias_position to be true")
	}
}

func TestConfigValidation_FailsWithLessThanTwoModels(t *testing.T) {
	cfg := config.Config{
		Tournament: config.TournamentConfig{Schedule: "round-robin"},
		Dataset:    config.DatasetConfig{Path: "test.jsonl"},
		Models: []config.ModelConfig{
			{ID: "m1", Provider: "mock"},
		},
	}
	if err := cfg.Validate(); err == nil {
		t.Fatalf("expected error validating config with only 1 model, got nil")
	}
}
