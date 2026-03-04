package wait

import (
	A "github.com/IBM/fp-go/v2/array"
	F "github.com/IBM/fp-go/v2/function"
	O "github.com/IBM/fp-go/v2/option"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
)

// stuckReasons is the set of container waiting reasons that indicate a stuck pod.
var stuckReasons = map[string]bool{
	"ImagePullBackOff":           true,
	"ErrImagePull":               true,
	"InvalidImageName":           true,
	"CrashLoopBackOff":           true,
	"CreateContainerConfigError": true,
	"CreateContainerError":       true,
	"ContainerCannotRun":         true,
}

// isStuckReason returns true if the waiting reason signals a stuck container.
var isStuckReason = func(reason string) bool {
	return stuckReasons[reason]
}

// allContainerStatuses concatenates init and regular container statuses for a pod.
var allContainerStatuses = func(pod corev1.Pod) []corev1.ContainerStatus {
	return append(pod.Status.InitContainerStatuses, pod.Status.ContainerStatuses...)
}

// toStuckReason extracts a stuck reason from a ContainerStatus if one exists.
// Uses O.FromNillable to safely handle the nullable Waiting pointer.
var toStuckReason = func(cs corev1.ContainerStatus) O.Option[string] {
	return F.Pipe3(
		cs.State.Waiting,
		O.FromNillable,
		O.Map(func(w *corev1.ContainerStateWaiting) string { return w.Reason }),
		O.Chain(O.FromPredicate(isStuckReason)),
	)
}

// FindStuckReason scans all pods and all containers for a known stuck reason.
// Returns Some(reason) on first match, None if all containers are healthy.
var FindStuckReason = func(pods *corev1.PodList) O.Option[string] {
	return F.Pipe3(
		pods.Items,
		A.Chain(allContainerStatuses),
		A.FilterMap(toStuckReason),
		A.Head,
	)
}

// ClassifyPodState returns JobStuck if any container is stuck, otherwise JobPending.
var ClassifyPodState = func(pods *corev1.PodList) JobState {
	return F.Pipe1(
		FindStuckReason(pods),
		O.Fold(
			F.Constant(JobPending),
			func(_ string) JobState { return JobStuck },
		),
	)
}

// conditionTypeMap maps Kubernetes JobConditionTypes to domain JobStates.
var conditionTypeMap = map[batchv1.JobConditionType]JobState{
	batchv1.JobComplete: JobComplete,
	batchv1.JobFailed:   JobFailed,
}

// isTerminalCondition returns true for a Complete or Failed condition with Status=True.
var isTerminalCondition = func(c batchv1.JobCondition) bool {
	return c.Status == corev1.ConditionTrue &&
		(c.Type == batchv1.JobComplete || c.Type == batchv1.JobFailed)
}

// conditionToJobState maps a terminal JobCondition to the corresponding JobState.
var conditionToJobState = func(c batchv1.JobCondition) JobState {
	return conditionTypeMap[c.Type]
}

// toTerminalState converts a single JobCondition into an Option[JobState].
// Returns None for non-terminal or Status!=True conditions.
var toTerminalState = func(c batchv1.JobCondition) O.Option[JobState] {
	return F.Pipe2(
		c,
		O.FromPredicate(isTerminalCondition),
		O.Map(conditionToJobState),
	)
}

// ExtractJobCondition returns the first terminal condition (Complete or Failed) from a Job.
var ExtractJobCondition = func(job *batchv1.Job) O.Option[JobState] {
	return F.Pipe2(
		job.Status.Conditions,
		A.FilterMap(toTerminalState),
		A.Head,
	)
}
