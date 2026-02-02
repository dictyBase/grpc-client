package aggregation

import (
	"fmt"

	"github.com/dictyBase/learn-golang/grpc/plasmid/goldenbraid/internal/fputil"
	"github.com/dictyBase/learn-golang/grpc/plasmid/goldenbraid/internal/types"
)

// FormatPlasmidRecord formats a single plasmid result as a display string.
// Pure function - no side effects.
func FormatPlasmidRecord(p types.PlasmidResult) string {
	summary := fputil.TruncateWords(p.Summary, 30)
	return fmt.Sprintf("ID: %s | Name: %s | Summary: %s", p.ID, p.Name, summary)
}
