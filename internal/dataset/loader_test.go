package dataset_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/alexandrmotologa/evalarena/internal/dataset"
)

func TestLoadDataset_JSONL(t *testing.T) {
	tempDir := t.TempDir()
	path := filepath.Join(tempDir, "test.jsonl")

	content := `{"id": "t1", "category": "math", "prompt": "What is 2+2?", "reference": "4"}
{"id": "t2", "category": "code", "prompt": "Write hello world in Go"}
`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test jsonl: %v", err)
	}

	items, err := dataset.LoadDataset(path, "jsonl")
	if err != nil {
		t.Fatalf("unexpected load error: %v", err)
	}

	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
	if items[0].ID != "t1" || items[0].Category != "math" {
		t.Errorf("unexpected item 0: %+v", items[0])
	}
	if err := dataset.Validate(items); err != nil {
		t.Errorf("validation failed: %v", err)
	}
}

func TestLoadDataset_CSV(t *testing.T) {
	tempDir := t.TempDir()
	path := filepath.Join(tempDir, "test.csv")

	content := `id,category,prompt,reference
c1,support,How do I cancel?,Go to settings
c2,support,Where is invoice?,Check billing tab
`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test csv: %v", err)
	}

	items, err := dataset.LoadDataset(path, "csv")
	if err != nil {
		t.Fatalf("unexpected load error: %v", err)
	}

	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
	if items[0].ID != "c1" || items[1].ID != "c2" {
		t.Errorf("unexpected IDs: %s, %s", items[0].ID, items[1].ID)
	}
}
