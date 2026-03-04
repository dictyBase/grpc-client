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
// Passes PollContext forward on success so evaluateOnce can be chained point-free.
var checkTimeout = func(ctx PollContext) IOE.IOEither[error, PollContext] {
	return func() E.Either[error, PollContext] {
		return E.FromPredicate(
			func(c PollContext) bool { return time.Now().Before(c.Deadline) },
			func(c PollContext) error {
				return fmt.Errorf("timeout waiting for job %s", c.Name)
			},
		)(ctx)
	}
}

// liftJobState lifts a JobState into IOEither.
// Used as the Some-case handler in O.Fold — passes terminal states through unchanged.
var liftJobState = func(s JobState) IOE.IOEither[error, JobState] {
	return IOE.Of[error](s)
}

// resolveState maps an Option[JobState] to an IOEither:
//
//	Some(state) → return that state directly.
//	None        → fetch pods and classify (Pending vs Stuck).
var resolveState = func(ctx PollContext) func(O.Option[JobState]) IOE.IOEither[error, JobState] {
	return O.Fold(
		func() IOE.IOEither[error, JobState] {
			return F.Pipe1(
				FetchPods(ctx),
				IOE.Map[error](ClassifyPodState),
			)
		},
		liftJobState,
	)
}

// evaluateOnce performs one evaluation cycle:
//  1. Fetch the Kubernetes Job.
//  2. Extract any terminal condition (Complete/Failed) → Option[JobState].
//  3. If None, fetch pods and classify (Pending vs Stuck).
var evaluateOnce = func(ctx PollContext) IOE.IOEither[error, JobState] {
	return F.Pipe2(
		FetchJob(ctx),
		IOE.Map[error](ExtractJobCondition),
		IOE.Chain(resolveState(ctx)),
	)
}

// sleepThenRetry sleeps for PollInterval then recurses into pollUntilDone.
// Declared as a func (not var) to break the init cycle with continueOrReturn and
// pollUntilDone — func declarations are always available and never participate in
// Go's variable initialization ordering.
func sleepThenRetry(ctx PollContext) IOE.IOEither[error, JobState] {
	return func() E.Either[error, JobState] {
		time.Sleep(PollInterval)
		return pollUntilDone(ctx)()
	}
}

// continueOrReturn decides whether to stop or keep polling.
//
//	terminal state (Complete/Failed/Stuck) → return it.
//	Pending                                → sleep PollInterval, then recurse.
//
// Declared as a func to break the init cycle.
func continueOrReturn(ctx PollContext) func(JobState) IOE.IOEither[error, JobState] {
	return func(state JobState) IOE.IOEither[error, JobState] {
		return F.Pipe2(
			state,
			O.FromPredicate(isTerminal),
			O.Fold(
				func() IOE.IOEither[error, JobState] { return sleepThenRetry(ctx) },
				liftJobState,
			),
		)
	}
}

// pollUntilDone is the recursive polling loop.
// Each iteration: check timeout → evaluate once → log state → continue or return.
// Declared as a func (not var) to break the three-way init cycle with
// sleepThenRetry and continueOrReturn.
func pollUntilDone(ctx PollContext) IOE.IOEither[error, JobState] {
	return F.Pipe3(
		checkTimeout(ctx),
		IOE.Chain(evaluateOnce),
		IOE.ChainFirstIOK[error](logPollState(ctx.Logger, ctx.Name)),
		IOE.Chain(continueOrReturn(ctx)),
	)
}
