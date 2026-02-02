package aggregation

import (
	"fmt"

	"github.com/dictyBase/learn-golang/grpc/plasmid/goldenbraid/internal/fputil"
	"github.com/dictyBase/learn-golang/grpc/plasmid/goldenbraid/internal/types"
)

var trun30words = fputil.TruncateWords(30)

// FormatPlasmidRecord formats a single plasmid result as a display string.
func FormatPlasmidRecord(p types.PlasmidResult) string {
	return fmt.Sprintf(
		"ID: %s | Name: %s | Summary: %s",
		p.ID,
		p.Name,
		trun30words(p.Summary),
	)
}
