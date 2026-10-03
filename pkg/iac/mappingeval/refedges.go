package mappingeval

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/plantonhq/planton/pkg/refannotations"
	"github.com/plantonhq/planton/shared/catalogkind"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

const stringValueOrRefFullName = "dev.planton.shared.foreignkey.v1.StringValueOrRef"

// RefEdge is one value_from reference found in a manifest's spec: the spec
// location that consumes a value, and the (kind, name) of the resource that
// produces it. Edges are the refs axis's unit of comparison -- the scorer
// checks that a proposal has an edge at the same spec location pointing at
// the matched counterpart of the same producer.
type RefEdge struct {
	// FieldPath locates the consuming field inside the spec, with repeated
	// elements indexed (e.g. "vpc_id", "routes[0].target_id"). Proto field
	// names, matching how the field is declared.
	FieldPath string
	// TargetKind is the producer's kind: the reference's own kind when set,
	// otherwise the consuming field's default_kind annotation. unspecified
	// when neither declares it (the edge is then compared by name alone).
	TargetKind catalogkind.CatalogKind
	// TargetName is the producer's metadata.name as the reference states it.
	TargetName string
}

// ExtractRefEdges walks a manifest's spec tree and returns every value_from
// reference as an edge. The walk mirrors the platform resolver's scope:
// StringValueOrRef fields anywhere in the spec -- singular, repeated, and
// nested inside (repeated) messages. Map-typed fields are not traversed (no
// spec models refs inside maps today).
func ExtractRefEdges(manifest proto.Message) ([]RefEdge, error) {
	top := manifest.ProtoReflect()
	specField := top.Descriptor().Fields().ByName("spec")
	if specField == nil || specField.Kind() != protoreflect.MessageKind {
		return nil, errors.Errorf("%s is not a KRM envelope (no spec message)", top.Descriptor().FullName())
	}
	if !top.Has(specField) {
		return nil, nil
	}
	var edges []RefEdge
	collectRefEdges(top.Get(specField).Message(), "", &edges)
	return edges, nil
}

// collectRefEdges recursively gathers value_from edges under path.
func collectRefEdges(msg protoreflect.Message, path string, edges *[]RefEdge) {
	fields := msg.Descriptor().Fields()
	for i := 0; i < fields.Len(); i++ {
		fd := fields.Get(i)
		if fd.Kind() != protoreflect.MessageKind || fd.IsMap() {
			continue
		}
		isRef := string(fd.Message().FullName()) == stringValueOrRefFullName

		if fd.IsList() {
			if !msg.Has(fd) {
				continue
			}
			list := msg.Get(fd).List()
			for j := 0; j < list.Len(); j++ {
				elementPath := fmt.Sprintf("%s[%d]", joinPath(path, string(fd.Name())), j)
				if isRef {
					appendRefEdge(fd, list.Get(j).Message(), elementPath, edges)
					continue
				}
				collectRefEdges(list.Get(j).Message(), elementPath, edges)
			}
			continue
		}

		if !msg.Has(fd) {
			continue
		}
		fieldPath := joinPath(path, string(fd.Name()))
		if isRef {
			appendRefEdge(fd, msg.Get(fd).Message(), fieldPath, edges)
			continue
		}
		collectRefEdges(msg.Get(fd).Message(), fieldPath, edges)
	}
}

// appendRefEdge records the edge when the StringValueOrRef carries the
// value_from arm; literals contribute no edge (their absence at a location
// where the ground truth has one is exactly what the refs axis penalizes).
func appendRefEdge(fd protoreflect.FieldDescriptor, refMsg protoreflect.Message, fieldPath string, edges *[]RefEdge) {
	ref, ok := refMsg.Interface().(*foreignkeyv1.StringValueOrRef)
	if !ok || ref.GetValueFrom() == nil {
		return
	}
	kind := refannotations.Of(fd).EffectiveKind(ref.GetValueFrom().GetKind())
	*edges = append(*edges, RefEdge{
		FieldPath:  fieldPath,
		TargetKind: kind,
		TargetName: ref.GetValueFrom().GetName(),
	})
}

func joinPath(path, segment string) string {
	if path == "" {
		return segment
	}
	return path + "." + segment
}
