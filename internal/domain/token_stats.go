package domain

// SavingsCategory represents the quality level of token savings.
// It categorizes savings into Good, Warning, or Critical based on percentage thresholds.
type SavingsCategory int

const (
	// SavingsCategoryGood indicates savings greater than 50%.
	SavingsCategoryGood SavingsCategory = iota
	// SavingsCategoryWarning indicates savings between 20% and 50% (inclusive).
	SavingsCategoryWarning
	// SavingsCategoryCritical indicates savings less than 20%.
	SavingsCategoryCritical
)

// String returns the string representation of a SavingsCategory.
func (c SavingsCategory) String() string {
	switch c {
	case SavingsCategoryGood:
		return "good"
	case SavingsCategoryWarning:
		return "warning"
	case SavingsCategoryCritical:
		return "critical"
	default:
		return "unknown"
	}
}

// TokenStats holds raw versus structured byte counts with computed savings.
type TokenStats struct {
	// RawBytes is the byte count of the original raw output.
	RawBytes int
	// StructuredBytes is the byte count of the structured JSON output.
	StructuredBytes int
	// SavedBytes is the difference between raw and structured bytes.
	SavedBytes int
}

// NewTokenStats creates a new TokenStats with computed SavedBytes.
func NewTokenStats(rawBytes, structuredBytes int) TokenStats {
	return TokenStats{
		RawBytes:        rawBytes,
		StructuredBytes: structuredBytes,
		SavedBytes:      rawBytes - structuredBytes,
	}
}

// SavingsPercent returns the percentage of bytes saved.
// Returns 0.0 if RawBytes is zero to avoid division by zero.
func (t TokenStats) SavingsPercent() float64 {
	if t.RawBytes == 0 {
		return 0.0
	}
	return float64(t.SavedBytes) / float64(t.RawBytes) * 100
}

// CompressionStats holds compression metrics including ratio and savings category.
type CompressionStats struct {
	// RawBytes is the byte count of the original raw output.
	RawBytes int
	// StructuredBytes is the byte count of the structured JSON output.
	StructuredBytes int
}

// NewCompressionStats creates a new CompressionStats.
func NewCompressionStats(rawBytes, structuredBytes int) CompressionStats {
	return CompressionStats{
		RawBytes:        rawBytes,
		StructuredBytes: structuredBytes,
	}
}

// Ratio returns the compression ratio (raw / structured).
// Returns 0.0 if StructuredBytes is zero to avoid division by zero.
func (c CompressionStats) Ratio() float64 {
	if c.StructuredBytes == 0 {
		return 0.0
	}
	return float64(c.RawBytes) / float64(c.StructuredBytes)
}

// SavingsPercent returns the percentage of bytes saved.
// Returns 0.0 if RawBytes is zero to avoid division by zero.
func (c CompressionStats) SavingsPercent() float64 {
	if c.RawBytes == 0 {
		return 0.0
	}
	saved := c.RawBytes - c.StructuredBytes
	return float64(saved) / float64(c.RawBytes) * 100
}

// SavingsCategory returns the quality category based on savings percentage.
// - Good: > 50%
// - Warning: 20-50% (inclusive)
// - Critical: < 20%
func (c CompressionStats) SavingsCategory() SavingsCategory {
	percent := c.SavingsPercent()
	switch {
	case percent > 50:
		return SavingsCategoryGood
	case percent >= 20:
		return SavingsCategoryWarning
	default:
		return SavingsCategoryCritical
	}
}
