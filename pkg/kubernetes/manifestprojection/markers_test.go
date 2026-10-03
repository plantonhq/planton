package manifestprojection_test

import (
	"testing"

	"github.com/plantonhq/planton/pkg/crkreflect"
	"github.com/plantonhq/planton/pkg/iac/specprojection"
	"github.com/plantonhq/planton/pkg/kubernetes/manifestprojection"
	"github.com/plantonhq/planton/shared/options"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// TestProjectionMarkers_SitWhereTheyAreHonored walks every registered kind's
// spec, through every message it reaches, and holds the three projection
// markers to the one place each is honored: the manifest projection of a
// Kubernetes custom resource. A marker anywhere else would be read by no
// engine, so the field would reach its module in the wrong shape.
//
//   - kubernetes_object_metadata: a top-level map<string, string> of the spec.
//   - kubernetes_int_or_string: a singular string field.
//   - kubernetes_list_valued_map: a map whose value is a message with exactly
//     one field, and that field repeated.
//
// And each only where a projection kind's spec reaches it: a shared message
// that carries a marker may be used by projection kinds alone.
func TestProjectionMarkers_SitWhereTheyAreHonored(t *testing.T) {
	marked := 0
	for _, kind := range crkreflect.KindsList() {
		msg, err := crkreflect.NewInstance(kind)
		if err != nil {
			continue
		}
		specField := msg.ProtoReflect().Descriptor().Fields().ByName("spec")
		if specField == nil || specField.Kind() != protoreflect.MessageKind {
			continue
		}
		spec := specField.Message()
		projection := manifestprojection.ProjectionOf(kind) != nil

		visited := map[protoreflect.FullName]bool{}
		var walk func(md protoreflect.MessageDescriptor)
		walk = func(md protoreflect.MessageDescriptor) {
			if visited[md.FullName()] {
				return
			}
			visited[md.FullName()] = true
			fields := md.Fields()
			for i := 0; i < fields.Len(); i++ {
				fd := fields.Get(i)
				objectMetadata := manifestprojection.ObjectMetadataOf(fd) != options.KubernetesObjectMetadataField_kubernetes_object_metadata_field_unspecified
				intOrString := specprojection.IsKubernetesIntOrStringField(fd)
				listValuedMap := specprojection.IsKubernetesListValuedMapField(fd)

				if objectMetadata || intOrString || listValuedMap {
					marked++
					if !projection {
						t.Errorf("%s: %s carries a projection marker, but %s is not a Kubernetes manifest projection kind, so no engine honors it", kind, fd.FullName(), kind)
					}
				}
				if objectMetadata && !(md == spec && fd.IsMap() && fd.MapValue().Kind() == protoreflect.StringKind) {
					t.Errorf("%s: kubernetes_object_metadata on %s, which is not a top-level map<string, string> of the spec", kind, fd.FullName())
				}
				if intOrString && (fd.Kind() != protoreflect.StringKind || fd.IsList() || fd.IsMap()) {
					t.Errorf("%s: kubernetes_int_or_string on %s, which is not a singular string field", kind, fd.FullName())
				}
				if listValuedMap && !isListWrapperMap(fd) {
					t.Errorf("%s: kubernetes_list_valued_map on %s, which is not a map whose value message holds exactly one repeated field", kind, fd.FullName())
				}

				switch {
				case fd.IsMap():
					if fd.MapValue().Kind() == protoreflect.MessageKind {
						walk(fd.MapValue().Message())
					}
				case fd.Kind() == protoreflect.MessageKind:
					walk(fd.Message())
				}
			}
		}
		walk(spec)
	}
	if marked == 0 {
		t.Fatal("no projection marker found in any kind; the walk is not reaching the specs")
	}
}

func isListWrapperMap(fd protoreflect.FieldDescriptor) bool {
	if !fd.IsMap() || fd.MapValue().Kind() != protoreflect.MessageKind {
		return false
	}
	wrapper := fd.MapValue().Message().Fields()
	return wrapper.Len() == 1 && wrapper.Get(0).IsList()
}
