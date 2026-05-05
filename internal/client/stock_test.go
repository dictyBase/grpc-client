package client

import (
	"testing"

	annotationpb "github.com/dictyBase/go-genproto/dictybaseapis/annotation"
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

	for i := range 100 {
		collection.Data[i] = &stockpb.PlasmidCollection_Data{
			Id: "plas" + string(rune(i)),
			Attributes: &stockpb.PlasmidAttributes{
				Name:    "Plasmid" + string(rune(i)),
				Summary: "A plasmid for testing purposes",
			},
		}
	}

	b.ResetTimer()
	for range b.N {
		_ = ToPlasmidResults(collection)
	}
}

func makeTestStrainData(
	id, label, createdBy, species, property string,
) *stockpb.StrainCollection_Data {
	return &stockpb.StrainCollection_Data{
		Id: id,
		Attributes: &stockpb.StrainAttributes{
			Label:               label,
			CreatedBy:           createdBy,
			Species:             species,
			DictyStrainProperty: property,
		},
	}
}

func TestToStrainResults(t *testing.T) {
	tests := []struct {
		name     string
		input    *stockpb.StrainCollection
		expected []domain.StrainResult
	}{
		{
			name: "empty collection",
			input: &stockpb.StrainCollection{
				Data: []*stockpb.StrainCollection_Data{},
			},
			expected: []domain.StrainResult{},
		},
		{
			name: "single strain",
			input: &stockpb.StrainCollection{
				Data: []*stockpb.StrainCollection_Data{
					makeTestStrainData("str001", "AX4", "user1", "D. discoideum", "REMI-seq"),
				},
			},
			expected: []domain.StrainResult{
				{
					ID:                  "str001",
					Label:               "AX4",
					CreatedBy:           "user1",
					Species:             "D. discoideum",
					DictyStrainProperty: "REMI-seq",
				},
			},
		},
		{
			name: "multiple strains",
			input: &stockpb.StrainCollection{
				Data: []*stockpb.StrainCollection_Data{
					makeTestStrainData("str001", "AX4", "user1", "D. discoideum", "REMI-seq"),
					makeTestStrainData("str002", "AX5", "user2", "D. discoideum", "general strain"),
				},
			},
			expected: []domain.StrainResult{
				{
					ID:                  "str001",
					Label:               "AX4",
					CreatedBy:           "user1",
					Species:             "D. discoideum",
					DictyStrainProperty: "REMI-seq",
				},
				{
					ID:                  "str002",
					Label:               "AX5",
					CreatedBy:           "user2",
					Species:             "D. discoideum",
					DictyStrainProperty: "general strain",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results := ToStrainResults(tt.input)
			require.Equal(t, tt.expected, results)
		})
	}
}

func TestBuildStrainFilter(t *testing.T) {
	tests := []struct {
		name     string
		stype    string
		expected string
	}{
		{
			name:     "REMI-seq",
			stype:    "REMI-seq",
			expected: "ontology==dicty_strain_property;tag==REMI-seq",
		},
		{
			name:     "general strain",
			stype:    "general strain",
			expected: "ontology==dicty_strain_property;tag==general strain",
		},
		{
			name:     "all",
			stype:    "all",
			expected: "ontology==dicty_strain_property;tag==REMI-seq,tag==general strain,tag==bacterial strain",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildStrainFilter(tt.stype)
			require.Equal(t, tt.expected, got)
		})
	}
}

func makeTestAnnotationData(
	id, entryID, tag, ontology, value, createdBy string,
	version int64,
) *annotationpb.TaggedAnnotationCollection_Data {
	return &annotationpb.TaggedAnnotationCollection_Data{
		Id: id,
		Attributes: &annotationpb.TaggedAnnotationAttributes{
			EntryId:   entryID,
			Tag:       tag,
			Ontology:  ontology,
			Value:     value,
			CreatedBy: createdBy,
			Version:   version,
		},
	}
}

func TestToAnnotationResults(t *testing.T) {
	tests := []struct {
		name     string
		input    *annotationpb.TaggedAnnotationCollection
		expected []domain.AnnotationResult
	}{
		{
			name: "empty collection",
			input: &annotationpb.TaggedAnnotationCollection{
				Data: []*annotationpb.TaggedAnnotationCollection_Data{},
			},
			expected: []domain.AnnotationResult{},
		},
		{
			name: "single annotation",
			input: &annotationpb.TaggedAnnotationCollection{
				Data: []*annotationpb.TaggedAnnotationCollection_Data{
					makeTestAnnotationData("ann001", "DDB_G123", "GO:0005634", "cellular_component", "nucleus", "user@test.org", 1),
				},
			},
			expected: []domain.AnnotationResult{
				{
					ID:        "ann001",
					EntryID:   "DDB_G123",
					Tag:       "GO:0005634",
					Ontology:  "cellular_component",
					Value:     "nucleus",
					CreatedBy: "user@test.org",
					Version:   1,
				},
			},
		},
		{
			name: "multiple annotations",
			input: &annotationpb.TaggedAnnotationCollection{
				Data: []*annotationpb.TaggedAnnotationCollection_Data{
					makeTestAnnotationData("ann001", "DDB_G123", "GO:0005634", "cellular_component", "nucleus", "user@test.org", 1),
					makeTestAnnotationData("ann002", "DDB_G456", "GO:0005737", "cellular_component", "cytoplasm", "curator@test.org", 3),
				},
			},
			expected: []domain.AnnotationResult{
				{
					ID:        "ann001",
					EntryID:   "DDB_G123",
					Tag:       "GO:0005634",
					Ontology:  "cellular_component",
					Value:     "nucleus",
					CreatedBy: "user@test.org",
					Version:   1,
				},
				{
					ID:        "ann002",
					EntryID:   "DDB_G456",
					Tag:       "GO:0005737",
					Ontology:  "cellular_component",
					Value:     "cytoplasm",
					CreatedBy: "curator@test.org",
					Version:   3,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results := ToAnnotationResults(tt.input)
			require.Equal(t, tt.expected, results)
		})
	}
}
