package wait

import (
	"fmt"
	"time"

	E "github.com/IBM/fp-go/v2/either"
	F "github.com/IBM/fp-go/v2/function"
	IOE "github.com/IBM/fp-go/v2/ioeither"
	O "github.com/IBM/fp-go/v2/option"
)

// checkTimeout fails with a timeout error if the deadline has passed.
// Passes PollContext forward on success so evaluate can be chained point-free.
func checkTimeout(ctx PollContext) IOE.IOEither[error, PollContext] {
	return func() E.Either[error, PollContext] {
		return E.FromPredicate(
			func(c PollContext) bool { return time.Now().Before(c.Deadline) },
			func(c PollContext) error {
				return fmt.Errorf("timeout waiting for job %s", c.Name)
			},
		)(ctx)
	}
}

// resolveState maps an Option[JobState] to an IOEither[error, PollContext]:
//
//	Some(state) → embed state in ctx and return directly.
//	None        → fetch pods, classify, embed result in ctx.
func resolveState(ctx PollContext) func(O.Option[JobState]) IOE.IOEither[error, PollContext] {
	withState := func(state JobState) PollContext {
		ctx.State = state
		return ctx
	}
	return O.Fold(
		func() IOE.IOEither[error, PollContext] {
			return F.Pipe2(
				FetchPods(ctx),
				IOE.Map[error](ClassifyPodState),
				IOE.Map[error](withState),
			)
		},
		func(state JobState) IOE.IOEither[error, PollContext] {
			return IOE.Of[error](withState(state))
		},
	)
}

// evaluate fetches the Job, classifies its state, and embeds the result in PollContext.
func evaluate(ctx PollContext) IOE.IOEither[error, PollContext] {
	return F.Pipe2(
		FetchJob(ctx),
		IOE.Map[error](ExtractJobCondition),
		IOE.Chain(resolveState(ctx)),
	)
}

// sleepThenRetry sleeps for PollInterval then recurses into pollUntilDone.
func sleepThenRetry(ctx PollContext) IOE.IOEither[error, PollContext] {
	return func() E.Either[error, PollContext] {
		time.Sleep(PollInterval)
		return pollUntilDone(ctx)()
	}
}

// continueOrReturn decides whether to stop or keep polling based on ctx.State.
//
//	terminal state (Complete/Failed/Stuck) → return ctx as-is.
//	Pending                                → sleep PollInterval, then recurse.
func continueOrReturn(ctx PollContext) IOE.IOEither[error, PollContext] {
	return F.Pipe2(
		ctx.State,
		O.FromPredicate(isTerminal),
		O.Fold(
			func() IOE.IOEither[error, PollContext] { return sleepThenRetry(ctx) },
			func(_ JobState) IOE.IOEither[error, PollContext] { return IOE.Of[error](ctx) },
		),
	)
}

// pollUntilDone is the recursive polling loop.
// Each iteration: check timeout → evaluate → log state → continue or return.
// All pipe steps are point-free — no closure over ctx inside the pipe body.
func pollUntilDone(ctx PollContext) IOE.IOEither[error, PollContext] {
	return F.Pipe3(
		checkTimeout(ctx),
		IOE.Chain(evaluate),
		IOE.ChainFirstIOK[error](logState),
		IOE.Chain(continueOrReturn),
	)
}
