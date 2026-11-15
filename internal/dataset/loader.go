package dataset

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Item represents a single benchmark evaluation prompt case.
type Item struct {
	ID           string                 `json:"id" yaml:"id"`
	Category     string                 `json:"category" yaml:"category"`
	Prompt       string                 `json:"prompt" yaml:"prompt"`
	SystemPrompt string                 `json:"system_prompt,omitempty" yaml:"system_prompt,omitempty"`
	Reference    string                 `json:"reference,omitempty" yaml:"reference,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty" yaml:"metadata,omitempty"`
}

// LoadDataset reads benchmark test cases from JSONL, CSV, or YAML formats.
func LoadDataset(path string, format string) ([]Item, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open dataset file %q: %w", path, err)
	}
	defer file.Close()

	normFormat := strings.ToLower(strings.TrimPrefix(format, "."))
	if normFormat == "" {
		ext := strings.ToLower(filepath.Ext(path))
		normFormat = strings.TrimPrefix(ext, ".")
	}

	switch normFormat {
	case "jsonl":
		return loadJSONL(file)
	case "csv":
		return loadCSV(file)
	case "yaml", "yml":
		return loadYAML(file)
	default:
		return nil, fmt.Errorf("unsupported dataset format %q (expected jsonl, csv, or yaml)", format)
	}
}

func loadJSONL(r io.Reader) ([]Item, error) {
	var items []Item
	scanner := bufio.NewScanner(r)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		var item Item
		if err := json.Unmarshal([]byte(line), &item); err != nil {
			return nil, fmt.Errorf("line %d: invalid JSON object: %w", lineNum, err)
		}

		if item.ID == "" {
			item.ID = fmt.Sprintf("prompt_%04d", lineNum)
		}
		if strings.TrimSpace(item.Prompt) == "" {
			return nil, fmt.Errorf("line %d: empty prompt text", lineNum)
		}

		items = append(items, item)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading JSONL: %w", err)
	}

	if len(items) == 0 {
		return nil, fmt.Errorf("dataset is empty")
	}

	return items, nil
}

func loadCSV(r io.Reader) ([]Item, error) {
	reader := csv.NewReader(r)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to parse CSV: %w", err)
	}

	if len(records) < 2 {
		return nil, fmt.Errorf("CSV must contain a header and at least one row")
	}

	header := records[0]
	colIdx := make(map[string]int)
	for i, col := range header {
		colIdx[strings.ToLower(strings.TrimSpace(col))] = i
	}

	promptIdx, hasPrompt := colIdx["prompt"]
	if !hasPrompt {
		return nil, fmt.Errorf("CSV header must contain a 'prompt' column")
	}

	idIdx, hasID := colIdx["id"]
	catIdx, hasCat := colIdx["category"]
	refIdx, hasRef := colIdx["reference"]
	sysIdx, hasSys := colIdx["system_prompt"]

	var items []Item
	for i, row := range records[1:] {
		prompt := strings.TrimSpace(row[promptIdx])
		if prompt == "" {
			continue
		}

		item := Item{
			Prompt: prompt,
		}

		if hasID && idIdx < len(row) && strings.TrimSpace(row[idIdx]) != "" {
			item.ID = strings.TrimSpace(row[idIdx])
		} else {
			item.ID = fmt.Sprintf("prompt_%04d", i+1)
		}

		if hasCat && catIdx < len(row) {
			item.Category = strings.TrimSpace(row[catIdx])
		}
		if hasRef && refIdx < len(row) {
			item.Reference = strings.TrimSpace(row[refIdx])
		}
		if hasSys && sysIdx < len(row) {
			item.SystemPrompt = strings.TrimSpace(row[sysIdx])
		}

		items = append(items, item)
	}

	if len(items) == 0 {
		return nil, fmt.Errorf("no valid prompt items found in CSV")
	}

	return items, nil
}

func loadYAML(r io.Reader) ([]Item, error) {
	var items []Item
	decoder := yaml.NewDecoder(r)
	if err := decoder.Decode(&items); err != nil {
		return nil, fmt.Errorf("failed to decode YAML dataset: %w", err)
	}

	for i := range items {
		if items[i].ID == "" {
			items[i].ID = fmt.Sprintf("prompt_%04d", i+1)
		}
		if strings.TrimSpace(items[i].Prompt) == "" {
			return nil, fmt.Errorf("item %d: empty prompt text", i+1)
		}
	}

	if len(items) == 0 {
		return nil, fmt.Errorf("dataset is empty")
	}

	return items, nil
}
