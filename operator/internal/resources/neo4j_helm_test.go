package resources

import (
	"testing"
)

func TestNeo4jHelmValues_Name(t *testing.T) {
	vals := Neo4jHelmValues("my-planton", "10Gi", "", Effective(SizingNeo4j, nil))
	neo4j, ok := vals["neo4j"].(map[string]any)
	if !ok {
		t.Fatal("expected neo4j to be a map")
	}
	if neo4j["name"] != "my-planton" {
		t.Errorf("expected name my-planton, got %v", neo4j["name"])
	}
}

func TestNeo4jHelmValues_LicenseAgreement(t *testing.T) {
	vals := Neo4jHelmValues("my-planton", "10Gi", "", Effective(SizingNeo4j, nil))
	neo4j := vals["neo4j"].(map[string]any)
	if neo4j["acceptLicenseAgreement"] != "yes" {
		t.Errorf("expected acceptLicenseAgreement yes, got %v", neo4j["acceptLicenseAgreement"])
	}
}

func TestNeo4jHelmValues_Resources(t *testing.T) {
	vals := Neo4jHelmValues("my-planton", "10Gi", "", Effective(SizingNeo4j, nil))
	neo4j := vals["neo4j"].(map[string]any)
	resources, ok := neo4j["resources"].(map[string]any)
	if !ok {
		t.Fatal("expected resources to be a map")
	}
	// The requests/limits form, which the chart takes whenever either key is
	// present: the house pattern keeps the memory limit and drops the CPU one.
	requests, _ := resources["requests"].(map[string]any)
	limits, _ := resources["limits"].(map[string]any)
	if requests["cpu"] != "1" || requests["memory"] != "2Gi" {
		t.Errorf("expected requests cpu 1 and memory 2Gi, got %v", requests)
	}
	if limits["memory"] != "2Gi" {
		t.Errorf("expected a 2Gi memory limit, got %v", limits)
	}
	if _, cpuLimited := limits["cpu"]; cpuLimited {
		t.Error("no CPU limit by design (requests-only, the house pattern)")
	}
}

// The chart's _volumeTemplate.tpl reads the size ONLY from
// volumes.data.<mode>.requests.storage (a bare "size" key is silently
// ignored -- the defect that made spec.components.graph.storageSize inert).
func TestNeo4jHelmValues_Storage(t *testing.T) {
	vals := Neo4jHelmValues("my-planton", "20Gi", "", Effective(SizingNeo4j, nil))
	data := vals["volumes"].(map[string]any)["data"].(map[string]any)
	if data["mode"] != "defaultStorageClass" {
		t.Errorf("expected mode defaultStorageClass, got %v", data["mode"])
	}
	requests := data["defaultStorageClass"].(map[string]any)["requests"].(map[string]any)
	if requests["storage"] != "20Gi" {
		t.Errorf("expected requests.storage 20Gi, got %v", requests["storage"])
	}
	if _, exists := data["size"]; exists {
		t.Error("a bare volumes.data.size key is ignored by the chart and must not be emitted")
	}
}

// A pinned StorageClass must switch the volume to "dynamic" mode: the chart
// REJECTS a storageClassName under "defaultStorageClass" mode.
func TestNeo4jHelmValues_StorageClassSwitchesToDynamicMode(t *testing.T) {
	vals := Neo4jHelmValues("my-planton", "20Gi", "fast-ssd", Effective(SizingNeo4j, nil))
	data := vals["volumes"].(map[string]any)["data"].(map[string]any)
	if data["mode"] != "dynamic" {
		t.Errorf("expected mode dynamic when a class is pinned, got %v", data["mode"])
	}
	dynamic := data["dynamic"].(map[string]any)
	if dynamic["storageClassName"] != "fast-ssd" {
		t.Errorf("expected storageClassName fast-ssd, got %v", dynamic["storageClassName"])
	}
	if dynamic["requests"].(map[string]any)["storage"] != "20Gi" {
		t.Errorf("expected requests.storage 20Gi, got %v", dynamic["requests"])
	}
}

func TestNeo4jReleaseName(t *testing.T) {
	if got := neo4jReleaseName("my-planton"); got != "my-planton-neo4j" {
		t.Errorf("expected my-planton-neo4j, got %s", got)
	}
}

func TestNeo4jAuthSecretName(t *testing.T) {
	if got := Neo4jAuthSecretName("my-planton"); got != "my-planton-neo4j-auth" {
		t.Errorf("expected my-planton-neo4j-auth, got %s", got)
	}
}

func TestNeo4jServiceHost(t *testing.T) {
	expected := "my-planton-neo4j.planton-system.svc.cluster.local"
	if got := Neo4jServiceHost("my-planton", "planton-system"); got != expected {
		t.Errorf("expected %s, got %s", expected, got)
	}
}

func TestNeo4jBoltURI(t *testing.T) {
	expected := "bolt://my-planton-neo4j.default.svc.cluster.local:7687"
	if got := Neo4jBoltURI("my-planton", "default"); got != expected {
		t.Errorf("expected %s, got %s", expected, got)
	}
}

func TestNeo4jHelmValues_ChartRendering(t *testing.T) {
	chartData := LoadNeo4jChart()
	if len(chartData) == 0 {
		t.Fatal("Neo4j chart data is empty")
	}

	values := Neo4jHelmValues("test", "10Gi", "", Effective(SizingNeo4j, nil))
	objs, err := RenderHelmChart(chartData, "test-neo4j", "default", values)
	if err != nil {
		t.Fatalf("failed to render Neo4j chart: %v", err)
	}
	if len(objs) == 0 {
		t.Fatal("expected rendered objects, got none")
	}

	kinds := make(map[string]bool)
	for _, obj := range objs {
		kinds[obj.GetKind()] = true
	}

	if !kinds["StatefulSet"] {
		t.Error("expected StatefulSet in rendered objects")
	}
	if !kinds["Service"] {
		t.Error("expected Service in rendered objects")
	}
}
