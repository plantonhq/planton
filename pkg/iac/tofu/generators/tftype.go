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
type TFField struct {
	Name     string
	Type     TFType
	Optional bool
}

func (o TFObject) Format(indent int) string {
	if len(o.Fields) == 0 {
		return "object({})"
	}

	indentStr := strings.Repeat("  ", indent)
	nextIndent := strings.Repeat("  ", indent+1)

	var lines []string
	for _, f := range o.Fields {
		typeExpr := f.Type.Format(indent + 1)
		if f.Optional {
			typeExpr = wrapOptional(typeExpr, f.Type)
		}
		lines = append(lines, fmt.Sprintf("%s%s = %s", nextIndent, f.Name, typeExpr))
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
	default: // TFObject and anything else: default to null
		return "", false
	}
}
