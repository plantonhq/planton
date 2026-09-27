package resources

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	v1 "github.com/plantonhq/planton/operator/api/v1"
)

// defaultTemporalOptions and defaultOpenFGAOptions are the charts' values as a
// platform that declares no sizing renders them.
func defaultTemporalOptions(crName, namespace string) TemporalHelmOptions {
	return TemporalHelmOptions{
		CRName:    crName,
		Namespace: namespace,
		Frontend:  Effective(SizingTemporalFrontend, nil),
		History:   Effective(SizingTemporalHistory, nil),
		Matching:  Effective(SizingTemporalMatching, nil),
		Worker:    Effective(SizingTemporalWorker, nil),
	}
}

func defaultOpenFGAOptions(crName, namespace string) OpenFGAHelmOptions {
	return OpenFGAHelmOptions{CRName: crName, Namespace: namespace, Resources: Effective(SizingOpenFGA, nil)}
}

func quantities(pairs ...string) corev1.ResourceList {
	out := corev1.ResourceList{}
	for i := 0; i+1 < len(pairs); i += 2 {
		out[corev1.ResourceName(pairs[i])] = resource.MustParse(pairs[i+1])
	}
	return out
}

func assertQuantity(t *testing.T, list corev1.ResourceList, name corev1.ResourceName, want string) {
	t.Helper()
	got, ok := list[name]
	if want == "" {
		if ok {
			t.Errorf("%s: expected unset, got %s", name, got.String())
		}
		return
	}
	if !ok || got.Cmp(resource.MustParse(want)) != 0 {
		t.Errorf("%s: expected %s, got %v", name, want, got.String())
	}
}

// Every registered default is in the house pattern, and every path names a
// component -- the one rule the registry holds itself to.
func TestComponentSizing_EveryDefaultIsTheHousePattern(t *testing.T) {
	for _, path := range SizingPaths() {
		s := ComponentSizing[path]
		assertHousePattern(t, path, s.Default)
		if s.Component == "" {
			t.Errorf("%s: names no component, so its sizing could never reach a status", path)
		}
		for name, floor := range s.Floor {
			if req, ok := s.Default.Requests[name]; ok && req.Cmp(floor) < 0 {
				t.Errorf("%s: the default %s request %s is below its own floor %s", path, name, req.String(), floor.String())
			}
		}
	}
}

func TestEffective_NoOverrideIsTheDefault(t *testing.T) {
	got := Effective(SizingControlPlane, nil)
	assertQuantity(t, got.Requests, corev1.ResourceCPU, "250m")
	assertQuantity(t, got.Requests, corev1.ResourceMemory, "1Gi")
	assertQuantity(t, got.Limits, corev1.ResourceMemory, "4Gi")
	assertQuantity(t, got.Limits, corev1.ResourceCPU, "")
}

// The red case this rule exists for: an override that sets one quantity kept
// nothing else of the default (the store's "replaced whole" rule), so a person
// raising only a memory limit silently lost the CPU and memory requests.
func TestEffective_OneQuantityKeepsEveryOther(t *testing.T) {
	got := Effective(SizingControlPlane, &v1.ComponentResources{Limits: quantities("memory", "6Gi")})
	assertQuantity(t, got.Limits, corev1.ResourceMemory, "6Gi")
	assertQuantity(t, got.Requests, corev1.ResourceCPU, "250m")
	assertQuantity(t, got.Requests, corev1.ResourceMemory, "1Gi")
}

func TestEffective_BothHalvesWin(t *testing.T) {
	got := Effective(SizingRunner, &v1.ComponentResources{
		Requests: quantities("cpu", "500m", "memory", "1Gi"),
		Limits:   quantities("memory", "3Gi"),
	})
	assertQuantity(t, got.Requests, corev1.ResourceCPU, "500m")
	assertQuantity(t, got.Requests, corev1.ResourceMemory, "1Gi")
	assertQuantity(t, got.Limits, corev1.ResourceMemory, "3Gi")
}

// A CPU limit is the adopter's to add (a cluster policy may require one); the
// default never carries one.
func TestEffective_AQuantityTheDefaultLacksIsAdded(t *testing.T) {
	got := Effective(SizingConsole, &v1.ComponentResources{Limits: quantities("cpu", "2", "ephemeral-storage", "2Gi")})
	assertQuantity(t, got.Limits, corev1.ResourceCPU, "2")
	assertQuantity(t, got.Limits, corev1.ResourceEphemeralStorage, "2Gi")
	assertQuantity(t, got.Limits, corev1.ResourceMemory, "2Gi")
}

// The result is the caller's: changing it never changes the registry.
func TestEffective_ReturnsACopy(t *testing.T) {
	got := Effective(SizingGateway, nil)
	got.Requests[corev1.ResourceMemory] = resource.MustParse("1Ti")
	assertQuantity(t, Effective(SizingGateway, nil).Requests, corev1.ResourceMemory, "64Mi")
}

func TestEffectiveFor_ReadsEachComponentsOwnField(t *testing.T) {
	spec := &v1.PlantonPlatformSpec{
		Temporal: &v1.TemporalSpec{History: &v1.TemporalServiceSpec{Resources: &v1.ComponentResources{Limits: quantities("memory", "3Gi")}}},
	}
	assertQuantity(t, EffectiveFor(SizingTemporalHistory, spec).Limits, corev1.ResourceMemory, "3Gi")
	assertQuantity(t, EffectiveFor(SizingTemporalFrontend, spec).Limits, corev1.ResourceMemory, "512Mi")
	assertQuantity(t, EffectiveFor(SizingControlPlane, spec).Limits, corev1.ResourceMemory, "4Gi")
}

func TestEffective_AnUnregisteredPathIsADefect(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("an unregistered path must panic: it is a defect in the operator, never input")
		}
	}()
	Effective("spec.nothing.resources", nil)
}

// A builder handed no sizing fails loudly instead of rendering an unsized pod.
func TestMustBeSized_AnEmptySizeIsADefect(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("an empty size must panic: its component resolved none")
		}
	}()
	ConsoleDeployment(ConsoleConfig{CRName: "planton", Namespace: "default", Version: "v1.0.0", Replicas: 1})
}
