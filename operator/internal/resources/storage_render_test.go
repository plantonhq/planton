package resources

import (
	"fmt"
	"testing"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// Storage settings are proven at the RENDER level, not the values level: each
// test renders the real embedded chart and asserts the storageClassName and
// storage request that land on the volume claims Kubernetes will see. A
// values-shape test cannot catch a key the chart silently ignores -- exactly
// how spec.components.graph.storageSize shipped inert (the Neo4j chart reads
// sizes only from <mode>.requests.storage).

// renderedVolumeClaims renders a chart and collects every volume claim across
// StatefulSet volumeClaimTemplates and standalone PVCs, keyed
// "{owner}/{claim}", valued {storageClassName (or "" when unset), size}.
func renderedVolumeClaims(t *testing.T, chartData []byte, releaseName string, values map[string]any) map[string][2]string {
	t.Helper()
	objs, err := RenderHelmChart(chartData, releaseName, "default", values)
	if err != nil {
		t.Fatalf("failed to render chart: %v", err)
	}
	claims := map[string][2]string{}
	record := func(owner, claim string, spec map[string]any) {
		class, _, _ := unstructured.NestedString(spec, "storageClassName")
		size, _, _ := unstructured.NestedString(spec, "resources", "requests", "storage")
		claims[fmt.Sprintf("%s/%s", owner, claim)] = [2]string{class, size}
	}
	for _, obj := range objs {
		switch obj.GetKind() {
		case "StatefulSet":
			vcts, found, _ := unstructured.NestedSlice(obj.Object, "spec", "volumeClaimTemplates")
			if !found {
				continue
			}
			for _, raw := range vcts {
				vct, ok := raw.(map[string]any)
				if !ok {
					continue
				}
				name, _, _ := unstructured.NestedString(vct, "metadata", "name")
				spec, _, _ := unstructured.NestedMap(vct, "spec")
				record(obj.GetName(), name, spec)
			}
		case "PersistentVolumeClaim":
			spec, _, _ := unstructured.NestedMap(obj.Object, "spec")
			record("pvc", obj.GetName(), spec)
		}
	}
	if len(claims) == 0 {
		t.Fatal("expected at least one volume claim in the rendered chart")
	}
	return claims
}

func assertAllClaims(t *testing.T, claims map[string][2]string, wantClass, wantSize string) {
	t.Helper()
	for name, got := range claims {
		if got[0] != wantClass {
			t.Errorf("claim %s: storageClassName = %q, want %q", name, got[0], wantClass)
		}
		if got[1] != wantSize {
			t.Errorf("claim %s: storage request = %q, want %q", name, got[1], wantSize)
		}
	}
}

func TestStorageRender_Valkey(t *testing.T) {
	pinned := renderedVolumeClaims(t, LoadValkeyChart(), "test-redis",
		ValkeyHelmValues(valkeyTestOptions("test", "800Gi", "trident")))
	assertAllClaims(t, pinned, "trident", "800Gi")

	unpinned := renderedVolumeClaims(t, LoadValkeyChart(), "test-redis",
		ValkeyHelmValues(valkeyTestOptions("test", "1Gi", "")))
	assertAllClaims(t, unpinned, "", "1Gi")
}

// The vault has no volume: its storage is the platform's database. The chart
// guards its claim template on one switch, and the operator turns it off, so
// the rendered StatefulSet must carry no claim at all -- a claim here would
// be a volume the archive does not cover.
func TestStorageRender_OpenBAO_NoVolume(t *testing.T) {
	objs, err := RenderHelmChart(LoadOpenBAOChart(), "test-openbao", "default",
		OpenBAOHelmValues(OpenBAOHelmOptions{CRName: "test", Namespace: "default", StoragePasswordSecretName: PostgreSQLVaultRoleSecretName("test")}))
	if err != nil {
		t.Fatalf("failed to render chart: %v", err)
	}
	for _, obj := range objs {
		switch obj.GetKind() {
		case "StatefulSet":
			if vcts, found, _ := unstructured.NestedSlice(obj.Object, "spec", "volumeClaimTemplates"); found && len(vcts) > 0 {
				t.Errorf("the vault must render no volume claim template; got %v", vcts)
			}
		case "PersistentVolumeClaim":
			t.Errorf("the vault must render no PersistentVolumeClaim; got %s", obj.GetName())
		}
	}
}

// Neo4j is the chart whose values contract already burned us once (the inert
// size key), so both mode arms are proven against the rendered claim.
func TestStorageRender_Neo4j(t *testing.T) {
	pinned := renderedVolumeClaims(t, LoadNeo4jChart(), "test-neo4j",
		Neo4jHelmValues("test", "800Gi", "trident"))
	assertAllClaims(t, pinned, "trident", "800Gi")

	unpinned := renderedVolumeClaims(t, LoadNeo4jChart(), "test-neo4j",
		Neo4jHelmValues("test", "20Gi", ""))
	assertAllClaims(t, unpinned, "", "20Gi")
}

// The CloudNativePG Cluster CR is not a Helm chart, but its storage block is
// the same contract class: assert the built CR carries class + size (and
// omits the class when unpinned -- the CNPG webhook rejects an empty string).
func TestStorageRender_PostgreSQLCluster(t *testing.T) {
	pinned := NewPostgreSQLCluster(PostgreSQLClusterOptions{
		CRName: "test", Namespace: "default", Instances: 1,
		StorageSize: "800Gi", StorageClassName: "trident",
	})
	storage, _, _ := unstructured.NestedMap(pinned.Object, "spec", "storage")
	if storage["size"] != "800Gi" {
		t.Errorf("storage.size = %v, want 800Gi", storage["size"])
	}
	if storage["storageClass"] != "trident" {
		t.Errorf("storage.storageClass = %v, want trident", storage["storageClass"])
	}

	unpinned := NewPostgreSQLCluster(PostgreSQLClusterOptions{
		CRName: "test", Namespace: "default", Instances: 1, StorageSize: "10Gi",
	})
	storage, _, _ = unstructured.NestedMap(unpinned.Object, "spec", "storage")
	if _, exists := storage["storageClass"]; exists {
		t.Error("storageClass must be omitted when unpinned")
	}
}
