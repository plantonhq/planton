package resources

import (
	"strings"
	"testing"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestOpenBAOHelmValues_FullnameOverride(t *testing.T) {
	vals := OpenBAOHelmValues("my-planton", "my-planton-openbao-storage")
	if vals["fullnameOverride"] != "my-planton-openbao" {
		t.Errorf("expected fullnameOverride my-planton-openbao, got %v", vals["fullnameOverride"])
	}
}

func TestOpenBAOHelmValues_Global(t *testing.T) {
	vals := OpenBAOHelmValues("my-planton", "my-planton-openbao-storage")
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
	vals := OpenBAOHelmValues("my-planton", "my-planton-openbao-storage")
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
	vals := OpenBAOHelmValues("my-planton", "my-planton-openbao-storage")
	server := vals["server"].(map[string]any)
	ha, ok := server["ha"].(map[string]any)
	if !ok {
		t.Fatal("expected ha to be a map")
	}
	if ha["enabled"] != false {
		t.Error("expected ha disabled")
	}
}

// The vault's storage is the platform's PostgreSQL: the config names the
// backend and its pool cap and NOTHING credential-bearing (the chart renders
// it into a ConfigMap); the connection URL reaches the process as the
// variable the backend reads, projected from the storage Secret; and the
// chart's volume is off.
func TestOpenBAOHelmValues_Storage(t *testing.T) {
	vals := OpenBAOHelmValues("my-planton", "my-planton-openbao-storage")
	server := vals["server"].(map[string]any)

	config := server["standalone"].(map[string]any)["config"].(string)
	if !strings.Contains(config, `storage "postgresql"`) {
		t.Errorf("config must name the postgresql backend, got:\n%s", config)
	}
	if strings.Contains(config, `storage "file"`) || strings.Contains(config, "/openbao/data") {
		t.Errorf("config must not name a file backend or a data path, got:\n%s", config)
	}
	if !strings.Contains(config, `max_parallel = "32"`) {
		t.Errorf("config must cap the backend's pool against the cluster's shared connection budget, got:\n%s", config)
	}
	for _, forbidden := range []string{"connection_url", "postgres://", "password"} {
		if strings.Contains(config, forbidden) {
			t.Errorf("config must carry nothing credential-bearing; found %q in:\n%s", forbidden, config)
		}
	}

	dataStorage, ok := server["dataStorage"].(map[string]any)
	if !ok || dataStorage["enabled"] != false {
		t.Errorf("dataStorage must be disabled -- the vault has no volume; got %v", server["dataStorage"])
	}

	env, ok := server["extraSecretEnvironmentVars"].([]any)
	if !ok || len(env) != 1 {
		t.Fatalf("expected exactly one secret environment variable, got %v", server["extraSecretEnvironmentVars"])
	}
	entry := env[0].(map[string]any)
	if entry["envName"] != OpenBAOStorageURLEnv || entry["secretName"] != "my-planton-openbao-storage" || entry["secretKey"] != OpenBAOStorageURLEnv {
		t.Errorf("the storage URL must reach the process as %s from the storage Secret's key of the same name, got %v", OpenBAOStorageURLEnv, entry)
	}
}

// The connection URL is composed in one place: the vault's own role and
// database on the platform's primary, with the platform's in-cluster TLS
// posture, and the password escaped so any Secret value parses.
func TestOpenBAOStorageURL(t *testing.T) {
	got := OpenBAOStorageURL("my-planton", "planton", "p@ss/word")
	want := "postgres://openbao:p%40ss%2Fword@my-planton-postgres-rw.planton.svc.cluster.local:5432/openbao?sslmode=disable"
	if got != want {
		t.Errorf("URL\n got: %s\nwant: %s", got, want)
	}
	if got := OpenBAOStorageSecretName("my-planton"); got != "my-planton-openbao-storage" {
		t.Errorf("storage Secret = %s", got)
	}
}

func TestOpenBAOHelmValues_UI(t *testing.T) {
	vals := OpenBAOHelmValues("my-planton", "my-planton-openbao-storage")
	ui, ok := vals["ui"].(map[string]any)
	if !ok {
		t.Fatal("expected ui to be a map")
	}
	if ui["enabled"] != true {
		t.Error("expected ui enabled")
	}
}

func TestOpenBAOHelmValues_InjectorDisabled(t *testing.T) {
	vals := OpenBAOHelmValues("my-planton", "my-planton-openbao-storage")
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
	vals := OpenBAOHelmValues("my-planton", "my-planton-openbao-storage")
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

	values := OpenBAOHelmValues("test", "test-openbao-storage")
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

		// The storage seams, at the render: the connection URL reaches the
		// process as the backend's variable from the storage Secret's key,
		// and the data mount the file backend needed is gone with the volume.
		envs, _, _ := unstructured.NestedSlice(c, "env")
		var urlEnv map[string]any
		for _, raw := range envs {
			if e, ok := raw.(map[string]any); ok && e["name"] == OpenBAOStorageURLEnv {
				urlEnv = e
			}
		}
		if urlEnv == nil {
			t.Fatalf("rendered container carries no %s variable; env=%v", OpenBAOStorageURLEnv, envs)
		}
		secretName, _, _ := unstructured.NestedString(urlEnv, "valueFrom", "secretKeyRef", "name")
		secretKey, _, _ := unstructured.NestedString(urlEnv, "valueFrom", "secretKeyRef", "key")
		if secretName != "test-openbao-storage" || secretKey != OpenBAOStorageURLEnv {
			t.Errorf("%s must be projected from the storage Secret's key of the same name, got %v", OpenBAOStorageURLEnv, urlEnv)
		}
		mounts, _, _ := unstructured.NestedSlice(c, "volumeMounts")
		for _, raw := range mounts {
			if m, ok := raw.(map[string]any); ok && m["mountPath"] == "/openbao/data" {
				t.Errorf("the vault has no volume, yet the container mounts a data path: %v", m)
			}
		}
		if vcts, found, _ := unstructured.NestedSlice(obj.Object, "spec", "volumeClaimTemplates"); found && len(vcts) > 0 {
			t.Errorf("the vault has no volume, yet the StatefulSet carries claim templates: %v", vcts)
		}
	}
}
