package aggregation_test

import (
	"testing"

	"github.com/dictyBase/learn-golang/grpc/plasmid/goldenbraid/internal/aggregation"
	"github.com/dictyBase/learn-golang/grpc/plasmid/goldenbraid/internal/domain"
	"github.com/dictyBase/learn-golang/grpc/plasmid/goldenbraid/internal/fputil"
	"github.com/stretchr/testify/require"
)

func TestTruncateWords(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		maxWords int
		expected string
	}{
		{
			name:     "short text - no truncation",
			input:    "This is a short summary",
			maxWords: 30,
			expected: "This is a short summary",
		},
		{
			name:     "long text - truncate to 5 words",
			input:    "This is a very long summary that needs to be truncated",
			maxWords: 5,
			expected: "This is a very long...",
		},
		{
			name:     "exact word count",
			input:    "One two three four five",
			maxWords: 5,
			expected: "One two three four five",
		},
		{
			name:     "empty string",
			input:    "",
			maxWords: 30,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := fputil.TruncateWords(tt.maxWords)(tt.input)
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestFormatPlasmidRecord(t *testing.T) {
	p := domain.PlasmidResult{
		ID:      "DBP0000001",
		Name:    "pDV-CFPC-5Hyg",
		Summary: "This is a test plasmid",
	}

	result := aggregation.FormatPlasmidRecord(p)
	expected := "ID: DBP0000001 | Name: pDV-CFPC-5Hyg | Summary: This is a test plasmid"
	require.Equal(t, expected, result)
}
