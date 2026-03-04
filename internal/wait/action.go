package wait

import (
	"context"
	"fmt"
	"time"

	E "github.com/IBM/fp-go/v2/either"
	F "github.com/IBM/fp-go/v2/function"
	IOE "github.com/IBM/fp-go/v2/ioeither"
	"github.com/dictyBase/learn-golang/grpc/plasmid/goldenbraid/internal/fputil"
	"github.com/urfave/cli/v3"
)

// validateTerminalState succeeds for Complete; fails with an error for any other state.
var validateTerminalState = E.FromPredicate(
	isComplete,
	func(s JobState) error {
		return fmt.Errorf("job terminated in state: %s", s)
	},
)

const defaultTimeout = 60 * time.Second

// parseDuration converts a CLI timeout string to time.Duration.
// Falls back to defaultTimeout (60s) on parse failure.
var parseDuration = func(s string) time.Duration {
	d, err := time.ParseDuration(s)
	if err != nil {
		return defaultTimeout
	}
	return d
}

// JobAction is the urfave/cli v3 action for the wait-job subcommand.
//
// Pipeline (7 steps):
//  1. Build Params from CLI flags
//  2. Inject Kubernetes client (Bind)
//  3. Compute deadline and attach logger (Let)
//  4. Run polling loop
//  5. Execute IOEither effect → Either
//  6. Validate terminal state (Complete → ok, else → error)
//  7. Fold Either → error
func JobAction(_ context.Context, cmd *cli.Command) error {
	return F.Pipe6(
		IOE.Of[error](Params{
			Name:      cmd.String("name"),
			Namespace: cmd.String("namespace"),
			Timeout:   parseDuration(cmd.String("timeout")),
		}),
		IOE.Bind(SetClient, CreateK8sClient),
		IOE.Let[error](SetPollReady, computeDeadline),
		IOE.Chain(pollUntilDone),
		fputil.ToEither[error, JobState],
		E.Chain(validateTerminalState),
		E.Fold(
			F.Identity[error],
			func(_ JobState) error { return nil },
		),
	)
}
