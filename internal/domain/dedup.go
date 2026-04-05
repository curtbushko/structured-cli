// Package domain contains the core domain types for structured-cli.
package domain

import "fmt"

// DedupConfig holds configuration for deduplication.
// It controls whether deduplication of repeated items is enabled.
type DedupConfig struct {
	// Enabled indicates whether deduplication is active.
	Enabled bool
}

// NewDedupConfig creates a new DedupConfig with default values.
// By default, deduplication is enabled.
func NewDedupConfig() DedupConfig {
	return DedupConfig{
		Enabled: true,
	}
}

// DedupResult represents the result of a deduplication operation.
// It contains counts of original and deduplicated items and the reduction percentage.
type DedupResult struct {
	// OriginalCount is the number of items before deduplication.
	OriginalCount int `json:"original_count"`

	// DedupedCount is the number of items after deduplication.
	DedupedCount int `json:"deduped_count"`

	// Reduction is the percentage reduction as a string (e.g., "50%").
	Reduction string `json:"reduction"`
}

// NewDedupResult creates a new DedupResult with the given original and deduped counts.
// It calculates the reduction percentage automatically.
func NewDedupResult(original, deduped int) DedupResult {
	var reduction string
	if original == 0 {
		reduction = "0%"
	} else {
		pct := ((original - deduped) * 100) / original
		reduction = fmt.Sprintf("%d%%", pct)
	}
	return DedupResult{
		OriginalCount: original,
		DedupedCount:  deduped,
		Reduction:     reduction,
	}
}

// DedupItem represents a deduplicated item with a count and sample.
// It wraps items that have been deduplicated, showing how many duplicates
// were found and providing a representative sample.
type DedupItem struct {
	// Count is the number of duplicate items found.
	Count int `json:"count"`

	// Sample is a representative example of the duplicated item.
	Sample any `json:"sample"`
}

// NewDedupItem creates a new DedupItem with the given count and sample.
func NewDedupItem(count int, sample any) DedupItem {
	return DedupItem{
		Count:  count,
		Sample: sample,
	}
}
