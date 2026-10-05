package refannotations

import (
	"testing"

	testgenericv1alpha2 "github.com/plantonhq/planton/catalog/_test/testcatalogkindgeneric/v1alpha2"
	digitaloceanprojectv1alpha1 "github.com/plantonhq/planton/catalog/digitalocean/digitaloceanproject/v1alpha1"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func genericField(t *testing.T, name protoreflect.Name) Field {
	t.Helper()
	fd := (&testgenericv1alpha2.TestCatalogKindGenericSpec{}).ProtoReflect().Descriptor().Fields().ByName(name)
	if fd == nil {
		t.Fatalf("fixture field %s missing", name)
	}
	return Of(fd)
}

const (
	generic    = catalogkind.CatalogKind_TestCatalogKindGeneric
	kubernetes = catalogkind.CatalogKind_TestCatalogKindKubernetes
)

func TestOf_DefaultOnly(t *testing.T) {
	f := genericField(t, "annotated_ref")
	assert.True(t, f.IsReference())
	assert.Equal(t, []Key{{Kind: generic, FieldPath: "status.outputs.id"}}, f.Keys())
	assert.Equal(t, []catalogkind.CatalogKind{generic}, f.Kinds())
	path, ok := f.DefaultPath(generic)
	assert.True(t, ok)
	assert.Equal(t, "status.outputs.id", path)
	assert.Equal(t, generic, f.EffectiveKind(catalogkind.CatalogKind_unspecified))
}

func TestOf_CandidatesOnly(t *testing.T) {
	fd := (&digitaloceanprojectv1alpha1.DigitalOceanProjectSpec{}).ProtoReflect().Descriptor().Fields().ByName("resources")
	f := Of(fd)
	assert.True(t, f.IsReference())
	assert.Equal(t, catalogkind.CatalogKind_unspecified, f.DefaultKind)
	assert.Len(t, f.Keys(), 8)
	assert.Equal(t, catalogkind.CatalogKind_DigitalOceanDroplet, f.Kinds()[0])
	assert.Equal(t, catalogkind.CatalogKind_unspecified, f.EffectiveKind(catalogkind.CatalogKind_unspecified),
		"a candidate list names no default: a reference must say its kind")
}

func TestOf_DefaultAndCandidatesListTheDefaultOnce(t *testing.T) {
	f := genericField(t, "candidate_ref")
	assert.Equal(t, []Key{
		{Kind: generic, FieldPath: "status.outputs.id"},
		{Kind: kubernetes, FieldPath: "status.outputs.external_hostname"},
		{Kind: kubernetes, FieldPath: "status.outputs.internal_hostname"},
	}, f.Keys())
	assert.Equal(t, []catalogkind.CatalogKind{generic, kubernetes}, f.Kinds())
}

func TestOf_AKindWithTwoKeysHasNoDefaultPath(t *testing.T) {
	f := genericField(t, "candidate_ref")
	_, ok := f.DefaultPath(kubernetes)
	assert.False(t, ok)
	assert.True(t, f.AcceptsPath(kubernetes, "status.outputs.internal_hostname"))
	assert.True(t, f.AcceptsPath(kubernetes, "status.outputs.internal_hostname.0"), "a path that extends a key is accepted")
	assert.False(t, f.AcceptsPath(kubernetes, "status.outputs.endpoint"))
	assert.True(t, f.AcceptsPath(catalogkind.CatalogKind_KubernetesNamespace, "spec.name"), "a kind with no key is unjudged")
}

func TestOf_NoAnnotations(t *testing.T) {
	f := genericField(t, "optional_ref")
	assert.False(t, f.IsReference())
	assert.Empty(t, f.Keys())
}
