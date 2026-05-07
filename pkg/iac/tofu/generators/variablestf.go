package generators

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"github.com/plantonhq/openmcf/internal/apidocs"
	"github.com/plantonhq/openmcf/pkg/strings/caseconverter"
	gendoc "github.com/pseudomuto/protoc-gen-doc"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

var defaultFieldDescriptions = map[string]string{
	"metadata": "Metadata for the resource, including name and labels",
	"spec":     "Specification for Deployment Component",
}

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
//
// Field descriptions are extracted from the proto API docs JSON.
func ProtoToVariablesTF(msg proto.Message) (string, error) {
	// API docs provide proto field descriptions for the generated variables.tf
	// comments. If unavailable (e.g., during unit tests or offline builds),
	// fall back to default descriptions rather than failing.
	apiDocsJSON, _ := apidocs.GetApiDocsJson()

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

		tfType, err := fieldToTFType(fd, md, rules, apiDocsJSON)
		if err != nil {
			return "", errors.Wrapf(err, "failed to convert field %q to terraform type", fieldName)
		}
		if tfType == nil {
			// Field was skipped by a type rule.
			continue
		}

		desc := resolveFieldDescription(apiDocsJSON, string(md.FullName()), fieldName, fd)
		typeStr := tfType.Format(1)

		fmt.Fprintf(&buf, "variable %q {\n  description = %q\n  type = %s\n}\n\n",
			caseconverter.ToSnakeCase(fieldName), desc, typeStr)
	}

	return strings.TrimSpace(buf.String()), nil
}

// fieldToTFType converts a proto field descriptor to a TFType, consulting type
// rules for skip/flatten decisions. Returns nil if the field should be skipped.
func fieldToTFType(fd protoreflect.FieldDescriptor, parentMD protoreflect.MessageDescriptor, rules map[string]TypeRule, apiDocsJSON *gendoc.Template) (TFType, error) {
	// Handle map fields first (before IsList, since maps are also "repeated" in proto).
	if fd.IsMap() {
		return mapFieldToTFType(fd, rules, apiDocsJSON)
	}

	// Handle repeated (list) fields.
	if fd.IsList() {
		elemType, err := scalarOrMsgToTFType(fd, parentMD, rules, apiDocsJSON)
		if err != nil {
			return nil, err
		}
		if elemType == nil {
			return nil, nil
		}
		return TFList{Elem: elemType}, nil
	}

	// Singular field.
	return scalarOrMsgToTFType(fd, parentMD, rules, apiDocsJSON)
}

// mapFieldToTFType converts a proto map<K, V> field to TFMap. The key type is
// always string in OpenMCF protos. The value type is determined by consulting
// type rules (a map<string, StringValueOrRef> becomes map(string)).
func mapFieldToTFType(fd protoreflect.FieldDescriptor, rules map[string]TypeRule, apiDocsJSON *gendoc.Template) (TFType, error) {
	valDesc := fd.MapValue()

	valType, err := mapValueToTFType(valDesc, rules, apiDocsJSON)
	if err != nil {
		return nil, err
	}
	if valType == nil {
		return nil, nil
	}

	return TFMap{Value: valType}, nil
}

// mapValueToTFType resolves the TFType for a map value descriptor.
func mapValueToTFType(valDesc protoreflect.FieldDescriptor, rules map[string]TypeRule, apiDocsJSON *gendoc.Template) (TFType, error) {
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
			return TFPrimitive("string"), nil
		}
		return msgDescToTFObject(valDesc.Message(), valDesc, rules, apiDocsJSON)
	default:
		return TFPrimitive("string"), nil
	}
}

// scalarOrMsgToTFType converts a single (non-map, non-list-wrapper) field to
// a TFType. For message-kind fields, consults type rules.
func scalarOrMsgToTFType(fd protoreflect.FieldDescriptor, parentMD protoreflect.MessageDescriptor, rules map[string]TypeRule, apiDocsJSON *gendoc.Template) (TFType, error) {
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
			return TFPrimitive("string"), nil
		}

		return msgDescToTFObject(fd.Message(), fd, rules, apiDocsJSON)
	default:
		return nil, fmt.Errorf("unsupported field kind: %v", fd.Kind())
	}
}

// msgDescToTFObject recursively converts a proto message descriptor to a
// TFObject, respecting type rules and skipping the "version" field inside
// metadata messages.
func msgDescToTFObject(md protoreflect.MessageDescriptor, fd protoreflect.FieldDescriptor, rules map[string]TypeRule, apiDocsJSON *gendoc.Template) (TFType, error) {
	fields := md.Fields()
	obj := TFObject{}

	shouldSkipVersion := strings.HasSuffix(strings.ToLower(string(md.Name())), "metadata")
	parentFullName := string(md.FullName())

	for i := 0; i < fields.Len(); i++ {
		f := fields.Get(i)
		fieldName := string(f.Name())

		if shouldSkipVersion && fieldName == "version" {
			continue
		}

		valType, err := fieldToTFType(f, md, rules, apiDocsJSON)
		if err != nil {
			return nil, err
		}
		if valType == nil {
			continue
		}

		desc := resolveFieldDescription(apiDocsJSON, parentFullName, fieldName, f)
		obj.Fields = append(obj.Fields, TFField{
			Name:        caseconverter.ToSnakeCase(fieldName),
			Description: desc,
			Type:        valType,
		})
	}

	return obj, nil
}

// resolveFieldDescription finds the best description for a proto field,
// checking API docs first, then falling back to message-level docs, then to
// hardcoded defaults.
func resolveFieldDescription(apiDocsJSON *gendoc.Template, parentFullName, fieldName string, fd protoreflect.FieldDescriptor) string {
	desc := findFieldDesc(apiDocsJSON, parentFullName, fieldName)
	if desc == "" && fd.Kind() == protoreflect.MessageKind && !fd.IsMap() {
		desc = findMsgDesc(apiDocsJSON, string(fd.Message().FullName()))
	}
	if desc == "" {
		desc = defaultFieldDescriptions[fieldName]
	}
	if desc == "" {
		desc = fmt.Sprintf("Description for %s", fieldName)
	}
	return desc
}

func findMsgDesc(apiDocsJSON *gendoc.Template, fullName string) string {
	if apiDocsJSON == nil {
		return ""
	}
	for _, f := range apiDocsJSON.Files {
		for _, m := range f.Messages {
			if m.FullName == fullName {
				return strings.TrimSpace(m.Description)
			}
		}
	}
	return ""
}

func findFieldDesc(apiDocsJSON *gendoc.Template, msgFullName, fieldName string) string {
	if apiDocsJSON == nil {
		return ""
	}
	for _, f := range apiDocsJSON.Files {
		for _, m := range f.Messages {
			if m.FullName == msgFullName {
				for _, fld := range m.Fields {
					if fld.Name == fieldName {
						return strings.TrimSpace(fld.Description)
					}
				}
			}
		}
	}
	return ""
}

// isWellKnownJSONType returns true for protobuf well-known types representing
// JSON, which should be mapped to string in Terraform.
func isWellKnownJSONType(fullName string) bool {
	switch fullName {
	case "google.protobuf.Struct", "google.protobuf.Value", "google.protobuf.ListValue":
		return true
	default:
		return false
	}
}
