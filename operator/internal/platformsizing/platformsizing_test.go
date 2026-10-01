package platformsizing

import (
	"strings"
	"testing"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	v1 "github.com/plantonhq/planton/operator/api/v1"
)

func list(pairs ...string) corev1.ResourceList {
	out := corev1.ResourceList{}
	for i := 0; i+1 < len(pairs); i += 2 {
		out[corev1.ResourceName(pairs[i])] = resource.MustParse(pairs[i+1])
	}
	return out
}

func mustContain(t *testing.T, got string, wants ...string) {
	t.Helper()
	for _, want := range wants {
		if !strings.Contains(got, want) {
			t.Errorf("expected %q in:\n%s", want, got)
		}
	}
}

// A platform that declares nothing runs every registered default.
func TestCheck_TheDefaultsAreRunnable(t *testing.T) {
	v := Check(&v1.PlantonPlatformSpec{})
	if !v.Runnable || v.Reason != ReasonRunnable {
		t.Fatalf("the operator's own defaults must be runnable, got %+v", v)
	}
}

// The red case this check exists for: a limit lowered alone below the
// default request, which Kubernetes would refuse only at pod admission.
func TestCheck_ALimitBelowTheDefaultRequestNamesTheRequestToSet(t *testing.T) {
	v := Check(&v1.PlantonPlatformSpec{ControlPlane: &v1.ControlPlaneSpec{
		Resources: &v1.ComponentResources{Limits: list("memory", "512Mi")},
	}})
	if v.Runnable || v.Reason != ReasonNotRunnable {
		t.Fatalf("expected a refusal, got %+v", v)
	}
	mustContain(t, v.Message, "spec.controlPlane.resources.limits.memory 512Mi is below 1Gi", "the request the operator uses by default", "set requests.memory too")
}

func TestCheck_ARequestAboveTheDefaultLimitNamesTheLimitToSet(t *testing.T) {
	v := Check(&v1.PlantonPlatformSpec{Runner: &v1.RunnerSpec{
		Resources: &v1.ComponentResources{Requests: list("memory", "3Gi")},
	}})
	mustContain(t, v.Message, "spec.runner.resources.requests.memory 3Gi is above 2Gi", "the limit the operator uses by default", "set limits.memory too, at least 3Gi")
}

func TestCheck_BothHalvesDeclaredAndCrossed(t *testing.T) {
	v := Check(&v1.PlantonPlatformSpec{Console: &v1.ConsoleSpec{
		Resources: &v1.ComponentResources{Requests: list("cpu", "2"), Limits: list("cpu", "1")},
	}})
	mustContain(t, v.Message, "spec.console.resources.requests.cpu 2 is above its limit 1", "lower the request or raise the limit")
}

// Neo4j's chart refuses requests under 500m CPU; the operator says so first,
// in words about the field the person wrote.
func TestCheck_BelowAChartsFloor(t *testing.T) {
	v := Check(&v1.PlantonPlatformSpec{Components: &v1.ComponentsSpec{Graph: &v1.Neo4jSpec{
		Resources: &v1.ComponentResources{Requests: list("cpu", "250m")},
	}}})
	mustContain(t, v.Message, "spec.components.graph.resources.requests.cpu 250m is below 500m", "the least this workload's chart will run with")
}

func TestCheck_TheStoresCeilingMustStayBelowItsLimit(t *testing.T) {
	v := Check(&v1.PlantonPlatformSpec{Database: &v1.DatabaseSpec{Redis: &v1.RedisSpec{MaxMemory: "1gb"}}})
	mustContain(t, v.Message, "spec.database.redis.maxMemory 1gb is not below the store's memory limit 1Gi", "raise spec.database.redis.resources.limits.memory above it")

	raised := Check(&v1.PlantonPlatformSpec{Database: &v1.DatabaseSpec{Redis: &v1.RedisSpec{
		MaxMemory: "1gb",
		Resources: &v1.ComponentResources{Limits: list("memory", "2Gi")},
	}}})
	if !raised.Runnable {
		t.Fatalf("a limit raised above the ceiling must be runnable, got %s", raised.Message)
	}
}

// Every problem is named, not only the first, so one edit fixes them all.
func TestCheck_EveryProblemIsNamed(t *testing.T) {
	v := Check(&v1.PlantonPlatformSpec{
		ControlPlane: &v1.ControlPlaneSpec{Resources: &v1.ComponentResources{Limits: list("memory", "512Mi")}},
		Gateway:      &v1.GatewaySpec{Resources: &v1.ComponentResources{Limits: list("memory", "32Mi")}},
	})
	mustContain(t, v.Message, "spec.controlPlane.resources", "spec.gateway.resources")
}

func TestValkeyBytes(t *testing.T) {
	for in, want := range map[string]int64{"768mb": 768 << 20, "2gb": 2 << 30, "1g": 1e9, "5k": 5000, "100": 100, "64b": 64} {
		if got, ok := valkeyBytes(in); !ok || got != want {
			t.Errorf("%s: expected %d, got %d (%v)", in, want, got, ok)
		}
	}
}
