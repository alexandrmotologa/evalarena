package config

// Config represents the top-level evalarena tournament configuration.
type Config struct {
	Version    string           `yaml:"version"`
	Tournament TournamentConfig `yaml:"tournament"`
	Judge      JudgeConfig      `yaml:"judge"`
	Models     []ModelConfig    `yaml:"models"`
	Dataset    DatasetConfig    `yaml:"dataset"`
	Gate       GateConfig       `yaml:"gate,omitempty"`
}

// TournamentConfig defines scheduling and execution options.
type TournamentConfig struct {
	Name        string `yaml:"name"`
	Schedule    string `yaml:"schedule"` // "round-robin", "swiss", "random"
	Rounds      int    `yaml:"rounds"`
	Concurrency int    `yaml:"concurrency"`
	Database    string `yaml:"database"`
}

// JudgeConfig defines the automated evaluator options.
type JudgeConfig struct {
	Provider       string  `yaml:"provider"` // "openai", "anthropic", "gemini", "ollama", "mock"
	Model          string  `yaml:"model"`
	Endpoint       string  `yaml:"endpoint,omitempty"`
	APIKey         string  `yaml:"api_key,omitempty"`
	Temperature    float64 `yaml:"temperature"`
	DebiasPosition bool    `yaml:"debias_position"`
	RubricFile     string  `yaml:"rubric_file,omitempty"`
	RubricContent  string  `yaml:"-"`
}

// ModelConfig defines an evaluated model candidate.
type ModelConfig struct {
	ID       string        `yaml:"id"`
	Provider string        `yaml:"provider"`
	Model    string        `yaml:"model"`
	Endpoint string        `yaml:"endpoint,omitempty"`
	APIKey   string        `yaml:"api_key,omitempty"`
	Pricing  PricingConfig `yaml:"pricing"`
}

// PricingConfig defines cost per million tokens in USD.
type PricingConfig struct {
	InputPerMillion  float64 `yaml:"input_per_million"`
	OutputPerMillion float64 `yaml:"output_per_million"`
}

// DatasetConfig defines the prompt benchmark dataset.
type DatasetConfig struct {
	Path   string `yaml:"path"`
	Format string `yaml:"format"` // "jsonl", "csv"
}

// GateConfig defines automated CI/CD quality gate thresholds.
type GateConfig struct {
	BaselineModel  string  `yaml:"baseline_model"`
	CandidateModel string  `yaml:"candidate_model"`
	MaxELODrop     float64 `yaml:"max_elo_drop"`
	MaxCostRatio   float64 `yaml:"max_cost_ratio,omitempty"`
}
