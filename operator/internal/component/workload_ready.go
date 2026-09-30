package component

import (
	"context"
	"sort"
	"time"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// A component that serves again after one of its containers was killed for
// memory is Ready -- it is serving -- but it is not Healthy. The kubelet keeps
// that container's last termination until its next one, so the kill is a fact
// the cluster still states; until a raised limit rolls the pods, the person
// is owed the kill and the field to raise, not "healthy". Without this answer
// a kill that recovered in a minute is invisible an hour later, and the next
// one arrives as an outage nobody was warned about.
//
// This file is the Ready twin of NotReady: every sized component returns its
// Ready answer through Base.Ready, naming the same workloads its not-ready
// branch explains. The ingress edges and the shared sub-operators (Tekton,
// CloudNativePG, the backup plugin) answer Ready directly: no field of the
// platform's spec sizes them, so a kill there has no field to name.
//
// The kill's sentence is the classifier's own (outOfMemory), so it reads in
// the same words whichever state the component is in. The controller already turns a failure reason into a Warning Event and its
// clearing into ComponentRecovered, so the kill and the fix both reach the
// platform's Event stream with nothing more here.

// Ready is the one ready answer every sized component returns: the
// component's own message, or -- when a current pod of any of its workloads
// carries an out-of-memory kill it recovered from -- Ready with reason
// OutOfMemory, the kill, when it happened, and the field to raise. Every read
// is best-effort and only reads: a workload or pods that cannot be read yield
// the component's own message, never a failed reconcile.
func (b *Base) Ready(ctx context.Context, c client.Client, namespace, message string, workloads ...WorkloadRef) Result {
	for _, workload := range workloads {
		ns := namespace
		if workload.Namespace != "" {
			ns = workload.Namespace
		}
		pods, _, _, found := workloadPods(ctx, c, ns, workload)
		if !found {
			continue
		}
		if expl := explainRecoveredKill(pods, workload.SizedBy); expl != nil {
			return Result{Ready: true, Reason: expl.Reason, Object: expl.Object, Message: expl.Message}
		}
	}
	return Result{Ready: true, Message: message}
}

// explainRecoveredKill is the pure reading behind Ready: the newest pod's
// out-of-memory kill, init containers first (the order the not-ready
// classifier reads), or nil when no current pod met one. Only the last
// termination is read -- a Ready pod's containers are running again, and a
// crash that was not a memory kill has no field to raise, so it says nothing
// here.
func explainRecoveredKill(pods []corev1.Pod, sizedBy string) *Explanation {
	sorted := make([]corev1.Pod, len(pods))
	copy(sorted, pods)
	sort.SliceStable(sorted, func(i, j int) bool {
		return sorted[i].CreationTimestamp.After(sorted[j].CreationTimestamp.Time)
	})
	for i := range sorted {
		pod := &sorted[i]
		for _, cs := range append(append([]corev1.ContainerStatus{}, pod.Status.InitContainerStatuses...), pod.Status.ContainerStatuses...) {
			last := cs.LastTerminationState.Terminated
			if last == nil || last.Reason != "OOMKilled" {
				continue
			}
			return outOfMemory(pod, cs, " at "+last.FinishedAt.UTC().Format(time.RFC3339)+" and is serving again", sizedBy)
		}
	}
	return nil
}
