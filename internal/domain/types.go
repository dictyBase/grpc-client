package domain

import (
	"fmt"
	"strings"

	E "github.com/IBM/fp-go/v2/either"
	F "github.com/IBM/fp-go/v2/function"
	IOE "github.com/IBM/fp-go/v2/ioeither"
	annotation "github.com/dictyBase/go-genproto/dictybaseapis/annotation"
	stockpb "github.com/dictyBase/go-genproto/dictybaseapis/stock"
	"google.golang.org/grpc"
)

// ListPlasmidsContext is the marker context for list plasmids operations
// following the modware-import pattern from plasmid_ontology.go
type ListPlasmidsContext struct{}

// ListPlasmidsConfig holds configuration for listing plasmids from the Stock API
// Embeds ListPlasmidsContext following the modware-import pattern
type ListPlasmidsConfig struct {
	ListPlasmidsContext
	ServerAddr string
	Port       string
	Filter     string
	Limit      int64
	Cursor     int64
	PlasmidID  string
	StrainID   string
	StrainType string
}

// StrainFilterAllowed holds the allowed strain type values for filtering.
var StrainFilterAllowed = []string{"REMI-seq", "general strain", "bacterial strain", "all"}

// WithConnection enriches ListPlasmidsConfig with gRPC connection
// Pattern from modware-import plasmid_ontology.go WithPlasmid (lines 127-130)
type WithConnection struct {
	ListPlasmidsConfig
	Connection *grpc.ClientConn
}

// PlasmidResult represents a processed plasmid from the API
type PlasmidResult struct {
	ID      string
	Name    string
	Summary string
}

// StrainResult represents a processed strain from the API
type StrainResult struct {
	ID                  string
	Label               string
	CreatedBy           string
	Species             string
	DictyStrainProperty string
}

// AnnotationResult represents a processed annotation from the API
type AnnotationResult struct {
	ID        string
	EntryID   string
	Tag       string
	Ontology  string
	Value     string
	CreatedBy string
	Version   int64
}

// Type aliases for IOEither-based functional composition
// Following modware-import pattern for cleaner function signatures

// ConfigIOE represents an IO operation that produces a ListPlasmidsConfig or an error
type ConfigIOE = IOE.IOEither[error, ListPlasmidsConfig]

// ResultsIOE represents an IO operation that produces a slice of PlasmidResult or an error
type ResultsIOE = IOE.IOEither[error, []PlasmidResult]

// CollectionIOE represents an IO operation that produces a PlasmidCollection or an error
type CollectionIOE = IOE.IOEither[error, *stockpb.PlasmidCollection]

// StrainCollectionIOE represents an IO operation that produces a StrainCollection or an error
type StrainCollectionIOE = IOE.IOEither[error, *stockpb.StrainCollection]

// StrainResultsIOE represents an IO operation that produces a slice of StrainResult or an error
type StrainResultsIOE = IOE.IOEither[error, []StrainResult]

// AnnotationCollectionIOE represents an IO operation that produces a TaggedAnnotationCollection or an error
type AnnotationCollectionIOE = IOE.IOEither[error, *annotation.TaggedAnnotationCollection]

// AnnotationResultsIOE represents an IO operation that produces a slice of AnnotationResult or an error
type AnnotationResultsIOE = IOE.IOEither[error, []AnnotationResult]

const MaxSummaryWords = 30

// ToEither executes an IOEither effect and returns the resulting Either.
func ToEither[ERR, A any](ioe IOE.IOEither[ERR, A]) E.Either[ERR, A] {
	return ioe()
}

// TruncateWords truncates a string to the specified number of words.
func TruncateWords(maxWords int, s string) string {
	words := strings.Fields(s)
	if len(words) > maxWords {
		return strings.Join(words[:maxWords], " ") + "..."
	}
	return s
}

var trun30words = F.Curry2(TruncateWords)(MaxSummaryWords)

// FormatPlasmidRecord formats a single plasmid result as a display string.
func FormatPlasmidRecord(p PlasmidResult) string {
	return fmt.Sprintf(
		"ID: %s | Name: %s | Summary: %s",
		p.ID,
		p.Name,
		trun30words(p.Summary),
	)
}

// FormatAnnotationRecord formats a single annotation result as a display string.
func FormatAnnotationRecord(a AnnotationResult) string {
	return fmt.Sprintf(
		"ID: %s | Entry: %s | Tag: %s | Ontology: %s | Value: %s | By: %s | v%d",
		a.ID,
		a.EntryID,
		a.Tag,
		a.Ontology,
		trun30words(a.Value),
		a.CreatedBy,
		a.Version,
	)
}
