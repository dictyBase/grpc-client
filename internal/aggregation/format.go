package aggregation

import (
	"fmt"

	A "github.com/IBM/fp-go/v2/array"
	F "github.com/IBM/fp-go/v2/function"
	IO "github.com/IBM/fp-go/v2/io"
	"github.com/dictyBase/learn-golang/grpc/plasmid/goldenbraid/internal/fputil"
	"github.com/dictyBase/learn-golang/grpc/plasmid/goldenbraid/internal/types"
)

// ForEach executes an IO operation for each element of the array.
// Replaces the missing A.ForEach from fp-go library.
func ForEach[A any](f func(A) IO.IO[A]) func([]A) IO.IO[[]A] {
	return func(as []A) IO.IO[[]A] {
		return func() []A {
			for _, a := range as {
				f(a)()
			}
			return as
		}
	}
}

// Println prints a string to stdout followed by a newline.
// Replaces the missing IO.Println from fp-go library.
func Println(s string) IO.IO[string] {
	return func() string {
		fmt.Println(s)
		return s
	}
}

// FormatPlasmidRecord formats a single plasmid result as a display string.
// Pure function - no side effects.
func FormatPlasmidRecord(p types.PlasmidResult) string {
	summary := fputil.TruncateWords(p.Summary, 30)
	return fmt.Sprintf("ID: %s | Name: %s | Summary: %s", p.ID, p.Name, summary)
}

// PrintResults creates an IO operation that prints each plasmid result.
// Uses ForEach instead of for loop - pure fp-go pattern.
func PrintResults(results []types.PlasmidResult) IO.IO[[]types.PlasmidResult] {
	return func() []types.PlasmidResult {
		F.Pipe2(
			results,
			A.Map(FormatPlasmidRecord),
			ForEach(Println),
		)()
		return results
	}
}
