package fputil

import (
	"strings"

	E "github.com/IBM/fp-go/v2/either"
	F "github.com/IBM/fp-go/v2/function"
	IOE "github.com/IBM/fp-go/v2/ioeither"
	O "github.com/IBM/fp-go/v2/option"
)

// ToEither executes an IOEither effect and returns the resulting Either.
// This is a helper to convert from IOEither (deferred computation) to Either (computed result).
// Pattern from modware-import/internal/fputil.
func ToEither[ERR, A any](ioe IOE.IOEither[ERR, A]) E.Either[ERR, A] {
	return ioe()
}

// TruncateWords truncates a string to the specified number of words.
// Pure function using O.Fold to avoid if statements.
var TruncateWords = F.Curry2(func(maxWords int, s string) string {
	words := strings.Fields(s)

	// Use Option monad to handle empty/short cases
	return F.Pipe2(
		words,
		O.FromPredicate(func(ws []string) bool {
			return len(ws) > maxWords
		}),
		O.Fold(
			func() string { return s }, // None: return original
			func(ws []string) string { // Some: truncate
				return strings.Join(ws[:maxWords], " ") + "..."
			},
		),
	)
})
