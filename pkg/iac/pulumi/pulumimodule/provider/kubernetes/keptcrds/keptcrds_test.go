package keptcrds

import (
	"errors"
	"strings"
	"testing"

	"github.com/plantonhq/planton/pkg/kubernetes/helmcrds"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	dynamicfake "k8s.io/client-go/dynamic/fake"
	clienttesting "k8s.io/client-go/testing"
)

// The resource registration is proven by the e2e lanes; what can be proven
// offline is everything the primitive decides before registering: how the
// cluster's answers map onto the ownership and version checks, and how the
// permission probe turns a denial into the explained refusal.

func crdObject(name string, labels, annotations map[string]string, managers ...string) *unstructured.Unstructured {
	object := &unstructured.Unstructured{}
	object.SetAPIVersion("apiextensions.k8s.io/v1")
	object.SetKind("CustomResourceDefinition")
	object.SetName(name)
	object.SetLabels(labels)
	object.SetAnnotations(annotations)
	fields := make([]metav1.ManagedFieldsEntry, 0, len(managers))
	for _, m := range managers {
		fields = append(fields, metav1.ManagedFieldsEntry{Manager: m, Operation: metav1.ManagedFieldsOperationApply})
	}
	object.SetManagedFields(fields)
	return object
}

func fakeCluster(objects ...runtime.Object) (*clusterReads, *dynamicfake.FakeDynamicClient) {
	scheme := runtime.NewScheme()
	client := dynamicfake.NewSimpleDynamicClientWithCustomListKinds(scheme, map[schema.GroupVersionResource]string{
		crdGVR: "CustomResourceDefinitionList",
	}, objects...)
	return &clusterReads{dynamic: client}, client
}

func TestToExistingReadsStampsAndOwnershipMarks(t *testing.T) {
	ours := toExisting(*crdObject("widgets.example",
		map[string]string{helmcrds.LabelSource: "widgets"},
		map[string]string{helmcrds.AnnotationSourceVersion: "0.120.0"},
		"kubectl", "kube-apiserver"))
	if ours.SourceLabel != "widgets" || ours.Version != "0.120.0" || len(ours.Managers) != 2 {
		t.Fatalf("stamped CRD mapped wrong: %+v", ours)
	}
	helmOwned := toExisting(*crdObject("legacy.example", nil,
		map[string]string{"meta.helm.sh/release-name": "by-hand", "meta.helm.sh/release-namespace": "ops"}, "helm"))
	if helmOwned.HelmReleaseName != "by-hand" || helmOwned.HelmReleaseNamespace != "ops" || helmOwned.Version != "" {
		t.Fatalf("Helm-owned CRD mapped wrong: %+v", helmOwned)
	}
}

func TestReadExistingSkipsAbsentAndReturnsPresent(t *testing.T) {
	cluster, _ := fakeCluster(crdObject("present.example",
		map[string]string{helmcrds.LabelSource: "demo"},
		map[string]string{helmcrds.AnnotationSourceVersion: "1.0.0"}))
	existing, err := cluster.readExisting([]helmcrds.CRD{{Name: "present.example"}, {Name: "absent.example"}})
	if err != nil {
		t.Fatal(err)
	}
	if len(existing) != 1 || existing[0].Name != "present.example" || existing[0].Version != "1.0.0" {
		t.Fatalf("expected only the present CRD with its stamp, got %+v", existing)
	}
}

// The order of the checks matters: a CRD another source stamped at a higher
// version is refused as someone else's, never as a downgrade.
func TestOwnershipIsRefusedBeforeVersion(t *testing.T) {
	src := helmcrds.Source{Repository: "https://charts.example.com", Chart: "demo", Version: "1.0.0"}
	existing := []helmcrds.ExistingCRD{helmcrds.ExistingFromObject("a.example",
		map[string]string{helmcrds.LabelSource: "other"},
		map[string]string{helmcrds.AnnotationSourceVersion: "9.0.0"}, nil)}
	err := helmcrds.CheckOwnership(existing, src)
	var f *helmcrds.Failure
	if !errors.As(err, &f) || !strings.Contains(f.Observed, "another Planton module") {
		t.Fatalf("expected the ownership refusal, got %v", err)
	}
}

// A downgrade is refused unless the kind's own upgrade rule judges version
// moves (Agent Substrate rolls back one minor, restoring the older
// definitions as upstream's rollback does); ownership is checked either way.
func TestDowngradeIsRefusedUnlessTheKindJudgesIt(t *testing.T) {
	src := helmcrds.Source{BundleURLs: []string{"https://example.invalid/{{version}}/crds.yaml"}, BundleName: "agent-substrate", Version: "v1.0.0"}
	stamped := func(source string) []helmcrds.ExistingCRD {
		return []helmcrds.ExistingCRD{helmcrds.ExistingFromObject("workerpools.ate.dev",
			map[string]string{helmcrds.LabelSource: source},
			map[string]string{helmcrds.AnnotationSourceVersion: "v1.1.0"}, nil)}
	}
	var f *helmcrds.Failure
	if err := refuseExisting(stamped("agent-substrate"), nil, Args{Source: src}); !errors.As(err, &f) || !strings.Contains(f.Meaning, "older CRD schema") {
		t.Fatalf("a downgrade must be refused by default, got %v", err)
	}
	judged := Args{Source: src, Policy: helmcrds.Policy{DowngradeJudgedByKind: true}}
	if err := refuseExisting(stamped("agent-substrate"), nil, judged); err != nil {
		t.Fatalf("a kind that judges its own version moves must be let through, got %v", err)
	}
	if err := refuseExisting(stamped("other"), nil, judged); !errors.As(err, &f) || !strings.Contains(f.Observed, "another Planton module") {
		t.Fatalf("ownership must still be checked, got %v", err)
	}
}

func answerReview(client *dynamicfake.FakeDynamicClient, allowedVerbs map[string]bool, username string) {
	client.PrependReactor("create", "selfsubjectaccessreviews", func(action clienttesting.Action) (bool, runtime.Object, error) {
		review := action.(clienttesting.CreateAction).GetObject().(*unstructured.Unstructured)
		verb, _, _ := unstructured.NestedString(review.Object, "spec", "resourceAttributes", "verb")
		answer := review.DeepCopy()
		_ = unstructured.SetNestedField(answer.Object, allowedVerbs[verb], "status", "allowed")
		if !allowedVerbs[verb] {
			_ = unstructured.SetNestedField(answer.Object, "RBAC: no rule grants "+verb, "status", "reason")
		}
		return true, answer, nil
	})
	client.PrependReactor("create", "selfsubjectreviews", func(action clienttesting.Action) (bool, runtime.Object, error) {
		answer := action.(clienttesting.CreateAction).GetObject().(*unstructured.Unstructured).DeepCopy()
		_ = unstructured.SetNestedField(answer.Object, username, "status", "userInfo", "username")
		return true, answer, nil
	})
}

func TestRefuseIfDeniedNamesIdentityVerbAndRemedy(t *testing.T) {
	cluster, client := fakeCluster()
	answerReview(client, map[string]bool{"get": true, "create": true, "patch": false}, "system:serviceaccount:ci:deployer")
	err := cluster.refuseIfDenied(Args{KeepOnUninstall: true})
	var f *helmcrds.Failure
	if !errors.As(err, &f) {
		t.Fatalf("expected the explained refusal, got %v", err)
	}
	for _, want := range []string{"system:serviceaccount:ci:deployer", "may not patch customresourcedefinitions.apiextensions.k8s.io at the cluster scope", "RBAC: no rule grants patch"} {
		if !strings.Contains(f.Observed, want) {
			t.Errorf("observation lacks %q: %s", want, f.Observed)
		}
	}
	if !strings.Contains(f.NextStep, "iac/permissions.yaml") || !strings.Contains(f.NextStep, "spec.crds.install") {
		t.Errorf("next step must name the permissions file and the namespaced way out: %s", f.NextStep)
	}
}

func TestRefuseIfDeniedPassesWhenEveryVerbIsAllowed(t *testing.T) {
	cluster, client := fakeCluster()
	answerReview(client, map[string]bool{"get": true, "create": true, "patch": true, "delete": true}, "admin")
	if err := cluster.refuseIfDenied(Args{KeepOnUninstall: false}); err != nil {
		t.Fatal(err)
	}
}

// Delete is only asked for when a destroy is meant to take the CRDs with it;
// the default keep posture never needs it.
func TestRefuseIfDeniedAsksDeleteOnlyWhenNotKeeping(t *testing.T) {
	cluster, client := fakeCluster()
	answerReview(client, map[string]bool{"get": true, "create": true, "patch": true, "delete": false}, "admin")
	if err := cluster.refuseIfDenied(Args{KeepOnUninstall: true}); err != nil {
		t.Fatalf("keep posture must not ask for delete: %v", err)
	}
	if err := cluster.refuseIfDenied(Args{KeepOnUninstall: false}); err == nil {
		t.Fatal("a destroy that deletes CRDs needs delete")
	}
}

func TestApplyWithoutInstallRegistersNothing(t *testing.T) {
	resources, err := Apply(nil, Args{Install: false})
	if err != nil || resources != nil {
		t.Fatalf("install=false must be a no-op, got %v %v", resources, err)
	}
}
