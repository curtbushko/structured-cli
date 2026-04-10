package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/curtbushko/structured-cli/internal/domain"
)

func TestTokenStats_New(t *testing.T) {
	// given: raw=1000, structured=400
	rawBytes := 1000
	structuredBytes := 400

	// when: NewTokenStats(1000, 400)
	stats := domain.NewTokenStats(rawBytes, structuredBytes)

	// then: RawBytes==1000, StructuredBytes==400, SavedBytes==600
	assert.Equal(t, rawBytes, stats.RawBytes)
	assert.Equal(t, structuredBytes, stats.StructuredBytes)
	assert.Equal(t, 600, stats.SavedBytes)
}

func TestTokenStats_SavingsPercent(t *testing.T) {
	// given: TokenStats with RawBytes=1000, StructuredBytes=400
	stats := domain.NewTokenStats(1000, 400)

	// when: call SavingsPercent()
	percent := stats.SavingsPercent()

	// then: returns 60.0
	assert.InDelta(t, 60.0, percent, 0.001)
}

func TestTokenStats_SavingsPercent_ZeroRaw(t *testing.T) {
	// given: TokenStats with RawBytes=0
	stats := domain.NewTokenStats(0, 0)

	// when: call SavingsPercent()
	percent := stats.SavingsPercent()

	// then: returns 0.0 (no division by zero)
	assert.InDelta(t, 0.0, percent, 0.001)
}

func TestCompressionStats_Ratio(t *testing.T) {
	// given: CompressionStats with RawBytes=1000, StructuredBytes=400
	stats := domain.NewCompressionStats(1000, 400)

	// when: call Ratio()
	ratio := stats.Ratio()

	// then: returns 2.5 (1000/400)
	assert.InDelta(t, 2.5, ratio, 0.001)
}

func TestCompressionStats_Ratio_ZeroStructured(t *testing.T) {
	// given: CompressionStats with StructuredBytes=0
	stats := domain.NewCompressionStats(1000, 0)

	// when: call Ratio()
	ratio := stats.Ratio()

	// then: returns 0.0 (no division by zero)
	assert.InDelta(t, 0.0, ratio, 0.001)
}

func TestCompressionStats_SavingsCategory_Good(t *testing.T) {
	// given: SavingsPercent > 50 (raw=1000, structured=400 => 60% savings)
	stats := domain.NewCompressionStats(1000, 400)

	// when: call SavingsCategory()
	category := stats.SavingsCategory()

	// then: returns SavingsCategoryGood
	assert.Equal(t, domain.SavingsCategoryGood, category)
}

func TestCompressionStats_SavingsCategory_Warning(t *testing.T) {
	// given: SavingsPercent between 20 and 50 (raw=1000, structured=700 => 30% savings)
	stats := domain.NewCompressionStats(1000, 700)

	// when: call SavingsCategory()
	category := stats.SavingsCategory()

	// then: returns SavingsCategoryWarning
	assert.Equal(t, domain.SavingsCategoryWarning, category)
}

func TestCompressionStats_SavingsCategory_Critical(t *testing.T) {
	// given: SavingsPercent < 20 (raw=1000, structured=900 => 10% savings)
	stats := domain.NewCompressionStats(1000, 900)

	// when: call SavingsCategory()
	category := stats.SavingsCategory()

	// then: returns SavingsCategoryCritical
	assert.Equal(t, domain.SavingsCategoryCritical, category)
}

func TestCompressionStats_SavingsCategory_BoundaryAt50(t *testing.T) {
	// given: SavingsPercent exactly 50 (raw=1000, structured=500)
	stats := domain.NewCompressionStats(1000, 500)

	// when: call SavingsCategory()
	category := stats.SavingsCategory()

	// then: returns SavingsCategoryWarning (50% is boundary, exclusive of Good)
	assert.Equal(t, domain.SavingsCategoryWarning, category)
}

func TestCompressionStats_SavingsCategory_BoundaryAt20(t *testing.T) {
	// given: SavingsPercent exactly 20 (raw=1000, structured=800)
	stats := domain.NewCompressionStats(1000, 800)

	// when: call SavingsCategory()
	category := stats.SavingsCategory()

	// then: returns SavingsCategoryWarning (20% is boundary, inclusive in Warning)
	assert.Equal(t, domain.SavingsCategoryWarning, category)
}

func TestSavingsCategory_String(t *testing.T) {
	tests := []struct {
		name     string
		category domain.SavingsCategory
		want     string
	}{
		{
			name:     "Good",
			category: domain.SavingsCategoryGood,
			want:     "good",
		},
		{
			name:     "Warning",
			category: domain.SavingsCategoryWarning,
			want:     "warning",
		},
		{
			name:     "Critical",
			category: domain.SavingsCategoryCritical,
			want:     "critical",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, tt.category.String())
		})
	}
}
