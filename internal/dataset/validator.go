package dataset

import (
	"fmt"
	"strings"
)

// Validate checks that items have unique IDs and non-empty prompt bodies.
func Validate(items []Item) error {
	if len(items) == 0 {
		return fmt.Errorf("dataset contains 0 items")
	}

	seenIDs := make(map[string]int)
	for i, item := range items {
		if strings.TrimSpace(item.Prompt) == "" {
			return fmt.Errorf("item at index %d has empty prompt", i)
		}

		if prevIndex, exists := seenIDs[item.ID]; exists {
			return fmt.Errorf("duplicate prompt ID %q detected at indices %d and %d", item.ID, prevIndex, i)
		}
		seenIDs[item.ID] = i
	}

	return nil
}
