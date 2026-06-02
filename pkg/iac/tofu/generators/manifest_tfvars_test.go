package generators

import (
	"strings"
	"testing"

	"github.com/hashicorp/hcl/v2/hclparse"

	"github.com/plantonhq/openmcf/apis/org/openmcf/provider/kubernetes"
	peerauthv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/kubernetes/kubernetespeerauthentication/v1"
	kubernetesredisv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/kubernetes/kubernetesredis/v1"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared"
	foreignkeyv1 "github.com/plantonhq/openmcf/apis/org/openmcf/shared/foreignkey/v1"
)

// newPeerAuthManifest builds a KubernetesPeerAuthentication (a manifest-projection
// kind) exercising: a flattened namespace foreign key, a multi-word nested key
// (selector.match_labels -> selector.matchLabels), an enum-like string, and a
// skipped orchestrator field (target_cluster).
func newPeerAuthManifest() *peerauthv1.KubernetesPeerAuthentication {
	return &peerauthv1.KubernetesPeerAuthentication{
		ApiVersion: "kubernetes.openmcf.org/v1",
		Kind:       "KubernetesPeerAuthentication",
		Metadata:   &shared.CloudResourceMetadata{Name: "pa-one"},
		Spec: &peerauthv1.KubernetesPeerAuthenticationSpec{
			TargetCluster: &kubernetes.KubernetesClusterSelector{ClusterName: "c1"},
			Namespace: &foreignkeyv1.StringValueOrRef{
				LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "mesh-ns"},
			},
			Selector: &kubernetes.KubernetesIstioApiWorkloadSelector{
				MatchLabels: map[string]string{"app": "web"},
			},
			Mtls: &peerauthv1.KubernetesPeerAuthenticationMutualTls{Mode: "STRICT"},
		},
	}
}

func TestProtoToManifestTFVars_CamelCasePrunedAndFlattened(t *testing.T) {
	got, err := ProtoToManifestTFVars(newPeerAuthManifest())
	if err != nil {
		t.Fatalf("ProtoToManifestTFVars: %v", err)
	}

	// camelCase CRD keys are preserved (not renamed to snake_case).
	if !strings.Contains(got, "matchLabels") {
		t.Errorf("expected camelCase key matchLabels, got:\n%s", got)
	}
	if strings.Contains(got, "match_labels") {
		t.Errorf("snake_case match_labels must not appear in manifest mode, got:\n%s", got)
	}

	// StringValueOrRef namespace is flattened to a plain string.
	if !strings.Contains(got, `"namespace" = "mesh-ns"`) && !strings.Contains(got, `namespace = "mesh-ns"`) {
		t.Errorf("namespace should be a flat string, got:\n%s", got)
	}
	if strings.Contains(got, `"value" = "mesh-ns"`) {
		t.Errorf("namespace should not appear as a nested {value} object, got:\n%s", got)
	}

	// Orchestrator-only field is skipped.
	if strings.Contains(got, "target_cluster") || strings.Contains(got, "targetCluster") {
		t.Errorf("target_cluster must be skipped, got:\n%s", got)
	}

	// protojson omits unset fields, so the manifest carries no nulls -- this is
	// why the projection module needs no oneOf/required-subfield pruning.
	if strings.Contains(got, "= null") {
		t.Errorf("manifest tfvars must not contain nulls, got:\n%s", got)
	}

	parser := hclparse.NewParser()
	if _, diags := parser.ParseHCL([]byte(got), "test.tfvars"); diags.HasErrors() {
		t.Errorf("generated manifest tfvars is not valid HCL: %s\n%s", diags.Error(), got)
	}
}

// TestRenderTFVars_DispatchesByKind proves the single kind-aware entry point picks
// the camelCase manifest path for a projection kind and the snake_case path for a
// provider-abstraction kind, so every runtime caller stays in sync with the module.
func TestRenderTFVars_DispatchesByKind(t *testing.T) {
	// HCL is emitted by iterating Go maps, so key ORDER is non-deterministic;
	// assert on structural markers (key casing) rather than exact string equality.

	// Projection kind -> camelCase manifest path.
	renderPA, err := RenderTFVars(newPeerAuthManifest())
	if err != nil {
		t.Fatalf("RenderTFVars(projection): %v", err)
	}
	if !strings.Contains(renderPA, "matchLabels") || strings.Contains(renderPA, "match_labels") {
		t.Errorf("projection kind should render camelCase via the manifest path:\n%s", renderPA)
	}

	// Provider-abstraction kind -> snake_case path (the converter must not be
	// flipped globally; only annotated kinds switch).
	redis := &kubernetesredisv1.KubernetesRedis{
		ApiVersion: "kubernetes.openmcf.org/v1",
		Kind:       "KubernetesRedis",
		Metadata:   &shared.CloudResourceMetadata{Name: "red-one"},
		Spec: &kubernetesredisv1.KubernetesRedisSpec{
			Container: &kubernetesredisv1.KubernetesRedisContainer{
				DiskSize:           "2Gi",
				PersistenceEnabled: true,
				Replicas:           1,
			},
		},
	}
	renderRedis, err := RenderTFVars(redis)
	if err != nil {
		t.Fatalf("RenderTFVars(provider): %v", err)
	}
	if !strings.Contains(renderRedis, "persistence_enabled") || strings.Contains(renderRedis, "persistenceEnabled") {
		t.Errorf("provider-abstraction kind should stay snake_case:\n%s", renderRedis)
	}
}
