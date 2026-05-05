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

const defaultTimeout = 60 * time.Second

// validateTerminalState succeeds for Complete; fails with an error for any other state.
var validateTerminalState = E.FromPredicate(
	isComplete,
	func(s JobState) error {
		return fmt.Errorf("job terminated in state: %s", s)
	},
)

// parseDuration converts a CLI timeout string to time.Duration.
// Falls back to defaultTimeout (60s) on parse failure.
func parseDuration(s string) time.Duration {
	return F.Pipe1(
		E.TryCatchError(time.ParseDuration(s)),
		E.GetOrElse(func(_ error) time.Duration { return defaultTimeout }),
	)
}

// extractState pulls the final JobState out of a completed PollContext.
func extractState(ctx PollContext) JobState { return ctx.State }

// JobAction is the urfave/cli v3 action for the wait-job subcommand.
//
// Pipeline (8 steps):
//  1. Build Params from CLI flags
//  2. Inject Kubernetes client (Bind)
//  3. Compute deadline and attach logger (Let)
//  4. Run polling loop (returns PollContext with final State)
//  5. Extract JobState from PollContext
//  6. Execute IOEither effect → Either
//  7. Validate terminal state (Complete → ok, else → error)
//  8. Fold Either → error
func JobAction(_ context.Context, cmd *cli.Command) error {
	return F.Pipe7(
		IOE.Of[error](Params{
			Name:       cmd.String("name"),
			Namespace:  cmd.String("namespace"),
			Timeout:    parseDuration(cmd.String("timeout")),
			Kubeconfig: cmd.String("kubeconfig"),
		}),
		IOE.Bind(SetClient, CreateK8sClient),
		IOE.Let[error](SetPollReady, computeDeadline),
		IOE.Chain(pollUntilDone),
		IOE.Map[error](extractState),
		fputil.ToEither,
		E.Chain(validateTerminalState),
		E.Fold(
			F.Identity[error],
			func(_ JobState) error { return nil },
		),
	)
}
