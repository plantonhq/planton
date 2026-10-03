package specprojection

import (
	"reflect"
	"testing"

	"github.com/plantonhq/planton/catalog/kubernetes"
	kubernetesservicemonitorv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesservicemonitor/v1alpha1"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

// These tests pin the two Kubernetes shape markers on a real projection kind's
// spec, through the same Project call both engines' modules make: an
// IntOrString written as a number or a name, and a list-valued map's wrappers
// written as bare lists with their elements projected (references collapsed)
// inside. They also pin that an explicitly empty object survives, since a
// required-but-empty selector means "select everything" upstream.

func literal(v string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: v}}
}

func monitorSpec(targetPort string) *kubernetesservicemonitorv1alpha1.KubernetesServiceMonitorSpec {
	return &kubernetesservicemonitorv1alpha1.KubernetesServiceMonitorSpec{
		Namespace: literal("monitoring"),
		Selector:  &kubernetes.KubernetesPrometheusOperatorApiLabelSelector{},
		Endpoints: []*kubernetesservicemonitorv1alpha1.KubernetesServiceMonitorEndpoint{{
			TargetPort: &targetPort,
			Params: map[string]*kubernetes.KubernetesPrometheusOperatorApiStringList{
				"module": {Values: []string{"http_2xx", "tcp"}},
				"empty":  {},
			},
			ProxyUrl: stringPtr("http://proxy:3128"),
			ProxyConnectHeader: map[string]*kubernetes.KubernetesPrometheusOperatorApiSecretKeySelectorList{
				"Proxy-Authorization": {Values: []*kubernetes.KubernetesPrometheusOperatorApiSecretKeySelector{
					{Name: literal("proxy-credentials"), Key: "header"},
				}},
			},
		}},
	}
}

func projectedEndpoint(t *testing.T, targetPort string) map[string]interface{} {
	t.Helper()
	data, err := Project(monitorSpec(targetPort), JSONKeys)
	if err != nil {
		t.Fatalf("Project: %v", err)
	}
	endpoints, ok := data["endpoints"].([]interface{})
	if !ok || len(endpoints) != 1 {
		t.Fatalf("endpoints = %#v, want one endpoint", data["endpoints"])
	}
	return endpoints[0].(map[string]interface{})
}

func TestProject_IntOrString_DigitsBecomeANumber(t *testing.T) {
	got := projectedEndpoint(t, "9090")["targetPort"]
	if got != float64(9090) {
		t.Errorf("targetPort = %#v (%T), want the number 9090", got, got)
	}
}

func TestProject_IntOrString_ANameStaysAString(t *testing.T) {
	got := projectedEndpoint(t, "http-metrics")["targetPort"]
	if got != "http-metrics" {
		t.Errorf("targetPort = %#v (%T), want the name \"http-metrics\"", got, got)
	}
}

func TestProject_ListValuedMap_WrappersBecomeBareLists(t *testing.T) {
	got := projectedEndpoint(t, "9090")["params"]
	want := map[string]interface{}{
		"module": []interface{}{"http_2xx", "tcp"},
		"empty":  []interface{}{},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("params = %#v, want %#v", got, want)
	}
}

func TestProject_ListValuedMap_ElementsAreProjected(t *testing.T) {
	got := projectedEndpoint(t, "9090")["proxyConnectHeader"]
	want := map[string]interface{}{
		"Proxy-Authorization": []interface{}{
			map[string]interface{}{"name": "proxy-credentials", "key": "header"},
		},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("proxyConnectHeader = %#v, want %#v (the reference collapsed to its name inside the list)", got, want)
	}
}

func TestProject_ListValuedMap_SnakeCaseStyleUnwrapsToo(t *testing.T) {
	data, err := Project(monitorSpec("9090"), SnakeCaseKeys)
	if err != nil {
		t.Fatalf("Project: %v", err)
	}
	endpoint := data["endpoints"].([]interface{})[0].(map[string]interface{})
	want := map[string]interface{}{
		"module": []interface{}{"http_2xx", "tcp"},
		"empty":  []interface{}{},
	}
	if got := endpoint["params"]; !reflect.DeepEqual(got, want) {
		t.Errorf("params = %#v, want %#v", got, want)
	}
	if got := endpoint["target_port"]; got != float64(9090) {
		t.Errorf("target_port = %#v, want the number 9090", got)
	}
}

func TestProject_EmptySelectorIsKept(t *testing.T) {
	data, err := Project(monitorSpec("9090"), JSONKeys)
	if err != nil {
		t.Fatalf("Project: %v", err)
	}
	selector, present := data["selector"]
	if !present {
		t.Fatalf("selector missing from the projection; an explicitly empty selector selects everything and must reach the object")
	}
	if got, ok := selector.(map[string]interface{}); !ok || len(got) != 0 {
		t.Errorf("selector = %#v, want an empty object", selector)
	}
}
