package ports

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/curtbushko/structured-cli/internal/domain"
)

// mockDeduplicator is a mock implementation of Deduplicator for testing.
type mockDeduplicator struct {
	dedupeData   any
	dedupeResult domain.DedupResult
}

func (m *mockDeduplicator) Dedupe(data any) (any, domain.DedupResult) {
	return m.dedupeData, m.dedupeResult
}

// TestDeduplicator_Interface tests that Deduplicator interface
// can be implemented and its methods are callable with correct signatures.
func TestDeduplicator_Interface(t *testing.T) {
	// Given: Create mock implementation of Deduplicator
	mock := &mockDeduplicator{
		dedupeData:   []string{"item1", "item2"},
		dedupeResult: domain.NewDedupResult(5, 2),
	}

	// Verify it implements the interface
	var deduplicator Deduplicator = mock
	require.NotNil(t, deduplicator)

	// When: Call Dedupe method
	data, result := deduplicator.Dedupe([]string{"item1", "item1", "item2", "item2", "item2"})

	// Then: Method is callable and returns expected values
	assert.NotNil(t, data)
	assert.Equal(t, 5, result.OriginalCount)
	assert.Equal(t, 2, result.DedupedCount)
	assert.Equal(t, "60%", result.Reduction)
}

// TestDeduplicator_MockImplements tests that a mock can implement the interface.
func TestDeduplicator_MockImplements(t *testing.T) {
	// Given: Create mock implementation
	mock := &mockDeduplicator{
		dedupeData:   map[string]any{"key": "value"},
		dedupeResult: domain.NewDedupResult(10, 5),
	}

	// When: Assign to interface type
	var deduplicator Deduplicator = mock

	// Then: Assignment succeeds (interface is implemented)
	require.NotNil(t, deduplicator)
}

// TestDeduplicator_DedupeSignature tests that Dedupe has
// the correct signature: (data any) (any, domain.DedupResult).
func TestDeduplicator_DedupeSignature(t *testing.T) {
	// Given: Create mock implementation
	mock := &mockDeduplicator{
		dedupeData:   nil,
		dedupeResult: domain.NewDedupResult(0, 0),
	}

	var deduplicator Deduplicator = mock

	// When: Call with different input types
	testCases := []struct {
		name string
		data any
	}{
		{"nil input", nil},
		{"string slice", []string{"a", "b", "c"}},
		{"map input", map[string]int{"count": 5}},
		{"struct input", struct{ Name string }{"test"}},
	}

	// Then: All calls work without panic and return correct types
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			data, result := deduplicator.Dedupe(tc.data)
			// Verify return types are correct (any and domain.DedupResult)
			_ = data // any type
			assert.IsType(t, domain.DedupResult{}, result)
		})
	}
}

// TestDeduplicator_DedupeReturnsDedupResult tests that Dedupe
// returns domain.DedupResult with correct field values.
func TestDeduplicator_DedupeReturnsDedupResult(t *testing.T) {
	// Given: Create mock implementation with specific result
	expectedResult := domain.NewDedupResult(100, 25)
	mock := &mockDeduplicator{
		dedupeData:   []string{"deduped"},
		dedupeResult: expectedResult,
	}

	var deduplicator Deduplicator = mock

	// When: Call Dedupe
	_, result := deduplicator.Dedupe("some input")

	// Then: Result is DedupResult with expected values
	assert.Equal(t, 100, result.OriginalCount)
	assert.Equal(t, 25, result.DedupedCount)
	assert.Equal(t, "75%", result.Reduction)
}
