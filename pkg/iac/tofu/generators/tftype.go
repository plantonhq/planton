package generators

import (
	"fmt"
	"strings"
)

// TFType represents a Terraform type expression. Implementations format
// themselves as valid HCL type constraint syntax.
type TFType interface {
	Format(indent int) string
}

// TFPrimitive represents a Terraform type rendered as a bare keyword: the
// primitives string, number, and bool, plus `any` for free-form JSON
// well-known types (google.protobuf.Struct/Value/ListValue) passed through
// verbatim.
type TFPrimitive string

func (p TFPrimitive) Format(_ int) string {
	return string(p)
}

// TFList represents a Terraform list type: list(elementType).
type TFList struct {
	Elem TFType
}

func (l TFList) Format(indent int) string {
	return fmt.Sprintf("list(%s)", l.Elem.Format(indent))
}

// TFMap represents a Terraform map type: map(valueType). Proto map<string, V>
// fields map to this type.
type TFMap struct {
	Value TFType
}

func (m TFMap) Format(indent int) string {
	return fmt.Sprintf("map(%s)", m.Value.Format(indent))
}

// TFFreeFormMap represents a proto map whose value is a free-form JSON well-known
// type (map<string, google.protobuf.Struct/Value/ListValue>). It renders as the
// bare `any` keyword rather than `map(any)`: each map entry is an arbitrary,
// independently-shaped JSON object, and Terraform's `map(any)` requires every
// element to converge to a single common type ("all map elements must have the
// same type"). Typing the whole attribute `any` lets Terraform infer a
// heterogeneous object instead. Its zero default is an empty map ({}) so a pruned
// field reconstructs to something a `for`/`for_each` can iterate.
type TFFreeFormMap struct{}

func (TFFreeFormMap) Format(_ int) string {
	return "any"
}

// TFFreeFormList represents a repeated proto field whose element type is
// free-form (a JSON well-known type, or a recursive message collapsed to
// `any`). It renders as the bare `any` keyword rather than `list(any)` for the
// same reason TFFreeFormMap exists: Terraform's `list(any)` requires every
// element to converge to a single common type ("all list elements must have
// the same type"), which arbitrary or recursive elements cannot guarantee.
// Its zero default is an empty list ([]) so a pruned field reconstructs to
// something a `for` expression can iterate.
type TFFreeFormList struct{}

func (TFFreeFormList) Format(_ int) string {
	return "any"
}

// TFObject represents a Terraform object type with named fields.
type TFObject struct {
	Fields []TFField
}

// TFField holds a single field within a TFObject.
//
// Optional marks the attribute as one the runtime tfvars renderer may omit.
// The renderer (ProtoToTFVars) marshals via protojson with EmitUnpopulated=false,
// so any proto field left at its zero value is absent from the emitted tfvars. A
// Terraform object type rejects a value that lacks a non-optional attribute, so
// every attribute that is not provably present must be declared optional with a
// default equal to the proto zero value -- otherwise the module fails input
// validation on a pruned tfvars. Required attributes (those the renderer always
// emits, identified from buf.validate constraints) stay bare.
//
// Presence marks a scalar with proto presence -- a proto3 `optional` field or
// a oneof member -- whose absence is meaningfully DIFFERENT from its zero value
// (a tri-state: the cloud default applies when unset, an explicit zero
// overrides it; or a sibling oneof arm is the one chosen). protojson emits
// such a field whenever it is set, including at its zero value, so the attribute must
// default to null (optional(type) with no literal) rather than to the zero
// value: collapsing "unset" into false/0 would silently override the cloud's
// own default on every resource that left the field out. Modules null-guard
// these attributes (`x == null ? ... : x`).
//
// Doc is the proto field's own documentation (the text above the field in
// the .proto source, read from the embedded protodocs index), rendered as
// `#` comment lines above the attribute. Nested object attributes cannot
// carry a Terraform `description`, so the comment is the only place the
// module can say what an input means -- and whoever wires main.tf reads the
// generated file first. Because the text is the proto's, it cannot drift
// from the spec the way a hand-written comment does. Note is the one extra
// sentence a type rule asks for when it collapses a wrapper message to a
// primitive (TypeRule.FlattenNote): the proto documentation describes the
// wrapper, the attribute is the primitive, and the note bridges the two.
type TFField struct {
	Name     string
	Type     TFType
	Optional bool
	Presence bool
	Doc      string
	Note     string
}

// commentLines returns the field's documentation and flatten note as the
// lines of a `#` comment block (without the marker), or nil when the field
// carries neither. Blank lines inside the documentation survive as bare `#`
// lines so paragraph breaks in the proto read as paragraph breaks here.
func (f TFField) commentLines() []string {
	var lines []string
	if doc := strings.TrimSpace(f.Doc); doc != "" {
		for _, l := range strings.Split(doc, "\n") {
			lines = append(lines, strings.TrimRight(l, " \t"))
		}
	}
	if note := strings.TrimSpace(f.Note); note != "" {
		lines = append(lines, note)
	}
	return lines
}

// Format renders the object type constraint. A documented attribute is set
// off from its neighbours by a blank line on each side so the comment reads
// as belonging to the attribute below it, the way every hand-authored module
// in the catalog laid its variables out; undocumented attributes stay in a
// compact run. Column alignment is not attempted here: the caller runs the
// whole file through the HCL formatter, which is the only way to match
// `tofu fmt` byte for byte.
func (o TFObject) Format(indent int) string {
	if len(o.Fields) == 0 {
		return "object({})"
	}

	indentStr := strings.Repeat("  ", indent)
	nextIndent := strings.Repeat("  ", indent+1)

	var lines []string
	prevDocumented := false
	for i, f := range o.Fields {
		typeExpr := f.Type.Format(indent + 1)
		if f.Optional {
			if f.Presence {
				// Tri-state: default to null, never to the zero value.
				typeExpr = fmt.Sprintf("optional(%s)", typeExpr)
			} else {
				typeExpr = wrapOptional(typeExpr, f.Type)
			}
		}
		comment := f.commentLines()
		documented := len(comment) > 0
		if i > 0 && (documented || prevDocumented) {
			lines = append(lines, "")
		}
		for _, c := range comment {
			if c == "" {
				lines = append(lines, nextIndent+"#")
				continue
			}
			lines = append(lines, nextIndent+"# "+c)
		}
		lines = append(lines, fmt.Sprintf("%s%s = %s", nextIndent, f.Name, typeExpr))
		prevDocumented = documented
	}

	return fmt.Sprintf("object({\n%s\n%s})", strings.Join(lines, "\n"), indentStr)
}

// wrapOptional wraps a formatted type expression in optional(...) with a default
// equal to the proto zero value for that type, so a tfvars that prunes the field
// (because it was unset/zero) reconstructs to the same zero rather than failing
// validation. Scalars/maps/lists carry an explicit zero default; nested objects
// and free-form `any` default to null (Terraform's implicit optional default),
// because there is no single literal zero for them -- consuming HCL null-guards
// these (the standard try()/!= null idiom in the modules).
func wrapOptional(typeExpr string, t TFType) string {
	if def, ok := zeroDefaultLiteral(t); ok {
		return fmt.Sprintf("optional(%s, %s)", typeExpr, def)
	}
	return fmt.Sprintf("optional(%s)", typeExpr)
}

// zeroDefaultLiteral returns the HCL literal for a type's proto zero value and
// true when one applies. It returns ("", false) for nested objects and `any`,
// signaling wrapOptional to omit the default (null).
func zeroDefaultLiteral(t TFType) (string, bool) {
	switch v := t.(type) {
	case TFPrimitive:
		switch string(v) {
		case "string":
			return `""`, true
		case "number":
			return "0", true
		case "bool":
			return "false", true
		default: // "any" and any future primitive: default to null
			return "", false
		}
	case TFList:
		return "[]", true
	case TFMap:
		return "{}", true
	case TFFreeFormMap:
		// Still a map semantically (just untyped values): default to an empty map.
		return "{}", true
	case TFFreeFormList:
		// Still a list semantically (just untyped elements): default to an
		// empty list.
		return "[]", true
	default: // TFObject and anything else: default to null
		return "", false
	}
}
