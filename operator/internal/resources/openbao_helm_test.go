package resources

import (
	"strings"
	"testing"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestOpenBAOHelmValues_FullnameOverride(t *testing.T) {
	vals := OpenBAOHelmValues("my-planton", "10Gi", "")
	if vals["fullnameOverride"] != "my-planton-openbao" {
		t.Errorf("expected fullnameOverride my-planton-openbao, got %v", vals["fullnameOverride"])
	}
}

func TestOpenBAOHelmValues_Global(t *testing.T) {
	vals := OpenBAOHelmValues("my-planton", "10Gi", "")
	global, ok := vals["global"].(map[string]any)
	if !ok {
		t.Fatal("expected global to be a map")
	}
	if global["enabled"] != true {
		t.Error("expected global enabled")
	}
	if global["tlsDisable"] != true {
		t.Error("expected global tlsDisable true")
	}
}

func TestOpenBAOHelmValues_Standalone(t *testing.T) {
	vals := OpenBAOHelmValues("my-planton", "10Gi", "")
	server, ok := vals["server"].(map[string]any)
	if !ok {
		t.Fatal("expected server to be a map")
	}
	standalone, ok := server["standalone"].(map[string]any)
	if !ok {
		t.Fatal("expected standalone to be a map")
	}
	if standalone["enabled"] != true {
		t.Error("expected standalone enabled")
	}
	config, ok := standalone["config"].(string)
	if !ok || config == "" {
		t.Fatal("expected non-empty standalone config string")
	}
}

func TestOpenBAOHelmValues_HADisabled(t *testing.T) {
	vals := OpenBAOHelmValues("my-planton", "10Gi", "")
	server := vals["server"].(map[string]any)
	ha, ok := server["ha"].(map[string]any)
	if !ok {
		t.Fatal("expected ha to be a map")
	}
	if ha["enabled"] != false {
		t.Error("expected ha disabled")
	}
}

func TestOpenBAOHelmValues_Storage(t *testing.T) {
	vals := OpenBAOHelmValues("my-planton", "20Gi", "")
	server := vals["server"].(map[string]any)
	dataStorage, ok := server["dataStorage"].(map[string]any)
	if !ok {
		t.Fatal("expected dataStorage to be a map")
	}
	if dataStorage["size"] != "20Gi" {
		t.Errorf("expected size 20Gi, got %v", dataStorage["size"])
	}
	if dataStorage["enabled"] != true {
		t.Error("expected dataStorage enabled")
	}
}

func TestOpenBAOHelmValues_UI(t *testing.T) {
	vals := OpenBAOHelmValues("my-planton", "10Gi", "")
	ui, ok := vals["ui"].(map[string]any)
	if !ok {
		t.Fatal("expected ui to be a map")
	}
	if ui["enabled"] != true {
		t.Error("expected ui enabled")
	}
}

func TestOpenBAOHelmValues_InjectorDisabled(t *testing.T) {
	vals := OpenBAOHelmValues("my-planton", "10Gi", "")
	injector, ok := vals["injector"].(map[string]any)
	if !ok {
		t.Fatal("expected injector to be a map")
	}
	if injector["enabled"] != false {
		t.Error("expected injector disabled")
	}
}

func TestOpenBAOReleaseName(t *testing.T) {
	if got := openbaoReleaseName("my-planton"); got != "my-planton-openbao" {
		t.Errorf("expected my-planton-openbao, got %s", got)
	}
}

func TestOpenBAOInitSecretName(t *testing.T) {
	if got := OpenBAOInitSecretName("my-planton"); got != "my-planton-openbao-init" {
		t.Errorf("expected my-planton-openbao-init, got %s", got)
	}
}

// The init Secret's note must explain the key material in plain language:
// what it unlocks, the cost of deleting it, and the own-your-own-Secret
// alternative. Deployed by default, this Secret exists on every install.
func TestOpenBAOInitSecretNote(t *testing.T) {
	note := OpenBAOInitSecretNote("my-planton")
	for _, want := range []string{
		"my-planton-openbao",
		"unseal",
		"Deleting this Secret",
		"spec.vault.initSecretName",
	} {
		if !strings.Contains(note, want) {
			t.Errorf("init Secret note must mention %q, got: %s", want, note)
		}
	}
}

// Deployed on every default install, the vault must schedule honestly:
// explicit requests, a memory limit, and (house pattern) no CPU limit.
func TestOpenBAOHelmValues_Resources(t *testing.T) {
	vals := OpenBAOHelmValues("my-planton", "2Gi", "")
	server := vals["server"].(map[string]any)
	res, ok := server["resources"].(map[string]any)
	if !ok {
		t.Fatal("expected server.resources to be set -- the chart ships none")
	}
	requests := res["requests"].(map[string]any)
	if requests["cpu"] == "" || requests["memory"] == "" {
		t.Errorf("expected cpu+memory requests, got %v", requests)
	}
	limits := res["limits"].(map[string]any)
	if limits["memory"] == "" {
		t.Errorf("expected a memory limit, got %v", limits)
	}
	if _, hasCPULimit := limits["cpu"]; hasCPULimit {
		t.Error("no CPU limit by design (requests-only, the house pattern)")
	}
}

func TestOpenBAOServiceHost(t *testing.T) {
	expected := "my-planton-openbao.planton-system.svc.cluster.local"
	if got := OpenBAOServiceHost("my-planton", "planton-system"); got != expected {
		t.Errorf("expected %s, got %s", expected, got)
	}
}

func TestOpenBAOAPIAddr(t *testing.T) {
	expected := "http://my-planton-openbao.default.svc.cluster.local:8200"
	if got := OpenBAOAPIAddr("my-planton", "default"); got != expected {
		t.Errorf("expected %s, got %s", expected, got)
	}
}

func TestOpenBAOHelmValues_ChartRendering(t *testing.T) {
	chartData := LoadOpenBAOChart()
	if len(chartData) == 0 {
		t.Fatal("OpenBAO chart data is empty")
	}

	values := OpenBAOHelmValues("test", "10Gi", "")
	objs, err := RenderHelmChart(chartData, "test-openbao", "default", values)
	if err != nil {
		t.Fatalf("failed to render OpenBAO chart: %v", err)
	}
	if len(objs) == 0 {
		t.Fatal("expected rendered objects, got none")
	}

	const kindStatefulSet = "StatefulSet"

	kinds := make(map[string]bool)
	for _, obj := range objs {
		kinds[obj.GetKind()] = true
	}

	if !kinds[kindStatefulSet] {
		t.Error("expected StatefulSet in rendered objects")
	}
	if !kinds["Service"] {
		t.Error("expected Service in rendered objects")
	}

	for _, obj := range objs {
		if obj.GetKind() == "ClusterRoleBinding" && obj.GetName() == "test-openbao-server-binding" {
			t.Error("auth-delegator ClusterRoleBinding must be disabled -- the operator RBAC cannot grant tokenreview permissions")
		}
	}

	// Render-level assertion (the inert-values-key lesson): the chart must
	// actually thread server.resources onto the container.
	for _, obj := range objs {
		if obj.GetKind() != kindStatefulSet {
			continue
		}
		containers, _, _ := unstructured.NestedSlice(obj.Object,
			"spec", "template", "spec", "containers")
		if len(containers) == 0 {
			t.Fatal("rendered StatefulSet has no containers")
		}
		c := containers[0].(map[string]any)
		cpu, _, _ := unstructured.NestedString(c, "resources", "requests", "cpu")
		memLimit, _, _ := unstructured.NestedString(c, "resources", "limits", "memory")
		if cpu == "" || memLimit == "" {
			t.Errorf("rendered container must carry the requests + memory limit, got resources=%v", c["resources"])
		}
	}
}
