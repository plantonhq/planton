package generators

import (
	"strings"
	"testing"

	"github.com/hashicorp/hcl/v2/hclparse"

	"github.com/plantonhq/planton/catalog/kubernetes"
	peerauthv1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetespeerauthentication/v1alpha1"
	kubernetesvalkeyv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesvalkey/v1alpha1"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

// newPeerAuthManifest builds a KubernetesPeerAuthentication (a manifest-projection
// kind) exercising: a flattened namespace foreign key, a multi-word nested key
// (selector.match_labels -> selector.matchLabels), and an enum-like string.
func newPeerAuthManifest() *peerauthv1.KubernetesPeerAuthentication {
	return &peerauthv1.KubernetesPeerAuthentication{
		ApiVersion: "kubernetes.planton.dev/v1alpha1",
		Kind:       "KubernetesPeerAuthentication",
		Metadata:   &shared.CatalogObjectMetadata{Name: "pa-one"},
		Spec: &peerauthv1.KubernetesPeerAuthenticationSpec{
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
	valkey := &kubernetesvalkeyv1alpha1.KubernetesValkey{
		ApiVersion: "kubernetes.planton.dev/v1alpha1",
		Kind:       "KubernetesValkey",
		Metadata:   &shared.CatalogObjectMetadata{Name: "vlk-one"},
		Spec: &kubernetesvalkeyv1alpha1.KubernetesValkeySpec{
			Config: &kubernetesvalkeyv1alpha1.KubernetesValkeyConfig{
				MaxMemory:  "256mb",
				AppendOnly: true,
			},
		},
	}
	renderValkey, err := RenderTFVars(valkey)
	if err != nil {
		t.Fatalf("RenderTFVars(provider): %v", err)
	}
	if !strings.Contains(renderValkey, "append_only") || strings.Contains(renderValkey, "appendOnly") {
		t.Errorf("provider-abstraction kind should stay snake_case:\n%s", renderValkey)
	}
}
