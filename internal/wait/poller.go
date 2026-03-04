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
// Passes PollContext forward on success so fetchCondition can be chained point-free.
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

// fetchCondition fetches the Job, extracts the terminal condition, and stores
// it in ctx.Condition so resolveState can read it point-free.
func fetchCondition(ctx PollContext) IOE.IOEither[error, PollContext] {
	setCondition := func(cond O.Option[JobState]) PollContext {
		ctx.Condition = cond
		return ctx
	}
	return F.Pipe2(
		FetchJob(ctx),
		IOE.Map[error](ExtractJobCondition),
		IOE.Map[error](setCondition),
	)
}

// resolveState reads ctx.Condition and either uses it directly (terminal) or
// fetches pods to classify (Pending vs Stuck), updating ctx.State.
func resolveState(ctx PollContext) IOE.IOEither[error, PollContext] {
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
	)(ctx.Condition)
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
// Each iteration: check timeout → fetch condition → resolve state → log → continue or return.
// All pipe steps are point-free — PollContext threads through without outer closures.
func pollUntilDone(ctx PollContext) IOE.IOEither[error, PollContext] {
	return F.Pipe4(
		checkTimeout(ctx),
		IOE.Chain(fetchCondition),
		IOE.Chain(resolveState),
		IOE.ChainFirstIOK[error](logState),
		IOE.Chain(continueOrReturn),
	)
}
