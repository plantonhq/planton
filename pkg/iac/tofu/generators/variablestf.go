package generators

import (
	"bytes"
	"fmt"
	"strings"

	"buf.build/gen/go/bufbuild/protovalidate/protocolbuffers/go/buf/validate"
	"github.com/pkg/errors"
	"github.com/plantonhq/planton/pkg/strings/caseconverter"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// metadataVariableDescription is the fixed description of the shared resource
// envelope variable.
const metadataVariableDescription = "Cloud resource metadata"

// cloudResourceMetadataFullName is the shared resource-envelope message. Its
// Terraform shape is an invariant across every kind, so it is emitted from one
// canonical block (canonicalMetadataObject) rather than derived per kind -- the
// proto carries no field constraints, and the envelope deliberately exposes only
// name/id/org/env/labels/annotations/tags to modules (slug/group/relationships
// are orchestrator concerns dropped during object conversion).
const cloudResourceMetadataFullName = "dev.planton.shared.CloudResourceMetadata"

// topLevelSkipFieldNames lists proto field names to skip at the top level of
// the resource message. These are proto envelope fields that have no meaning
// in Terraform.
var topLevelSkipFieldNames = map[string]bool{
	"api_version": true,
	"kind":        true,
	"status":      true,
}

// ProtoToVariablesTF generates Terraform variable definitions from a proto
// message using proto reflection. It consults the shared TypeRule registry to:
//   - Skip orchestrator-only fields (ValueFromRef)
//   - Flatten wrapper types to primitives (StringValueOrRef -> string)
//   - Handle proto maps as map(valueType) instead of misrepresenting them as objects
//   - Mark every non-required attribute optional() with its proto zero default
//
// The output is fully deterministic and offline: it depends only on the compiled
// proto descriptor (types + buf.validate constraints), never on a network call
// or external docs source. Determinism is what lets the committed variables.tf be
// guarded against drift by regenerating and comparing.
func ProtoToVariablesTF(msg proto.Message) (string, error) {
	md := msg.ProtoReflect().Descriptor()
	rules := DefaultRules()

	var buf bytes.Buffer
	fields := md.Fields()

	// Path-scoped cycle guard for the descriptor walk (see msgDescToTFObject).
	visited := map[protoreflect.FullName]bool{md.FullName(): true}

	for i := 0; i < fields.Len(); i++ {
		fd := fields.Get(i)
		fieldName := string(fd.Name())

		if topLevelSkipFieldNames[fieldName] {
			continue
		}

		var tfType TFType
		var desc string
		if isCloudResourceMetadataField(fd) {
			// The resource envelope is uniform across kinds: emit the canonical
			// block instead of deriving from the (constraint-free) proto, which
			// would wrongly mark every attribute required and leak orchestrator
			// fields.
			tfType = canonicalMetadataObject()
			desc = metadataVariableDescription
		} else {
			t, err := fieldToTFType(fd, md, rules, visited)
			if err != nil {
				return "", errors.Wrapf(err, "failed to convert field %q to terraform type", fieldName)
			}
			if t == nil {
				// Field was skipped by a type rule.
				continue
			}
			tfType = t
			desc = variableDescription(md, fieldName)
		}
		typeStr := tfType.Format(1)

		fmt.Fprintf(&buf, "variable %q {\n  description = %q\n  type = %s\n}\n\n",
			caseconverter.ToSnakeCase(fieldName), desc, typeStr)
	}

	return strings.TrimSpace(buf.String()), nil
}

// variableDescription returns a deterministic one-line description for a
// top-level variable. The spec variable reads "<Kind> specification" (matching
// the established module style); any other top-level field falls back to a
// generic, stable phrasing.
func variableDescription(resourceMD protoreflect.MessageDescriptor, fieldName string) string {
	if fieldName == "spec" {
		return fmt.Sprintf("%s specification", resourceMD.Name())
	}
	return fmt.Sprintf("%s %s", resourceMD.Name(), fieldName)
}

// fieldToTFType converts a proto field descriptor to a TFType, consulting type
// rules for skip/flatten decisions. Returns nil if the field should be skipped.
// The visited set is the path-scoped cycle guard threaded through the whole
// descriptor walk (see msgDescToTFObject).
func fieldToTFType(fd protoreflect.FieldDescriptor, parentMD protoreflect.MessageDescriptor, rules map[string]TypeRule, visited map[protoreflect.FullName]bool) (TFType, error) {
	// A manifest-only word gets no variable: the tfvars converter never sends
	// it, so a declared attribute would be dead on every module. Callers read
	// a nil type as "skipped", the same way a Skip type rule is read.
	if isManifestOnlyField(fd) {
		return nil, nil
	}

	// Handle map fields first (before IsList, since maps are also "repeated" in proto).
	if fd.IsMap() {
		return mapFieldToTFType(fd, rules, visited)
	}

	// Handle repeated (list) fields.
	if fd.IsList() {
		elemType, err := scalarOrMsgToTFType(fd, parentMD, rules, visited)
		if err != nil {
			return nil, err
		}
		if elemType == nil {
			return nil, nil
		}
		// Free-form content (a JSON well-known type, or a recursive message
		// collapsed to `any`) cannot appear ANYWHERE inside a list element
		// type: Terraform unifies every element of a list(...) to one
		// concrete type, and an embedded `any` placeholder must resolve to
		// the same concrete type across all elements -- which arbitrary or
		// recursive content cannot guarantee ("element types must all match
		// for conversion to list"). The whole attribute becomes `any`.
		if containsFreeForm(elemType) {
			return TFFreeFormList{}, nil
		}
		return TFList{Elem: elemType}, nil
	}

	// Singular field.
	return scalarOrMsgToTFType(fd, parentMD, rules, visited)
}

// mapFieldToTFType converts a proto map<K, V> field to TFMap. The key type is
// always string in Planton protos. The value type is determined by consulting
// type rules (a map<string, StringValueOrRef> becomes map(string)).
//
// A map whose value is a free-form JSON well-known type (map<string,
// google.protobuf.Struct/Value/ListValue>) is the exception: its entries are
// independently-shaped, so it cannot be a homogeneous map(any). It becomes a
// TFFreeFormMap (rendered `any`) -- see that type for the rationale.
func mapFieldToTFType(fd protoreflect.FieldDescriptor, rules map[string]TypeRule, visited map[protoreflect.FullName]bool) (TFType, error) {
	valDesc := fd.MapValue()

	if valDesc.Kind() == protoreflect.MessageKind &&
		isWellKnownJSONType(string(valDesc.Message().FullName())) {
		return TFFreeFormMap{}, nil
	}

	valType, err := mapValueToTFType(valDesc, rules, visited)
	if err != nil {
		return nil, err
	}
	if valType == nil {
		return nil, nil
	}

	// Same unification constraint as lists: free-form content anywhere in
	// the value type cannot live inside map(...) -- see containsFreeForm.
	if containsFreeForm(valType) {
		return TFFreeFormMap{}, nil
	}

	return TFMap{Value: valType}, nil
}

// mapValueToTFType resolves the TFType for a map value descriptor.
func mapValueToTFType(valDesc protoreflect.FieldDescriptor, rules map[string]TypeRule, visited map[protoreflect.FullName]bool) (TFType, error) {
	switch valDesc.Kind() {
	case protoreflect.StringKind:
		return TFPrimitive("string"), nil
	case protoreflect.BoolKind:
		return TFPrimitive("bool"), nil
	case protoreflect.Int32Kind, protoreflect.Int64Kind,
		protoreflect.Uint32Kind, protoreflect.Uint64Kind,
		protoreflect.Sint32Kind, protoreflect.Sint64Kind,
		protoreflect.Fixed32Kind, protoreflect.Fixed64Kind,
		protoreflect.Sfixed32Kind, protoreflect.Sfixed64Kind,
		protoreflect.FloatKind, protoreflect.DoubleKind:
		return TFPrimitive("number"), nil
	case protoreflect.EnumKind:
		return TFPrimitive("string"), nil
	case protoreflect.MessageKind:
		fullName := string(valDesc.Message().FullName())
		if rule, ok := rules[fullName]; ok {
			if rule.Skip {
				return nil, nil
			}
			if rule.FlattenTo != "" {
				return TFPrimitive(rule.FlattenTo), nil
			}
		}
		if isWellKnownJSONType(fullName) {
			return TFPrimitive("any"), nil
		}
		if visited[protoreflect.FullName(fullName)] {
			// Descriptor cycle: see the rationale in scalarOrMsgToTFType.
			return TFPrimitive("any"), nil
		}
		return msgDescToTFObject(valDesc.Message(), rules, visited)
	default:
		return TFPrimitive("string"), nil
	}
}

// scalarOrMsgToTFType converts a single (non-map, non-list-wrapper) field to
// a TFType. For message-kind fields, consults type rules.
func scalarOrMsgToTFType(fd protoreflect.FieldDescriptor, parentMD protoreflect.MessageDescriptor, rules map[string]TypeRule, visited map[protoreflect.FullName]bool) (TFType, error) {
	switch fd.Kind() {
	case protoreflect.StringKind:
		return TFPrimitive("string"), nil
	case protoreflect.BoolKind:
		return TFPrimitive("bool"), nil
	case protoreflect.Int32Kind, protoreflect.Int64Kind,
		protoreflect.Uint32Kind, protoreflect.Uint64Kind,
		protoreflect.Sint32Kind, protoreflect.Sint64Kind,
		protoreflect.Fixed32Kind, protoreflect.Fixed64Kind,
		protoreflect.Sfixed32Kind, protoreflect.Sfixed64Kind,
		protoreflect.FloatKind, protoreflect.DoubleKind:
		return TFPrimitive("number"), nil
	case protoreflect.BytesKind:
		return TFPrimitive("string"), nil
	case protoreflect.EnumKind:
		return TFPrimitive("string"), nil
	case protoreflect.MessageKind:
		fullName := string(fd.Message().FullName())

		if rule, ok := rules[fullName]; ok {
			if rule.Skip {
				return nil, nil
			}
			if rule.FlattenTo != "" {
				return TFPrimitive(rule.FlattenTo), nil
			}
		}

		if isWellKnownJSONType(fullName) {
			return TFPrimitive("any"), nil
		}

		if visited[protoreflect.FullName(fullName)] {
			// Descriptor cycle: the message (directly or transitively) contains
			// itself -- a legitimate proto shape for naturally recursive
			// structures (e.g. boolean statement trees). Terraform's type
			// language cannot express a recursive object constraint, so the
			// recursive subtree collapses to `any` -- the same treatment as
			// google.protobuf.Struct. The tfvars renderer is value-driven and
			// unaffected: values inside the subtree still emit with proto
			// field names and flattened wrapper types, so modules read the
			// exact shapes a fully typed contract would carry.
			return TFPrimitive("any"), nil
		}

		return msgDescToTFObject(fd.Message(), rules, visited)
	default:
		return nil, fmt.Errorf("unsupported field kind: %v", fd.Kind())
	}
}

// msgDescToTFObject recursively converts a proto message descriptor to a
// TFObject, respecting type rules and skipping the "version" field inside
// metadata messages. Each attribute is marked optional unless the proto field is
// required (see isRequiredField).
//
// The visited set is a path-scoped cycle guard (mark on entry, unmark on exit --
// the same idiom the refcheck and secret-coverage descriptor walks use): a
// message may legitimately appear at many sibling paths, but re-entering a type
// already on the CURRENT path means the descriptor graph is recursive and the
// walk would never terminate. The message-kind converters break such cycles by
// collapsing the recursive subtree to `any`.
func msgDescToTFObject(md protoreflect.MessageDescriptor, rules map[string]TypeRule, visited map[protoreflect.FullName]bool) (TFType, error) {
	visited[md.FullName()] = true
	defer delete(visited, md.FullName())

	fields := md.Fields()
	obj := TFObject{}

	shouldSkipVersion := strings.HasSuffix(strings.ToLower(string(md.Name())), "metadata")

	for i := 0; i < fields.Len(); i++ {
		f := fields.Get(i)
		fieldName := string(f.Name())

		if shouldSkipVersion && fieldName == "version" {
			continue
		}

		valType, err := fieldToTFType(f, md, rules, visited)
		if err != nil {
			return nil, err
		}
		if valType == nil {
			continue
		}

		obj.Fields = append(obj.Fields, TFField{
			Name:     caseconverter.ToSnakeCase(fieldName),
			Type:     valType,
			Optional: !isRequiredField(f),
			Presence: f.HasOptionalKeyword(),
		})
	}

	return obj, nil
}

// containsFreeForm reports whether a type carries free-form (`any`-typed)
// content anywhere in its structure. Lists and maps must not embed such
// content in their element/value types (Terraform cannot unify `any` across
// elements), so their converters collapse to the free-form container types
// when this returns true.
func containsFreeForm(t TFType) bool {
	switch v := t.(type) {
	case TFPrimitive:
		return string(v) == "any"
	case TFFreeFormMap, TFFreeFormList:
		return true
	case TFList:
		return containsFreeForm(v.Elem)
	case TFMap:
		return containsFreeForm(v.Value)
	case TFObject:
		for _, f := range v.Fields {
			if containsFreeForm(f.Type) {
				return true
			}
		}
		return false
	default:
		return false
	}
}

// isRequiredField reports whether a proto field must always be present in the
// rendered tfvars, and therefore must stay a bare (non-optional) attribute. The
// source of truth is buf.validate: a field is required if it is explicitly
// (buf.validate.field).required, or if it carries a presence-implying constraint
// (string min_len >= 1, repeated min_items >= 1). Everything else is optional,
// because the renderer prunes unset/zero fields and a bare attribute would then
// fail object validation.
func isRequiredField(fd protoreflect.FieldDescriptor) bool {
	opts := fd.Options()
	if opts == nil {
		return false
	}
	if !proto.HasExtension(opts, validate.E_Field) {
		return false
	}
	rules, ok := proto.GetExtension(opts, validate.E_Field).(*validate.FieldRules)
	if !ok || rules == nil {
		return false
	}
	if rules.GetRequired() {
		return true
	}
	if s := rules.GetString(); s != nil && s.GetMinLen() >= 1 {
		return true
	}
	if r := rules.GetRepeated(); r != nil && r.GetMinItems() >= 1 {
		return true
	}
	return false
}

// isCloudResourceMetadataField reports whether a field is the shared resource
// metadata envelope, which is emitted from the canonical block.
func isCloudResourceMetadataField(fd protoreflect.FieldDescriptor) bool {
	return fd.Kind() == protoreflect.MessageKind &&
		!fd.IsMap() && !fd.IsList() &&
		string(fd.Message().FullName()) == cloudResourceMetadataFullName
}

// canonicalMetadataObject returns the fixed Terraform shape of the shared
// resource metadata envelope: name is always present (required); the rest are
// optional with their zero-value defaults so a pruned tfvars validates. This
// mirrors the contract every module relies on and is identical across kinds.
func canonicalMetadataObject() TFObject {
	return TFObject{Fields: []TFField{
		{Name: "name", Type: TFPrimitive("string")},
		{Name: "id", Type: TFPrimitive("string"), Optional: true},
		{Name: "org", Type: TFPrimitive("string"), Optional: true},
		{Name: "env", Type: TFPrimitive("string"), Optional: true},
		{Name: "labels", Type: TFMap{Value: TFPrimitive("string")}, Optional: true},
		{Name: "annotations", Type: TFMap{Value: TFPrimitive("string")}, Optional: true},
		{Name: "tags", Type: TFList{Elem: TFPrimitive("string")}, Optional: true},
	}}
}

// isWellKnownJSONType returns true for protobuf well-known types representing
// free-form JSON, which are mapped to the Terraform `any` type (the nested
// JSON value is passed through verbatim).
func isWellKnownJSONType(fullName string) bool {
	switch fullName {
	case "google.protobuf.Struct", "google.protobuf.Value", "google.protobuf.ListValue":
		return true
	default:
		return false
	}
}
