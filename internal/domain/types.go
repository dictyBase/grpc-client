package domain

import (
	IOE "github.com/IBM/fp-go/v2/ioeither"
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
}

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

// Type aliases for IOEither-based functional composition
// Following modware-import pattern for cleaner function signatures

// ConfigIOE represents an IO operation that produces a ListPlasmidsConfig or an error
type ConfigIOE = IOE.IOEither[error, ListPlasmidsConfig]

// ResultsIOE represents an IO operation that produces a slice of PlasmidResult or an error
type ResultsIOE = IOE.IOEither[error, []PlasmidResult]

// CollectionIOE represents an IO operation that produces a PlasmidCollection or an error
type CollectionIOE = IOE.IOEither[error, *stockpb.PlasmidCollection]
