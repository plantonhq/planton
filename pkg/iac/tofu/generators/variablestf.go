package generators

import (
	"bytes"
	"fmt"
	"strings"

	"buf.build/gen/go/bufbuild/protovalidate/protocolbuffers/go/buf/validate"
	"github.com/pkg/errors"
	"github.com/plantonhq/openmcf/pkg/strings/caseconverter"
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
const cloudResourceMetadataFullName = "org.openmcf.shared.CloudResourceMetadata"

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
//   - Skip orchestrator-only fields (KubernetesClusterSelector, ValueFromRef)
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
			t, err := fieldToTFType(fd, md, rules)
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
func fieldToTFType(fd protoreflect.FieldDescriptor, parentMD protoreflect.MessageDescriptor, rules map[string]TypeRule) (TFType, error) {
	// Handle map fields first (before IsList, since maps are also "repeated" in proto).
	if fd.IsMap() {
		return mapFieldToTFType(fd, rules)
	}

	// Handle repeated (list) fields.
	if fd.IsList() {
		elemType, err := scalarOrMsgToTFType(fd, parentMD, rules)
		if err != nil {
			return nil, err
		}
		if elemType == nil {
			return nil, nil
		}
		return TFList{Elem: elemType}, nil
	}

	// Singular field.
	return scalarOrMsgToTFType(fd, parentMD, rules)
}

// mapFieldToTFType converts a proto map<K, V> field to TFMap. The key type is
// always string in OpenMCF protos. The value type is determined by consulting
// type rules (a map<string, StringValueOrRef> becomes map(string)).
func mapFieldToTFType(fd protoreflect.FieldDescriptor, rules map[string]TypeRule) (TFType, error) {
	valDesc := fd.MapValue()

	valType, err := mapValueToTFType(valDesc, rules)
	if err != nil {
		return nil, err
	}
	if valType == nil {
		return nil, nil
	}

	return TFMap{Value: valType}, nil
}

// mapValueToTFType resolves the TFType for a map value descriptor.
func mapValueToTFType(valDesc protoreflect.FieldDescriptor, rules map[string]TypeRule) (TFType, error) {
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
		return msgDescToTFObject(valDesc.Message(), rules)
	default:
		return TFPrimitive("string"), nil
	}
}

// scalarOrMsgToTFType converts a single (non-map, non-list-wrapper) field to
// a TFType. For message-kind fields, consults type rules.
func scalarOrMsgToTFType(fd protoreflect.FieldDescriptor, parentMD protoreflect.MessageDescriptor, rules map[string]TypeRule) (TFType, error) {
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

		return msgDescToTFObject(fd.Message(), rules)
	default:
		return nil, fmt.Errorf("unsupported field kind: %v", fd.Kind())
	}
}

// msgDescToTFObject recursively converts a proto message descriptor to a
// TFObject, respecting type rules and skipping the "version" field inside
// metadata messages. Each attribute is marked optional unless the proto field is
// required (see isRequiredField).
func msgDescToTFObject(md protoreflect.MessageDescriptor, rules map[string]TypeRule) (TFType, error) {
	fields := md.Fields()
	obj := TFObject{}

	shouldSkipVersion := strings.HasSuffix(strings.ToLower(string(md.Name())), "metadata")

	for i := 0; i < fields.Len(); i++ {
		f := fields.Get(i)
		fieldName := string(f.Name())

		if shouldSkipVersion && fieldName == "version" {
			continue
		}

		valType, err := fieldToTFType(f, md, rules)
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
		})
	}

	return obj, nil
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
