package aggregation

// PlasmidResult represents a processed plasmid from the API.
type PlasmidResult struct {
	ID      string
	Name    string
	Summary string
}

// StrainResult represents a processed strain from the API.
type StrainResult struct {
	ID                  string
	Label               string
	CreatedBy           string
	Species             string
	DictyStrainProperty string
}

// AnnotationResult represents a processed annotation from the API.
type AnnotationResult struct {
	ID        string
	EntryID   string
	Tag       string
	Ontology  string
	Value     string
	CreatedBy string
	Version   int64
}

// StrainFilterAllowed holds the allowed strain type values for filtering.
var StrainFilterAllowed = []string{"REMI-seq", "general strain", "bacterial strain", "all"}
