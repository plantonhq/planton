package component

import (
	"context"
	"testing"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	v1 "github.com/plantonhq/planton/operator/api/v1"
)

// The Ready answer is pinned here: a component serving again after an
// out-of-memory kill reads Ready with the kill and the field to raise, and a
// component whose pods never met one reads its own message. A recovered kill
// that reads "healthy" is the defect these tests exist to catch -- the next
// kill then arrives as an outage nobody was warned about.

var killedAt = time.Date(2026, 9, 30, 14, 2, 5, 0, time.UTC)

// servingAfterKill is a Ready pod whose container was killed for memory and
// is running again on its next attempt, the state the kubelet leaves it in.
func servingAfterKill(name string, labels map[string]string, limit string) corev1.Pod {
	pod := runningPod(name, time.Minute, 1)
	pod.Labels = labels
	pod.Status.Conditions = []corev1.PodCondition{{Type: corev1.PodReady, Status: corev1.ConditionTrue}}
	pod.Spec.Containers = []corev1.Container{{Name: "main", Resources: corev1.ResourceRequirements{
		Limits: corev1.ResourceList{corev1.ResourceMemory: resource.MustParse(limit)}}}}
	pod.Status.ContainerStatuses[0].LastTerminationState = corev1.ContainerState{
		Terminated: &corev1.ContainerStateTerminated{Reason: "OOMKilled", ExitCode: 137, FinishedAt: metav1.NewTime(killedAt)}}
	return pod
}

func TestExplainRecoveredKill_NamesTheKillTheTimeAndTheField(t *testing.T) {
	pod := servingAfterKill("planton-control-plane-5f7c-x2", nil, "6Gi")
	expl := explainRecoveredKill([]corev1.Pod{pod}, "spec.controlPlane.resources")
	if expl == nil || expl.Reason != v1.ComponentReasonOutOfMemory {
		t.Fatalf("expected OutOfMemory, got %+v", expl)
	}
	if expl.Object == nil || expl.Object.Kind != "Pod" || expl.Object.Name != pod.Name {
		t.Errorf("the pod is the object, got %+v", expl.Object)
	}
	mustContain(t, expl.Message,
		`container "main" of pod planton-control-plane-5f7c-x2`,
		"memory limit of 6Gi at 2026-09-30T14:02:05Z and is serving again",
		"1 restarts",
		"raise spec.controlPlane.resources.limits.memory", "keep their defaults")
}

func TestExplainRecoveredKill_SaysNothingWithoutAKill(t *testing.T) {
	clean := runningPod("planton-console-1", time.Hour, 0)
	crashed := runningPod("planton-console-2", time.Minute, 1)
	crashed.Status.ContainerStatuses[0].LastTerminationState = corev1.ContainerState{
		Terminated: &corev1.ContainerStateTerminated{Reason: "Error", ExitCode: 1}}
	for name, pods := range map[string][]corev1.Pod{
		"no pods":                       nil,
		"a pod that never restarted":    {clean},
		"a restart that was not a kill": {crashed},
	} {
		if expl := explainRecoveredKill(pods, "spec.console.resources"); expl != nil {
			t.Errorf("%s: a Ready component with no memory kill reads its own message, got %+v", name, expl)
		}
	}
}

// An init container killed for memory blocked everything after it, the same
// order the not-ready classifier reads.
func TestExplainRecoveredKill_ReadsInitContainers(t *testing.T) {
	pod := runningPod("planton-identity-1", time.Minute, 0)
	pod.Status.InitContainerStatuses = []corev1.ContainerStatus{{
		Name: "realm-import", RestartCount: 1,
		LastTerminationState: corev1.ContainerState{Terminated: &corev1.ContainerStateTerminated{Reason: "OOMKilled", FinishedAt: metav1.NewTime(killedAt)}},
	}}
	expl := explainRecoveredKill([]corev1.Pod{pod}, "spec.identity.resources")
	if expl == nil {
		t.Fatal("an init container's kill must be read")
	}
	mustContain(t, expl.Message, `container "realm-import"`, "no limit set: the node itself ran out of memory")
}

func TestReady(t *testing.T) {
	ctx := context.Background()
	const namespace = "planton"
	selector := map[string]string{"app.kubernetes.io/name": "control-plane"}
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Name: "planton-control-plane", Namespace: namespace},
		Spec:       appsv1.DeploymentSpec{Selector: &metav1.LabelSelector{MatchLabels: selector}},
	}
	withPods := func(pods ...corev1.Pod) client.Client {
		objs := []client.Object{deployment.DeepCopy()}
		for i := range pods {
			pods[i].Namespace = namespace
			objs = append(objs, &pods[i])
		}
		return fake.NewClientBuilder().WithScheme(bindingScheme(t)).WithObjects(objs...).Build()
	}
	controlPlane := DeploymentRef("planton-control-plane").Sized("spec.controlPlane.resources")
	var b Base

	t.Run("a kill the pod recovered from reads Ready with the kill", func(t *testing.T) {
		c := withPods(servingAfterKill("planton-control-plane-1", selector, "6Gi"))
		got := b.Ready(ctx, c, namespace, "ControlPlane healthy", controlPlane)
		if !got.Ready || got.Reason != v1.ComponentReasonOutOfMemory {
			t.Fatalf("expected Ready with OutOfMemory, got %+v", got)
		}
		mustContain(t, got.Message, "is serving again", "raise spec.controlPlane.resources.limits.memory")
	})

	t.Run("pods that never met a kill read the component's own message", func(t *testing.T) {
		c := withPods(runningPod("planton-control-plane-1", time.Hour, 0))
		got := b.Ready(ctx, c, namespace, "ControlPlane healthy", controlPlane)
		if got != (Result{Ready: true, Message: "ControlPlane healthy"}) {
			t.Errorf("expected the plain Ready answer, got %+v", got)
		}
	})

	t.Run("another workload's pods are never read", func(t *testing.T) {
		stranger := servingAfterKill("someone-else-1", map[string]string{"app.kubernetes.io/name": "not-planton"}, "1Gi")
		got := b.Ready(ctx, withPods(stranger), namespace, "ControlPlane healthy", controlPlane)
		if got.Reason != "" {
			t.Errorf("a pod outside the workload's selector says nothing about it, got %+v", got)
		}
	})

	t.Run("a workload that cannot be read yields the plain answer, never an error", func(t *testing.T) {
		c := fake.NewClientBuilder().WithScheme(bindingScheme(t)).Build()
		got := b.Ready(ctx, c, namespace, "ControlPlane healthy", controlPlane)
		if got != (Result{Ready: true, Message: "ControlPlane healthy"}) {
			t.Errorf("expected the plain Ready answer, got %+v", got)
		}
	})

	// Temporal's four services each run their own Deployment; a kill in
	// history, the busiest, is the one this component would otherwise hide
	// behind a healthy frontend.
	t.Run("every workload of a multi-workload component is read", func(t *testing.T) {
		historySelector := map[string]string{"app.kubernetes.io/component": "history"}
		history := &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{Name: "planton-temporal-history", Namespace: namespace},
			Spec:       appsv1.DeploymentSpec{Selector: &metav1.LabelSelector{MatchLabels: historySelector}},
		}
		killed := servingAfterKill("planton-temporal-history-1", historySelector, "1Gi")
		killed.Namespace = namespace
		c := fake.NewClientBuilder().WithScheme(bindingScheme(t)).WithObjects(deployment.DeepCopy(), history, &killed).Build()
		got := b.Ready(ctx, c, namespace, "Temporal healthy",
			controlPlane,
			DeploymentRef("planton-temporal-history").Sized("spec.temporal.history.resources"))
		if got.Reason != v1.ComponentReasonOutOfMemory {
			t.Fatalf("expected the history kill, got %+v", got)
		}
		mustContain(t, got.Message, "raise spec.temporal.history.resources.limits.memory")
	})

	// A CloudNativePG cluster has no Deployment selector; its pods carry the
	// cluster label, the same way the not-ready classifier finds them.
	t.Run("a PostgreSQL cluster's pods are found by their cluster label", func(t *testing.T) {
		killed := servingAfterKill("planton-postgres-1", map[string]string{cnpgClusterLabel: "planton-postgres"}, "2Gi")
		killed.Namespace = namespace
		c := fake.NewClientBuilder().WithScheme(bindingScheme(t)).WithObjects(&killed).Build()
		got := b.Ready(ctx, c, namespace, "PostgreSQL healthy", PostgresClusterRef("planton-postgres").Sized("spec.database.postgresql.resources"))
		if got.Reason != v1.ComponentReasonOutOfMemory {
			t.Fatalf("expected the database's kill, got %+v", got)
		}
	})
}
