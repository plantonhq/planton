package manifestprojection

import (
	"github.com/pkg/errors"
	"github.com/plantonhq/planton/pkg/iac/specprojection"
	"github.com/plantonhq/planton/shared"
	"github.com/plantonhq/planton/shared/catalogkind"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// Manifest is the shape every catalog manifest message has.
type Manifest interface {
	proto.Message
	GetKind() string
	GetMetadata() *shared.CatalogObjectMetadata
}

// Object is the custom resource a projection kind's manifest renders to.
type Object struct {
	APIVersion  string
	Kind        string
	Name        string
	Namespace   string
	Labels      map[string]string
	Annotations map[string]string
	// Spec is the custom resource's spec: the projected spec minus the
	// envelope, keyed by the custom resource's own JSON keys.
	Spec map[string]interface{}
}

// Render turns a projection kind's manifest into the object both engines
// apply. References in the spec must already be resolved to literal values,
// which the platform does before any module runs.
func Render(m Manifest) (*Object, error) {
	kind, ok := catalogkind.CatalogKind_value[m.GetKind()]
	if !ok {
		return nil, errors.Errorf("unknown kind %q", m.GetKind())
	}
	proj := ProjectionOf(catalogkind.CatalogKind(kind))
	if proj == nil {
		return nil, errors.Errorf("%s is not a Kubernetes manifest projection kind", m.GetKind())
	}

	specField := m.ProtoReflect().Descriptor().Fields().ByName("spec")
	if specField == nil || specField.Kind() != protoreflect.MessageKind {
		return nil, errors.Errorf("%s manifest has no message-typed spec", m.GetKind())
	}
	specMsg := m.ProtoReflect().Get(specField).Message().Interface()

	spec, err := specprojection.Project(specMsg, specprojection.JSONKeys)
	if err != nil {
		return nil, errors.Wrapf(err, "project %s spec", m.GetKind())
	}

	envelope := EnvelopeOf(specField.Message())
	obj := &Object{
		APIVersion: proj.GetApiVersion(),
		Kind:       proj.GetKind(),
		Name:       m.GetMetadata().GetName(),
	}
	if envelope.Namespaced() {
		ns, _ := spec[envelope.NamespaceKey].(string)
		obj.Namespace = ns
	}
	ownLabels := stringMap(spec[envelope.LabelsKey])
	if envelope.AnnotationsKey != "" {
		if a := stringMap(spec[envelope.AnnotationsKey]); len(a) > 0 {
			obj.Annotations = a
		}
	}
	for key := range spec {
		if envelope.Holds(key) {
			delete(spec, key)
		}
	}
	obj.Spec = spec

	md := m.GetMetadata()
	obj.Labels = MergeLabels(ownLabels, IdentityLabels(Identity{
		Kind:         m.GetKind(),
		Name:         md.GetName(),
		ID:           md.GetId(),
		Organization: md.GetOrg(),
		Environment:  md.GetEnv(),
	}))
	return obj, nil
}

// ProjectionOf returns a kind's kubernetes_manifest_projection, or nil when
// the kind is not a projection kind. It reads the kind enum's options
// directly, so callers need not link the registry's message map.
func ProjectionOf(kind catalogkind.CatalogKind) *catalogkind.KubernetesManifestProjection {
	value := kind.Descriptor().Values().ByNumber(kind.Number())
	if value == nil || value.Options() == nil {
		return nil
	}
	meta, _ := proto.GetExtension(value.Options(), catalogkind.E_KindMeta).(*catalogkind.CatalogKindMeta)
	return meta.GetKubernetesManifestProjection()
}

func stringMap(v interface{}) map[string]string {
	raw, ok := v.(map[string]interface{})
	if !ok {
		return nil
	}
	out := make(map[string]string, len(raw))
	for k, val := range raw {
		if s, ok := val.(string); ok {
			out[k] = s
		}
	}
	return out
}
