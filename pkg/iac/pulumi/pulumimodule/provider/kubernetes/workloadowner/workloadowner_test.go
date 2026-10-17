package workloadowner

import (
	"errors"
	"strings"
	"testing"

	"github.com/plantonhq/planton/pkg/failure"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	dynamicfake "k8s.io/client-go/dynamic/fake"
)

// The decision is pinned here offline: an absent workload and one this kind
// created pass, anything else is refused before registration with its owner
// named and the three-part remedy, in words that name the workload's kind; a
// namespace the kind creates must be absent or its own, one another component
// owns must exist, and every namespace the kind installs into is checked. The
// live lanes prove the refusals against a real cluster.

const kind = "KubernetesAgentSandboxController"

func deployment(name string, labels map[string]string, managers ...string) *unstructured.Unstructured {
	object := &unstructured.Unstructured{}
	object.SetAPIVersion("apps/v1")
	object.SetKind("Deployment")
	object.SetNamespace("agent-sandbox-system")
	object.SetName(name)
	object.SetLabels(labels)
	fields := make([]metav1.ManagedFieldsEntry, 0, len(managers))
	for _, m := range managers {
		fields = append(fields, metav1.ManagedFieldsEntry{Manager: m})
	}
	object.SetManagedFields(fields)
	return object
}

func namespaceObject(labels map[string]string, managers ...string) *unstructured.Unstructured {
	object := &unstructured.Unstructured{}
	object.SetAPIVersion("v1")
	object.SetKind("Namespace")
	object.SetName("agent-sandbox-system")
	object.SetLabels(labels)
	fields := make([]metav1.ManagedFieldsEntry, 0, len(managers))
	for _, m := range managers {
		fields = append(fields, metav1.ManagedFieldsEntry{Manager: m})
	}
	object.SetManagedFields(fields)
	return object
}

func cluster(objects ...runtime.Object) *dynamicfake.FakeDynamicClient {
	return dynamicfake.NewSimpleDynamicClient(runtime.NewScheme(), objects...)
}

var (
	controller     = Workload{Kind: Deployment, Namespace: "agent-sandbox-system", Name: "agent-sandbox-controller"}
	created        = []Namespace{{Name: "agent-sandbox-system", Create: true}}
	ownedByAnother = []Namespace{{Name: "agent-sandbox-system", Create: false}}
)

func TestAbsentDeploymentPasses(t *testing.T) {
	if err := refuse(cluster(), kind, created, []Workload{controller}); err != nil {
		t.Fatalf("a first install finds nothing to refuse: %v", err)
	}
}

func TestDeploymentThisKindCreatedPasses(t *testing.T) {
	ours := deployment("agent-sandbox-controller", map[string]string{failure.WorkloadOwnerLabel: kind}, "pulumi-kubernetes")
	if err := refuse(cluster(ours), kind, created, []Workload{controller}); err != nil {
		t.Fatalf("a re-run over this kind's own Deployment must pass: %v", err)
	}
}

func TestDeploymentAnotherInstallationCreatedIsRefused(t *testing.T) {
	foreign := deployment("agent-sandbox-controller", map[string]string{"app": "agent-sandbox-controller"}, "kubectl-client-side-apply")
	err := refuse(cluster(foreign), kind, created, []Workload{controller})
	var f *failure.Failure
	if !errors.As(err, &f) {
		t.Fatalf("expected the explained refusal, got %v", err)
	}
	for part, want := range map[string]string{
		f.Observed: "Deployment agent-sandbox-system/agent-sandbox-controller already runs on the cluster and was not created by this KubernetesAgentSandboxController (last written by field manager(s) kubectl-client-side-apply)",
		f.Meaning:  "another installation of the same software runs here",
		f.NextStep: "kubectl -n agent-sandbox-system label deployment agent-sandbox-controller planton.ai/kind=KubernetesAgentSandboxController",
	} {
		if !strings.Contains(part, want) {
			t.Errorf("refusal lacks %q:\n%s", want, err)
		}
	}
}

func TestAnotherKindsLabelIsNotThisKinds(t *testing.T) {
	otherKind := deployment("agent-sandbox-controller", map[string]string{failure.WorkloadOwnerLabel: "KubernetesHelmRelease"})
	if err := refuse(cluster(otherKind), kind, created, []Workload{controller}); err == nil {
		t.Fatal("a Deployment another catalog kind created is not this kind's to take over")
	}
}

func TestUnmarkedDeploymentSaysSo(t *testing.T) {
	err := refuse(cluster(deployment("agent-sandbox-controller", nil)), kind, created, []Workload{controller})
	if err == nil || !strings.Contains(err.Error(), "it carries no ownership marks") {
		t.Fatalf("an unmarked Deployment must be named as such, got %v", err)
	}
}

func TestNamespaceAnotherComponentOwnsMustExist(t *testing.T) {
	err := refuse(cluster(), kind, ownedByAnother, []Workload{controller})
	var f *failure.Failure
	if !errors.As(err, &f) {
		t.Fatalf("expected the explained refusal, got %v", err)
	}
	for part, want := range map[string]string{
		f.Observed: "namespace agent-sandbox-system does not exist on the cluster, and this KubernetesAgentSandboxController does not create it (spec.create_namespace is false)",
		f.NextStep: "deploy the KubernetesNamespace named agent-sandbox-system first",
	} {
		if !strings.Contains(part, want) {
			t.Errorf("refusal lacks %q:\n%s", want, err)
		}
	}
	present := namespaceObject(map[string]string{"planton.ai/kind": "KubernetesNamespace"})
	if err := refuse(cluster(present), kind, ownedByAnother, []Workload{controller}); err != nil {
		t.Fatalf("a namespace another component created is exactly what create_namespace false expects: %v", err)
	}
}

func TestNamespaceThisKindCreatesMustBeAbsentOrItsOwn(t *testing.T) {
	ours := namespaceObject(map[string]string{failure.WorkloadOwnerLabel: kind}, "pulumi-kubernetes")
	if err := refuse(cluster(ours), kind, created, []Workload{controller}); err != nil {
		t.Fatalf("a re-run over this kind's own namespace must pass: %v", err)
	}
	foreign := namespaceObject(map[string]string{failure.WorkloadOwnerLabel: "KubernetesNamespace"}, "pulumi-kubernetes")
	err := refuse(cluster(foreign), kind, created, []Workload{controller})
	var f *failure.Failure
	if !errors.As(err, &f) {
		t.Fatalf("expected the explained refusal, got %v", err)
	}
	for part, want := range map[string]string{
		f.Observed: "namespace agent-sandbox-system already exists on the cluster and was not created by this KubernetesAgentSandboxController (last written by field manager(s) pulumi-kubernetes)",
		f.Meaning:  "destroying it later would delete the namespace and everything other components put in it",
		f.NextStep: "set spec.create_namespace to false",
	} {
		if !strings.Contains(part, want) {
			t.Errorf("refusal lacks %q:\n%s", want, err)
		}
	}
}

func TestForeignDeploymentIsNamedBeforeTheNamespace(t *testing.T) {
	foreignNamespace := namespaceObject(nil)
	foreignController := deployment("agent-sandbox-controller", nil, "addon-manager")
	err := refuse(cluster(foreignNamespace, foreignController), kind, created, []Workload{controller})
	if err == nil || !strings.Contains(err.Error(), "Deployment agent-sandbox-system/agent-sandbox-controller") {
		t.Fatalf("the other installation's controller names its owner best, so it is refused first; got %v", err)
	}
}

func TestTurningCreateNamespaceOffOnALiveInstallationIsRefused(t *testing.T) {
	ours := namespaceObject(map[string]string{failure.WorkloadOwnerLabel: kind}, "pulumi-kubernetes")
	err := refuse(cluster(ours), kind, ownedByAnother, []Workload{controller})
	var f *failure.Failure
	if !errors.As(err, &f) {
		t.Fatalf("expected the explained refusal, got %v", err)
	}
	for part, want := range map[string]string{
		f.Observed: "namespace agent-sandbox-system was created by this KubernetesAgentSandboxController, and spec.create_namespace is now false",
		f.Meaning:  "would delete the namespace",
		f.NextStep: "destroy this resource first",
	} {
		if !strings.Contains(part, want) {
			t.Errorf("refusal lacks %q:\n%s", want, err)
		}
	}
}

func daemonSet(namespace, name string, labels map[string]string, managers ...string) *unstructured.Unstructured {
	object := deployment(name, labels, managers...)
	object.SetKind("DaemonSet")
	object.SetNamespace(namespace)
	return object
}

func TestForeignDaemonSetIsRefusedInItsOwnWords(t *testing.T) {
	nodeAgent := Workload{Kind: DaemonSet, Namespace: "ate-system", Name: "atelet"}
	foreign := daemonSet("ate-system", "atelet", nil, "kubectl-client-side-apply")
	err := refuse(cluster(foreign), kind, created, []Workload{nodeAgent})
	var f *failure.Failure
	if !errors.As(err, &f) {
		t.Fatalf("expected the explained refusal, got %v", err)
	}
	for part, want := range map[string]string{
		f.Observed: "DaemonSet ate-system/atelet already runs on the cluster and was not created by this KubernetesAgentSandboxController (last written by field manager(s) kubectl-client-side-apply)",
		f.NextStep: "kubectl -n ate-system label daemonset atelet planton.ai/kind=KubernetesAgentSandboxController",
	} {
		if !strings.Contains(part, want) {
			t.Errorf("refusal lacks %q:\n%s", want, err)
		}
	}
	ours := daemonSet("ate-system", "atelet", map[string]string{failure.WorkloadOwnerLabel: kind})
	if err := refuse(cluster(ours), kind, created, []Workload{nodeAgent}); err != nil {
		t.Fatalf("this kind's own DaemonSet must pass: %v", err)
	}
}

func TestEveryNamespaceIsChecked(t *testing.T) {
	both := []Namespace{{Name: "agent-sandbox-system", Create: true}, {Name: "podcertificate-controller-system", Create: true}}
	foreign := namespaceObject(nil, "kubectl-create")
	foreign.SetName("podcertificate-controller-system")
	err := refuse(cluster(foreign), kind, both, nil)
	if err == nil || !strings.Contains(err.Error(), "namespace podcertificate-controller-system already exists") {
		t.Fatalf("the second namespace another installation holds must be refused, got %v", err)
	}
}

func TestUnknownWorkloadKindIsAnError(t *testing.T) {
	if err := refuse(cluster(), kind, created, []Workload{{Kind: "ReplicaSet", Namespace: "ns", Name: "x"}}); err == nil {
		t.Fatal("a kind the check cannot read must not pass silently")
	}
}
