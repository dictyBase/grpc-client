package wait

import (
	"log/slog"

	IO "github.com/IBM/fp-go/v2/io"
)

// logPollState returns a logging IO action for use with IOE.ChainFirstIOK.
// Logs the current poll state without altering the value flowing through the pipe.
var logPollState = func(logger *slog.Logger, jobName string) func(JobState) IO.IO[struct{}] {
	return func(state JobState) IO.IO[struct{}] {
		return func() struct{} {
			logger.Info("polling",
				"job", jobName,
				"state", string(state),
			)
			return struct{}{}
		}
	}
}
