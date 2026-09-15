package component

import (
	"context"
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

// The gate's decisions are what these tests pin -- install vs resume vs
// respect-foreign vs wait. The loader stub returns no objects (invocation is
// the observable), and the real apply path is exercised by envtest and the
// live lab.

const (
	testCRDName    = "widgets.example.io"
	testNamespace  = "sub-op-system"
	testDeployment = "sub-op-controller"
)

func subOperatorScheme(t *testing.T) *runtime.Scheme {
	t.Helper()
	s := runtime.NewScheme()
	if err := clientgoscheme.AddToScheme(s); err != nil {
		t.Fatalf("building scheme: %v", err)
	}
	crdGVK := schema.GroupVersionKind{
		Group: "apiextensions.k8s.io", Version: "v1", Kind: "CustomResourceDefinition",
	}
	s.AddKnownTypeWithName(crdGVK, &unstructured.Unstructured{})
	s.AddKnownTypeWithName(crdGVK.GroupVersion().WithKind("CustomResourceDefinitionList"),
		&unstructured.UnstructuredList{})
	return s
}

func testCRD(managedBy string) *unstructured.Unstructured {
	crd := &unstructured.Unstructured{}
	crd.SetGroupVersionKind(schema.GroupVersionKind{
		Group: "apiextensions.k8s.io", Version: "v1", Kind: "CustomResourceDefinition",
	})
	crd.SetName(testCRDName)
	if managedBy != "" {
		crd.SetLabels(map[string]string{"app.kubernetes.io/managed-by": managedBy})
	}
	return crd
}

func testControllerDeployment(available bool) *appsv1.Deployment {
	deploy := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Name: testDeployment, Namespace: testNamespace},
	}
	if available {
		// A completed rollout, as the readiness test defines it: the spec
		// observed, one replica on the current template, none stale, available.
		deploy.Status.ObservedGeneration = deploy.Generation
		deploy.Status.Replicas = 1
		deploy.Status.UpdatedReplicas = 1
		deploy.Status.AvailableReplicas = 1
	}
	return deploy
}

// runGate builds a fake cluster from the given objects and runs the gate,
// returning its verdict plus how many times the manifest loader was invoked.
func runGate(t *testing.T, skip bool, objs ...client.Object) (ready bool, applies int) {
	t.Helper()
	return runGateWith(t, skip, false, objs...)
}

func runGateWith(t *testing.T, skip, reapplyWhileNotReady bool, objs ...client.Object) (ready bool, applies int) {
	t.Helper()
	c := fake.NewClientBuilder().WithScheme(subOperatorScheme(t)).WithObjects(objs...).Build()

	base := &Base{}
	ready, err := base.EnsureSubOperator(context.Background(), c, SubOperatorOptions{
		LogName:       "test-sub-operator",
		SkipRequested: skip,
		CRDName:       testCRDName,
		Loader: func() ([]*unstructured.Unstructured, error) {
			applies++
			return nil, nil
		},
		Namespace:            testNamespace,
		Deployments:          []string{testDeployment},
		ReapplyWhileNotReady: reapplyWhileNotReady,
	})
	if err != nil {
		t.Fatalf("EnsureSubOperator: %v", err)
	}
	return ready, applies
}

// A release the operator owns alone may ask to be re-applied on every pass
// its controller stands unready: an apply refused on its last object leaves
// the Deployment starving for something only a re-apply lands. A serving
// controller is still never re-applied.
func TestEnsureSubOperator_ReapplyWhileNotReadyHealsAPartialInstall(t *testing.T) {
	ready, applies := runGateWith(t, false, true, testCRD(SSAFieldManager), testControllerDeployment(false))
	if ready {
		t.Error("an unavailable controller cannot be ready")
	}
	if applies != 1 {
		t.Errorf("an unready controller under ReapplyWhileNotReady is re-applied once per pass, applied %d times", applies)
	}
	ready, applies = runGateWith(t, false, true, testCRD(SSAFieldManager), testControllerDeployment(true))
	if !ready || applies != 0 {
		t.Errorf("a serving controller is never re-applied: ready=%v applies=%d", ready, applies)
	}
}

func TestEnsureSubOperator_SkipTrustsTheAdopter(t *testing.T) {
	ready, applies := runGate(t, true)
	if !ready {
		t.Error("skip must report the sub-operator usable -- the adopter promised an install")
	}
	if applies != 0 {
		t.Errorf("skip must not apply manifests, applied %d times", applies)
	}
}

func TestEnsureSubOperator_AbsentCRDInstalls(t *testing.T) {
	ready, applies := runGate(t, false)
	if ready {
		t.Error("a fresh install cannot be ready in the same pass")
	}
	if applies != 1 {
		t.Errorf("expected exactly one manifest apply, got %d", applies)
	}
}

// A CRD applied by anything other than this operator (helm, GitOps) means a
// foreign install whose controller Deployment name and namespace are unknown:
// respect it exactly like an explicit skip instead of waiting forever on a
// Deployment that will never appear under our names. Both foreign shapes are
// pinned -- another tool's managed-by label AND no label at all.
func TestEnsureSubOperator_ForeignInstallRespected(t *testing.T) {
	for _, managedBy := range []string{"Helm", ""} {
		ready, applies := runGate(t, false, testCRD(managedBy))
		if !ready {
			t.Errorf("managed-by %q: a foreign install must be respected as pre-installed", managedBy)
		}
		if applies != 0 {
			t.Errorf("managed-by %q: a foreign install must never be re-applied, applied %d times", managedBy, applies)
		}
	}
}

// The stranding defect this gate exists to kill: applies land CRDs before
// Deployments, so an apply that dies partway leaves the CRD present and the
// controller missing -- a CRD-presence-only gate then says "installed"
// forever while readiness waits on a controller that never deployed.
func TestEnsureSubOperator_PartialInstallResumes(t *testing.T) {
	ready, applies := runGate(t, false, testCRD(SSAFieldManager))
	if ready {
		t.Error("a partial install cannot be ready")
	}
	if applies != 1 {
		t.Errorf("expected the resume apply, got %d applies", applies)
	}
}

func TestEnsureSubOperator_ExistingUnreadyWaitsWithoutReapply(t *testing.T) {
	ready, applies := runGate(t, false, testCRD(SSAFieldManager), testControllerDeployment(false))
	if ready {
		t.Error("an unavailable controller cannot be ready")
	}
	if applies != 0 {
		t.Errorf("a complete install must never be re-applied, applied %d times", applies)
	}
}

func TestEnsureSubOperator_CompleteInstallReady(t *testing.T) {
	ready, applies := runGate(t, false, testCRD(SSAFieldManager), testControllerDeployment(true))
	if !ready {
		t.Error("expected ready with our CRD and an available controller")
	}
	if applies != 0 {
		t.Errorf("a complete install must never be re-applied, applied %d times", applies)
	}
}
