package manifestgraph

import (
	"fmt"

	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

const stringValueOrRefFullName = "dev.planton.shared.foreignkey.v1.StringValueOrRef"

// RefUse is one populated valueFrom reference found in a manifest: the spec
// field it sits on (whose descriptor carries the foreign-key annotations),
// the reference itself, and the capability to replace the whole
// StringValueOrRef in place — so collection, validation, edge derivation,
// and literal rewriting all ride ONE traversal and can never disagree about
// what counts as a reference.
type RefUse struct {
	// FieldPath is the proto field-name dot path to the reference,
	// e.g. "spec.kms_key_name", "spec.iam_members[0].member", or
	// "spec.ref_map.primary" for a map entry.
	FieldPath string

	// Field is the field declaring the StringValueOrRef — the carrier of the
	// default_kind / default_kind_field_path / containment_exempt annotations.
	Field protoreflect.FieldDescriptor

	// Ref is the populated reference.
	Ref *foreignkeyv1.ValueFromRef

	// replace swaps the whole StringValueOrRef value at this position.
	replace func(*foreignkeyv1.StringValueOrRef)
}

// Replace substitutes the StringValueOrRef at this reference's position —
// the resolution seam. Passing a literal-armed value is how a resolved
// reference becomes a plain value.
func (u RefUse) Replace(v *foreignkeyv1.StringValueOrRef) {
	u.replace(v)
}

// LiteralUse is one populated LITERAL arm of a StringValueOrRef found in a
// manifest: the same field-and-path shape as RefUse, with the authored
// string instead of a reference. A literal is not a reference — the
// platform never resolves it — but on a field whose annotation declares a
// default kind it may NAME a sibling in the same set (a route naming its
// gateway, a certificate naming its issuer), which is an ordering fact the
// graph derives (see BuildGraph, the literal-sibling source).
type LiteralUse struct {
	// FieldPath is the proto field-name dot path to the literal, in the same
	// grammar RefUse.FieldPath uses.
	FieldPath string

	// Field is the field declaring the StringValueOrRef — the carrier of the
	// default_kind annotation the literal is matched against.
	Field protoreflect.FieldDescriptor

	// Value is the authored literal.
	Value string
}

// CollectRefUses walks every populated field of a loaded manifest — singular,
// repeated, nested, and MAP-typed containers included — and returns each
// StringValueOrRef carrying a valueFrom reference. Literal values are not
// references and are skipped.
func CollectRefUses(msg proto.Message) []RefUse {
	refs, _ := collectUses(msg)
	return refs
}

// CollectLiteralUses walks the same fields CollectRefUses walks and returns
// each StringValueOrRef carrying a non-empty literal. References are skipped.
func CollectLiteralUses(msg proto.Message) []LiteralUse {
	_, literals := collectUses(msg)
	return literals
}

// collectUses is the ONE traversal both collectors ride: every populated
// StringValueOrRef is sorted into its arm here, so a reference and a literal
// can never be found by two walkers that disagree about a container.
func collectUses(msg proto.Message) ([]RefUse, []LiteralUse) {
	var refs []RefUse
	var literals []LiteralUse
	walkPopulated(msg.ProtoReflect(), "", func(fd protoreflect.FieldDescriptor, path string,
		svor *foreignkeyv1.StringValueOrRef, replace func(*foreignkeyv1.StringValueOrRef)) {
		switch {
		case svor.GetValueFrom() != nil:
			refs = append(refs, RefUse{FieldPath: path, Field: fd, Ref: svor.GetValueFrom(), replace: replace})
		case svor.GetValue() != "":
			literals = append(literals, LiteralUse{FieldPath: path, Field: fd, Value: svor.GetValue()})
		}
	})
	return refs, literals
}

// visitor receives every populated StringValueOrRef the traversal meets: the
// field declaring it (the annotation carrier), its dot path, the value, and
// the capability to replace it in place.
type visitor func(fd protoreflect.FieldDescriptor, path string,
	svor *foreignkeyv1.StringValueOrRef, replace func(*foreignkeyv1.StringValueOrRef))

func walkPopulated(m protoreflect.Message, prefix string, visit visitor) {
	m.Range(func(fd protoreflect.FieldDescriptor, v protoreflect.Value) bool {
		path := string(fd.Name())
		if prefix != "" {
			path = prefix + "." + path
		}
		switch {
		case fd.IsMap():
			if fd.MapValue().Kind() != protoreflect.MessageKind {
				return true
			}
			mp := v.Map()
			mp.Range(func(k protoreflect.MapKey, mv protoreflect.Value) bool {
				visitMessage(fd, mv.Message(), path+"."+k.String(), visit,
					func(replacement *foreignkeyv1.StringValueOrRef) {
						mp.Set(k, protoreflect.ValueOfMessage(replacement.ProtoReflect()))
					})
				return true
			})
		case fd.IsList():
			if fd.Kind() != protoreflect.MessageKind {
				return true
			}
			list := v.List()
			for i := 0; i < list.Len(); i++ {
				idx := i
				visitMessage(fd, list.Get(i).Message(), fmt.Sprintf("%s[%d]", path, i), visit,
					func(replacement *foreignkeyv1.StringValueOrRef) {
						list.Set(idx, protoreflect.ValueOfMessage(replacement.ProtoReflect()))
					})
			}
		case fd.Kind() == protoreflect.MessageKind:
			visitMessage(fd, v.Message(), path, visit,
				func(replacement *foreignkeyv1.StringValueOrRef) {
					m.Set(fd, protoreflect.ValueOfMessage(replacement.ProtoReflect()))
				})
		}
		return true
	})
}

// visitMessage either hands a populated StringValueOrRef to the visitor (the
// annotations are read off the declaring field fd) or recurses.
func visitMessage(fd protoreflect.FieldDescriptor, m protoreflect.Message, path string, visit visitor,
	replace func(*foreignkeyv1.StringValueOrRef)) {
	md := m.Descriptor()
	if string(md.FullName()) == stringValueOrRefFullName {
		if svor, ok := m.Interface().(*foreignkeyv1.StringValueOrRef); ok {
			visit(fd, path, svor, replace)
		}
		return
	}
	walkPopulated(m, path, visit)
}
