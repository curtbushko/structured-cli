// Package application contains the business logic and use cases for structured-cli.
// This layer orchestrates domain logic and depends on ports (interfaces) - never adapters.
package application

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/curtbushko/structured-cli/internal/domain"
)

// Deduper performs two-stage deduplication on parsed command output.
// Stage 1: Raw text - collapse identical lines before parsing.
// Stage 2: JSON objects - collapse identical objects within same array level.
//
// It implements the ports.Deduplicator interface.
type Deduper struct {
	config domain.DedupConfig
}

// NewDeduper creates a new Deduper with default configuration.
// The deduper is enabled by default.
func NewDeduper() *Deduper {
	return &Deduper{
		config: domain.NewDedupConfig(),
	}
}

// NewDeduperWithConfig creates a new Deduper with custom configuration.
func NewDeduperWithConfig(config domain.DedupConfig) *Deduper {
	return &Deduper{
		config: config,
	}
}

// Dedupe takes any data structure and returns the deduplicated data along with statistics.
// It handles:
// - Raw text strings: collapses identical lines
// - JSON arrays: collapses identical objects at same level
// - Other data: passes through unchanged
//
// Returns the deduplicated data and statistics about the operation.
func (d *Deduper) Dedupe(data any) (any, domain.DedupResult) {
	// Handle disabled deduper
	if !d.config.Enabled {
		return data, domain.NewDedupResult(0, 0)
	}

	// Handle nil data
	if data == nil {
		return nil, domain.NewDedupResult(0, 0)
	}

	// Handle string data (raw text deduplication)
	if str, ok := data.(string); ok {
		return d.dedupeRawText(str)
	}

	// Handle array data (JSON array deduplication)
	if arr, ok := data.([]any); ok {
		return d.dedupeArray(arr)
	}

	// Handle map data (look for nested arrays)
	if m, ok := data.(map[string]any); ok {
		return d.dedupeMap(m)
	}

	// Pass through other data unchanged
	return data, domain.NewDedupResult(0, 0)
}

// dedupeRawText collapses identical lines in text, adding repeat counts.
func (d *Deduper) dedupeRawText(text string) (string, domain.DedupResult) {
	if text == "" {
		return text, domain.NewDedupResult(0, 0)
	}

	lines := strings.Split(text, "\n")
	originalCount := len(lines)

	// Group identical lines preserving order of first occurrence
	type lineGroup struct {
		line  string
		count int
	}

	seen := make(map[string]int) // maps line to index in groups
	var groups []lineGroup

	for _, line := range lines {
		if idx, exists := seen[line]; exists {
			groups[idx].count++
		} else {
			seen[line] = len(groups)
			groups = append(groups, lineGroup{line: line, count: 1})
		}
	}

	// Build deduplicated output
	var result []string
	for _, g := range groups {
		if g.count > 1 {
			result = append(result, fmt.Sprintf("%s (repeated %d times)", g.line, g.count))
		} else {
			result = append(result, g.line)
		}
	}

	dedupedCount := len(groups)
	return strings.Join(result, "\n"), domain.NewDedupResult(originalCount, dedupedCount)
}

// dedupeArray collapses identical objects at the same array level.
func (d *Deduper) dedupeArray(arr []any) ([]any, domain.DedupResult) {
	if len(arr) == 0 {
		return arr, domain.NewDedupResult(0, 0)
	}

	originalCount := len(arr)

	// Group identical objects preserving order of first occurrence
	type itemGroup struct {
		sample any
		count  int
	}

	var groups []itemGroup
	var serialized []string // for quick comparison

	for _, item := range arr {
		// First, recursively process any nested arrays in this item
		item = d.processNested(item)

		// Serialize for comparison
		itemJSON := d.serialize(item)

		// Check if we've seen this item before
		found := false
		for idx, s := range serialized {
			if s == itemJSON {
				groups[idx].count++
				found = true
				break
			}
		}

		if !found {
			serialized = append(serialized, itemJSON)
			groups = append(groups, itemGroup{sample: item, count: 1})
		}
	}

	// Build deduplicated output
	var result []any
	for _, g := range groups {
		if g.count > 1 {
			// Add count field to the sample
			if m, ok := g.sample.(map[string]any); ok {
				// Clone the map to avoid modifying the original
				newMap := make(map[string]any)
				for k, v := range m {
					newMap[k] = v
				}
				newMap["count"] = g.count
				result = append(result, newMap)
			} else {
				// For non-map items, wrap in a structure
				result = append(result, map[string]any{
					"sample": g.sample,
					"count":  g.count,
				})
			}
		} else {
			result = append(result, g.sample)
		}
	}

	dedupedCount := len(groups)
	return result, domain.NewDedupResult(originalCount, dedupedCount)
}

// dedupeMap processes a map looking for nested arrays to deduplicate.
func (d *Deduper) dedupeMap(m map[string]any) (map[string]any, domain.DedupResult) {
	result := make(map[string]any)
	var totalOriginal, totalDeduped int

	for k, v := range m {
		if arr, ok := v.([]any); ok {
			deduped, stats := d.dedupeArray(arr)
			result[k] = deduped
			totalOriginal += stats.OriginalCount
			totalDeduped += stats.DedupedCount
		} else if nested, ok := v.(map[string]any); ok {
			deduped, stats := d.dedupeMap(nested)
			result[k] = deduped
			totalOriginal += stats.OriginalCount
			totalDeduped += stats.DedupedCount
		} else {
			result[k] = v
		}
	}

	return result, domain.NewDedupResult(totalOriginal, totalDeduped)
}

// processNested recursively processes nested structures within an item.
func (d *Deduper) processNested(item any) any {
	switch v := item.(type) {
	case map[string]any:
		result := make(map[string]any)
		for k, val := range v {
			result[k] = d.processNested(val)
		}
		return result
	case []any:
		// Don't deduplicate nested arrays here - that's handled by dedupeMap
		// Just process each item recursively
		result := make([]any, len(v))
		for i, val := range v {
			result[i] = d.processNested(val)
		}
		return result
	default:
		return item
	}
}

// serialize converts an item to a JSON string for comparison.
func (d *Deduper) serialize(item any) string {
	// For comparison purposes, we need deterministic serialization
	bytes, err := json.Marshal(item)
	if err != nil {
		return fmt.Sprintf("%v", item)
	}
	return string(bytes)
}
