package manifestgraph

import (
	"testing"

	testk8sv1alpha1 "github.com/plantonhq/planton/catalog/_test/testcatalogkindkubernetes/v1alpha1"
	auth0connectionv1alpha1 "github.com/plantonhq/planton/catalog/auth0/auth0connection/v1alpha1"
	"github.com/plantonhq/planton/shared"
	"github.com/plantonhq/planton/shared/catalogkind"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
	"github.com/stretchr/testify/assert"
)

func ref(kind catalogkind.CatalogKind, name, fieldPath string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
			ValueFrom: &foreignkeyv1.ValueFromRef{Kind: kind, Name: name, FieldPath: fieldPath},
		},
	}
}

func literal(v string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: v}}
}

// TestResolveRefs_SingularAndMapContainers pins the substitution seam across
// the container shapes the old spec-tree walker could not rewrite (maps) and
// the plain singular case.
func TestResolveRefs_SingularAndMapContainers(t *testing.T) {
	msg := &testk8sv1alpha1.TestCatalogKindKubernetes{
		Metadata: &shared.CatalogObjectMetadata{Name: "consumer", Env: "dev"},
		Spec: &testk8sv1alpha1.TestCatalogKindKubernetesSpec{
			Namespace: ref(catalogkind.CatalogKind_KubernetesNamespace, "apps", "spec.name"),
			RefMap: map[string]*foreignkeyv1.StringValueOrRef{
				"primary": ref(catalogkind.CatalogKind_TestCatalogKindGeneric, "producer", "status.outputs.id"),
				"keep":    literal("untouched"),
			},
		},
	}

	lookup := func(id Identity) (map[string]string, bool) {
		switch id.Slug {
		case "apps":
			return map[string]string{"spec.name": "apps-real"}, true
		case "producer":
			return map[string]string{"id": "prod-123"}, true
		}
		return nil, false
	}

	resolved, findings := ResolveRefs(msg, "dev", lookup)

	assert.Equal(t, 2, resolved)
	assert.Empty(t, findings)
	assert.Equal(t, "apps-real", msg.Spec.Namespace.GetValue())
	assert.Equal(t, "prod-123", msg.Spec.RefMap["primary"].GetValue())
	assert.Equal(t, "untouched", msg.Spec.RefMap["keep"].GetValue())
}

// TestResolveRefs_ListContainer pins substitution inside repeated
// StringValueOrRef fields, elements resolving independently so literals and
// references mix.
func TestResolveRefs_ListContainer(t *testing.T) {
	msg := &auth0connectionv1alpha1.Auth0Connection{
		Metadata: &shared.CatalogObjectMetadata{Name: "conn", Env: "dev"},
		Spec: &auth0connectionv1alpha1.Auth0ConnectionSpec{
			EnabledClients: []*foreignkeyv1.StringValueOrRef{
				literal("client-literal"),
				ref(catalogkind.CatalogKind_Auth0Client, "web-app", "status.outputs.client_id"),
			},
		},
	}

	lookup := func(id Identity) (map[string]string, bool) {
		if id.Slug == "web-app" {
			return map[string]string{"client_id": "auth0|abc"}, true
		}
		return nil, false
	}

	resolved, findings := ResolveRefs(msg, "dev", lookup)

	assert.Equal(t, 1, resolved)
	assert.Empty(t, findings)
	assert.Equal(t, "client-literal", msg.Spec.EnabledClients[0].GetValue())
	assert.Equal(t, "auth0|abc", msg.Spec.EnabledClients[1].GetValue())
}

// TestResolveRefs_FindingClasses pins the two failure classes apart: a
// target with no outputs (unresolved — the reference stays) and a resolved
// target missing the referenced field (missing-output — the composition
// itself is wrong).
func TestResolveRefs_FindingClasses(t *testing.T) {
	msg := &testk8sv1alpha1.TestCatalogKindKubernetes{
		Metadata: &shared.CatalogObjectMetadata{Name: "consumer", Env: "dev"},
		Spec: &testk8sv1alpha1.TestCatalogKindKubernetesSpec{
			Namespace: ref(catalogkind.CatalogKind_KubernetesNamespace, "apps", "spec.name"),
			RefMap: map[string]*foreignkeyv1.StringValueOrRef{
				// The path is VALID on the kind's outputs proto — but the
				// deployed module's captured outputs won't carry it.
				"broken": ref(catalogkind.CatalogKind_TestCatalogKindGeneric, "producer", "status.outputs.id"),
			},
		},
	}

	lookup := func(id Identity) (map[string]string, bool) {
		if id.Slug == "producer" {
			return map[string]string{"other": "prod-123"}, true // no "id"
		}
		return nil, false // "apps" has no outputs available
	}

	resolved, findings := ResolveRefs(msg, "dev", lookup)

	assert.Equal(t, 0, resolved)
	classes := map[FindingClass]int{}
	for _, f := range findings {
		classes[f.Class]++
	}
	assert.Equal(t, 1, classes[FindingUnresolvedRef], "namespace target has no outputs")
	assert.Equal(t, 1, classes[FindingMissingOutput], "producer lacks the referenced output")
	// Both references stay untouched.
	assert.NotNil(t, msg.Spec.Namespace.GetValueFrom())
	assert.NotNil(t, msg.Spec.RefMap["broken"].GetValueFrom())
}
