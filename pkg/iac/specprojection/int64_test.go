package specprojection

import (
	"reflect"
	"testing"

	"github.com/plantonhq/planton/catalog/kubernetes"
	kubernetesgkecomputeclassv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesgkecomputeclass/v1alpha1"
	stripecouponv1alpha1 "github.com/plantonhq/planton/catalog/stripe/stripecoupon/v1alpha1"
)

// These tests pin how 64-bit integers leave the projection. Protojson writes
// them as strings; the manifest-projection path (JSONKeys) must hand a
// Kubernetes API server numbers, because it refuses a string for an integer
// property, while the snake_case path keeps the strings its typed Terraform
// variables already convert. Singular, repeated and map-valued fields are all
// covered, and a value too large to be exact as a JSON number stays a string
// rather than being rounded.

func int64Ptr(v int64) *int64 { return &v }

func TestProject_Int64_SingularBecomesANumberForAKubernetesObject(t *testing.T) {
	data, err := Project(&kubernetesgkecomputeclassv1alpha1.KubernetesGkeComputeClassGpu{Count: int64Ptr(2)}, JSONKeys)
	if err != nil {
		t.Fatalf("Project: %v", err)
	}
	if got := data["count"]; got != float64(2) {
		t.Errorf("count = %#v (%T), want the number 2", got, got)
	}
}

func TestProject_Int64_RepeatedElementsBecomeNumbers(t *testing.T) {
	data, err := Project(&kubernetes.WorkloadPodSecurityContext{SupplementalGroups: []int64{1000, 2000}}, JSONKeys)
	if err != nil {
		t.Fatalf("Project: %v", err)
	}
	want := []interface{}{float64(1000), float64(2000)}
	if got := data["supplementalGroups"]; !reflect.DeepEqual(got, want) {
		t.Errorf("supplementalGroups = %#v, want %#v", got, want)
	}
}

func TestProject_Int64_MapValuesBecomeNumbers(t *testing.T) {
	data, err := Project(&stripecouponv1alpha1.StripeCouponSpec{CurrencyOptions: map[string]int64{"eur": 500}}, JSONKeys)
	if err != nil {
		t.Fatalf("Project: %v", err)
	}
	want := map[string]interface{}{"eur": float64(500)}
	if got := data["currencyOptions"]; !reflect.DeepEqual(got, want) {
		t.Errorf("currencyOptions = %#v, want %#v", got, want)
	}
}

func TestProject_Int64_BeyondFiftyThreeBitsStaysExact(t *testing.T) {
	data, err := Project(&kubernetesgkecomputeclassv1alpha1.KubernetesGkeComputeClassGpu{Count: int64Ptr(1<<53 + 1)}, JSONKeys)
	if err != nil {
		t.Fatalf("Project: %v", err)
	}
	if got := data["count"]; got != "9007199254740993" {
		t.Errorf("count = %#v (%T), want the exact string 9007199254740993", got, got)
	}
}

func TestProject_Int64_SnakeCaseKeysKeepProtojsonStrings(t *testing.T) {
	data, err := Project(&kubernetes.WorkloadPodSecurityContext{SupplementalGroups: []int64{1000}}, SnakeCaseKeys)
	if err != nil {
		t.Fatalf("Project: %v", err)
	}
	want := []interface{}{"1000"}
	if got := data["supplemental_groups"]; !reflect.DeepEqual(got, want) {
		t.Errorf("supplemental_groups = %#v, want %#v", got, want)
	}
}
