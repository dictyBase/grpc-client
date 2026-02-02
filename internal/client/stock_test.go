package client

import (
	"testing"

	stockpb "github.com/dictyBase/go-genproto/dictybaseapis/stock"
	"github.com/dictyBase/learn-golang/grpc/plasmid/goldenbraid/internal/domain"
	"github.com/stretchr/testify/require"
)

func makeTestPlasmid(id, name, summary string) *stockpb.PlasmidCollection_Data {
	return &stockpb.PlasmidCollection_Data{
		Id: id,
		Attributes: &stockpb.PlasmidAttributes{
			Name:    name,
			Summary: summary,
		},
	}
}

func TestToPlasmidResults(t *testing.T) {
	tests := []struct {
		name     string
		input    *stockpb.PlasmidCollection
		expected []domain.PlasmidResult
	}{
		{
			name: "empty collection",
			input: &stockpb.PlasmidCollection{
				Data: []*stockpb.PlasmidCollection_Data{},
			},
			expected: []domain.PlasmidResult{},
		},
		{
			name: "single plasmid",
			input: &stockpb.PlasmidCollection{
				Data: []*stockpb.PlasmidCollection_Data{
					makeTestPlasmid("plas001", "TestPlasmid", "A test plasmid"),
				},
			},
			expected: []domain.PlasmidResult{
				{ID: "plas001", Name: "TestPlasmid", Summary: "A test plasmid"},
			},
		},
		{
			name: "multiple plasmids",
			input: &stockpb.PlasmidCollection{
				Data: []*stockpb.PlasmidCollection_Data{
					makeTestPlasmid("plas001", "Plasmid1", "First plasmid"),
					makeTestPlasmid("plas002", "Plasmid2", "Second plasmid"),
					makeTestPlasmid("plas003", "Plasmid3", "Third plasmid"),
				},
			},
			expected: []domain.PlasmidResult{
				{ID: "plas001", Name: "Plasmid1", Summary: "First plasmid"},
				{ID: "plas002", Name: "Plasmid2", Summary: "Second plasmid"},
				{ID: "plas003", Name: "Plasmid3", Summary: "Third plasmid"},
			},
		},
		{
			name: "plasmids with empty summary",
			input: &stockpb.PlasmidCollection{
				Data: []*stockpb.PlasmidCollection_Data{
					makeTestPlasmid("plas004", "NoSummaryPlasmid", ""),
				},
			},
			expected: []domain.PlasmidResult{
				{ID: "plas004", Name: "NoSummaryPlasmid", Summary: ""},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results := ToPlasmidResults(tt.input)
			require.Equal(t, tt.expected, results)
		})
	}
}

func BenchmarkToPlasmidResults(b *testing.B) {
	collection := &stockpb.PlasmidCollection{
		Data: make([]*stockpb.PlasmidCollection_Data, 100),
	}

	for i := 0; i < 100; i++ {
		collection.Data[i] = &stockpb.PlasmidCollection_Data{
			Id: "plas" + string(rune(i)),
			Attributes: &stockpb.PlasmidAttributes{
				Name:    "Plasmid" + string(rune(i)),
				Summary: "A plasmid for testing purposes",
			},
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ToPlasmidResults(collection)
	}
}
