package aggregation

import (
	"fmt"

	"github.com/dictyBase/learn-golang/grpc/plasmid/goldenbraid/internal/domain"
	"github.com/dictyBase/learn-golang/grpc/plasmid/goldenbraid/internal/fputil"
)

const MaxSummaryWords = 30

var trun30words = fputil.TruncateWords(MaxSummaryWords)

// FormatPlasmidRecord formats a single plasmid result as a display string.
func FormatPlasmidRecord(p domain.PlasmidResult) string {
	return fmt.Sprintf(
		"ID: %s | Name: %s | Summary: %s",
		p.ID,
		p.Name,
		trun30words(p.Summary),
	)
}
