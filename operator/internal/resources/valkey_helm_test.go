package resources

import (
	"strings"
	"testing"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// valkeyTestOptions is the persisted default posture with the sizing the
// component would resolve, so a test that cares about one knob names only it.
func valkeyTestOptions(crName, storageSize, storageClass string) ValkeyHelmOptions {
	return ValkeyHelmOptions{
		CRName:          crName,
		Persistence:     true,
		StorageSize:     storageSize,
		StorageClass:    storageClass,
		MaxMemory:       ValkeyDefaultMaxMemory,
		MaxMemoryPolicy: ValkeyDefaultMaxMemoryPolicy,
		Resources:       ValkeyDefaultResources(),
	}
}

func TestValkeyHelmValues_FullnameOverride(t *testing.T) {
	vals := ValkeyHelmValues(valkeyTestOptions("my-planton", "1Gi", ""))
	if vals["fullnameOverride"] != "my-planton-redis" {
		t.Errorf("expected fullnameOverride my-planton-redis, got %v", vals["fullnameOverride"])
	}
}

func TestValkeyHelmValues_Architecture(t *testing.T) {
	vals := ValkeyHelmValues(valkeyTestOptions("my-planton", "1Gi", ""))
	if vals["architecture"] != "standalone" {
		t.Errorf("expected architecture standalone, got %v", vals["architecture"])
	}
}

func TestValkeyHelmValues_Auth(t *testing.T) {
	vals := ValkeyHelmValues(valkeyTestOptions("my-planton", "1Gi", ""))
	auth, ok := vals["auth"].(map[string]any)
	if !ok {
		t.Fatal("expected auth to be a map")
	}
	if auth["existingSecret"] != "my-planton-redis-credentials" {
		t.Errorf("expected existingSecret my-planton-redis-credentials, got %v", auth["existingSecret"])
	}
	if auth["existingSecretPasswordKey"] != RedisSecretKey {
		t.Errorf("expected existingSecretPasswordKey %s, got %v", RedisSecretKey, auth["existingSecretPasswordKey"])
	}
}

func TestValkeyHelmValues_Persistence(t *testing.T) {
	vals := ValkeyHelmValues(valkeyTestOptions("my-planton", "5Gi", ""))
	primary, ok := vals["primary"].(map[string]any)
	if !ok {
		t.Fatal("expected primary to be a map")
	}
	persistence, ok := primary["persistence"].(map[string]any)
	if !ok {
		t.Fatal("expected persistence to be a map")
	}
	if persistence["size"] != "5Gi" {
		t.Errorf("expected persistence size 5Gi, got %v", persistence["size"])
	}
	if persistence["enabled"] != true {
		t.Error("expected persistence enabled")
	}
	if _, set := persistence["storageClass"]; set {
		t.Error("an unset storage class must be omitted, never rendered as \"\" (which disables provisioning)")
	}
}

func TestValkeyHelmValues_Image(t *testing.T) {
	// The image must stay pinned to the BSD-licensed Valkey engine; a drift
	// back to a Redis 8+ image would reintroduce the RSALv2/SSPLv1/AGPLv3
	// licensing family the platform deliberately avoids distributing or
	// directing self-hosted customers to run.
	vals := ValkeyHelmValues(valkeyTestOptions("my-planton", "1Gi", ""))
	image, ok := vals["image"].(map[string]any)
	if !ok {
		t.Fatal("expected image to be a map")
	}
	if image["repository"] != "bitnamilegacy/valkey" {
		t.Errorf("expected repository bitnamilegacy/valkey, got %v", image["repository"])
	}
}

func TestRedisSecretName(t *testing.T) {
	name := RedisSecretName("my-planton")
	if name != "my-planton-redis-credentials" {
		t.Errorf("expected my-planton-redis-credentials, got %s", name)
	}
}

func TestRedisServiceHost(t *testing.T) {
	host := RedisServiceHost("my-planton", "planton-system")
	expected := "my-planton-redis-primary.planton-system.svc.cluster.local"
	if host != expected {
		t.Errorf("expected %s, got %s", expected, host)
	}
}

// renderValkey renders the embedded chart and hands back the primary
// StatefulSet and the server ConfigMap: sizing is proven at the RENDER level
// (the chart's resourcesPreset silently wins over a values key it ignores),
// the same discipline storage_render_test.go applies to volumes.
func renderValkey(t *testing.T, opts ValkeyHelmOptions) (sts *unstructured.Unstructured, conf string) {
	t.Helper()
	objs, err := RenderHelmChart(LoadValkeyChart(), "test-redis", "default", ValkeyHelmValues(opts))
	if err != nil {
		t.Fatalf("failed to render Valkey chart: %v", err)
	}
	for _, obj := range objs {
		switch obj.GetKind() {
		case "StatefulSet":
			sts = obj
		case "ConfigMap":
			if data, found, _ := unstructured.NestedString(obj.Object, "data", "valkey.conf"); found {
				conf = data
			}
		}
	}
	if sts == nil {
		t.Fatal("expected a StatefulSet in the rendered objects")
	}
	if conf == "" {
		t.Fatal("expected the server ConfigMap (valkey.conf) in the rendered objects")
	}
	return sts, conf
}

func TestValkeyHelmValues_ChartRendering(t *testing.T) {
	chartData := LoadValkeyChart()
	if len(chartData) == 0 {
		t.Fatal("Valkey chart data is empty")
	}
	sts, _ := renderValkey(t, valkeyTestOptions("test", "1Gi", ""))
	// The readiness check and RedisServiceHost depend on the chart's
	// "primary" naming; a chart bump that renames the workload must fail here
	// rather than in a live cluster.
	if sts.GetName() != "test-redis-primary" {
		t.Errorf("expected StatefulSet test-redis-primary, got %s", sts.GetName())
	}
}

// The store's sizing lands on the container Kubernetes will run, not only in
// the values map: without an explicit resources block the chart applies its
// "nano" preset (a 192Mi limit), which is how a 158Mi dataset was OOM-killed
// into a crash loop. Requests for both, a memory limit, no CPU limit.
func TestValkeyRender_ResourceFloorReachesTheContainer(t *testing.T) {
	sts, _ := renderValkey(t, valkeyTestOptions("test", "1Gi", ""))
	containers, _, _ := unstructured.NestedSlice(sts.Object, "spec", "template", "spec", "containers")
	if len(containers) == 0 {
		t.Fatal("expected a container on the StatefulSet")
	}
	container := containers[0].(map[string]any)
	cpuReq, _, _ := unstructured.NestedString(container, "resources", "requests", "cpu")
	memReq, _, _ := unstructured.NestedString(container, "resources", "requests", "memory")
	memLim, _, _ := unstructured.NestedString(container, "resources", "limits", "memory")
	_, cpuLimited, _ := unstructured.NestedString(container, "resources", "limits", "cpu")
	if cpuReq != valkeyDefaultCPURequest || memReq != valkeyDefaultMemoryRequest {
		t.Errorf("expected requests %s/%s on the container, got %s/%s (the chart's preset must not win)",
			valkeyDefaultCPURequest, valkeyDefaultMemoryRequest, cpuReq, memReq)
	}
	if memLim != valkeyDefaultMemoryLimit {
		t.Errorf("expected the memory limit %s on the container, got %q", valkeyDefaultMemoryLimit, memLim)
	}
	if cpuLimited {
		t.Error("CPU must not be limited (requests-only, the house pattern)")
	}
}

// The ceiling and the eviction policy are the fix for the crash loop, and the
// persisted posture keeps the append-only file (the build-log stream survives
// a restart) with snapshots off. Proven on the mounted valkey.conf.
func TestValkeyRender_ServerConfigurationCarriesTheCeilingAndPersistence(t *testing.T) {
	sts, conf := renderValkey(t, valkeyTestOptions("test", "1Gi", ""))
	for _, line := range []string{
		"maxmemory " + ValkeyDefaultMaxMemory,
		"maxmemory-policy " + ValkeyDefaultMaxMemoryPolicy,
		"appendonly yes",
		`save ""`,
	} {
		if !strings.Contains(conf, line) {
			t.Errorf("valkey.conf must carry %q; got:\n%s", line, conf)
		}
	}
	vcts, _, _ := unstructured.NestedSlice(sts.Object, "spec", "volumeClaimTemplates")
	if len(vcts) != 1 {
		t.Errorf("the persisted posture renders one volume claim template, got %d", len(vcts))
	}
}

// With persistence off the store is a pure in-memory cache: no claim template
// (an emptyDir), no append-only file to rewrite for nothing, the ceiling and
// eviction unchanged.
func TestValkeyRender_PersistenceOffRendersNoClaimAndNoAppendOnlyFile(t *testing.T) {
	opts := valkeyTestOptions("test", "1Gi", "")
	opts.Persistence = false
	sts, conf := renderValkey(t, opts)
	if vcts, found, _ := unstructured.NestedSlice(sts.Object, "spec", "volumeClaimTemplates"); found && len(vcts) > 0 {
		t.Errorf("persistence off must render no volume claim template; got %v", vcts)
	}
	if !strings.Contains(conf, "appendonly no") {
		t.Errorf("persistence off must turn the append-only file off; got:\n%s", conf)
	}
	if !strings.Contains(conf, "maxmemory "+ValkeyDefaultMaxMemory) {
		t.Errorf("the ceiling applies in both postures; got:\n%s", conf)
	}
}

// A platform that sizes the store itself is honored verbatim, in the kind's
// vocabulary, and the ceiling it names reaches the server.
func TestValkeyRender_DeclaredSizingWins(t *testing.T) {
	opts := valkeyTestOptions("test", "1Gi", "")
	opts.MaxMemory = "2gb"
	opts.MaxMemoryPolicy = "volatile-ttl"
	opts.Resources = corev1.ResourceRequirements{
		Requests: corev1.ResourceList{corev1.ResourceMemory: resource.MustParse("1Gi")},
		Limits:   corev1.ResourceList{corev1.ResourceMemory: resource.MustParse("3Gi")},
	}
	sts, conf := renderValkey(t, opts)
	if !strings.Contains(conf, "maxmemory 2gb") || !strings.Contains(conf, "maxmemory-policy volatile-ttl") {
		t.Errorf("the declared ceiling and policy must reach valkey.conf; got:\n%s", conf)
	}
	containers, _, _ := unstructured.NestedSlice(sts.Object, "spec", "template", "spec", "containers")
	memLim, _, _ := unstructured.NestedString(containers[0].(map[string]any), "resources", "limits", "memory")
	if memLim != "3Gi" {
		t.Errorf("the declared memory limit must reach the container, got %q", memLim)
	}
	if _, set, _ := unstructured.NestedString(containers[0].(map[string]any), "resources", "requests", "cpu"); set {
		t.Error("a request the platform did not declare must not be invented")
	}
}
