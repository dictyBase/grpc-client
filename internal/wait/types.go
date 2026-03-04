package wait

import (
	"log/slog"
	"os"
	"time"

	F "github.com/IBM/fp-go/v2/function"
	"k8s.io/client-go/kubernetes"
)

// JobState represents the observed state of a Kubernetes Job.
type JobState string

const (
	JobPending  JobState = "Pending"
	JobComplete JobState = "Complete"
	JobFailed   JobState = "Failed"
	JobStuck    JobState = "Stuck"
)

// PollInterval is the delay between polling cycles.
const PollInterval = 5 * time.Second

// Params holds CLI-supplied parameters for the wait-job command.
type Params struct {
	Name      string
	Namespace string
	Timeout   time.Duration
}

// WithClient enriches Params with an injected Kubernetes client.
type WithClient struct {
	Params
	Client kubernetes.Interface
}

// PollContext is the fully-enriched context used throughout the polling loop.
type PollContext struct {
	WithClient
	Logger   *slog.Logger
	Deadline time.Time
}

// setClient is a curried setter used with IOE.Bind to inject the K8s client.
var SetClient = F.Curry2(func(c kubernetes.Interface, p Params) WithClient {
	return WithClient{Params: p, Client: c}
})

// setPollReady is a curried setter used with IOE.Let to attach the logger and deadline.
var SetPollReady = F.Curry2(func(deadline time.Time, c WithClient) PollContext {
	return PollContext{
		WithClient: c,
		Logger:     slog.New(slog.NewJSONHandler(os.Stdout, nil)),
		Deadline:   deadline,
	}
})

// computeDeadline derives the polling deadline from Params.Timeout.
// Pure function — safe for use with IOE.Let.
var computeDeadline = func(c WithClient) time.Time {
	return time.Now().Add(c.Timeout)
}

// isTerminal returns true for any JobState that ends the polling loop.
var isTerminal = func(s JobState) bool {
	return s == JobComplete || s == JobFailed || s == JobStuck
}

// isComplete returns true only for a successfully completed job.
var isComplete = func(s JobState) bool {
	return s == JobComplete
}
