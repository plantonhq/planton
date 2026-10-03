package catalogkindreflect

import (
	"os"

	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"gopkg.in/yaml.v3"
)

const (
	KindProtoFieldName = "kind"
)

// ExtractKindFromTargetManifest reads a YAML file from the given path and returns the value of the 'kind' key.
func ExtractKindFromTargetManifest(targetManifest string) (string, error) {
	// Check if the file exists
	if _, err := os.Stat(targetManifest); os.IsNotExist(err) {
		return "", errors.Wrapf(err, "file not found: %s", targetManifest)
	}

	// Read the YAML file
	fileContent, err := os.ReadFile(targetManifest)
	if err != nil {
		return "", errors.Wrapf(err, "failed to read file: %s", targetManifest)
	}

	// Parse the YAML content
	var yamlData map[string]interface{}
	if err := yaml.Unmarshal(fileContent, &yamlData); err != nil {
		return "", errors.Wrapf(err, "failed to unmarshal YAML content from file: %s", targetManifest)
	}

	// Extract the 'kind' key
	kind, ok := yamlData["kind"]
	if !ok {
		return "", errors.Errorf("key 'kind' not found in YAML file: %s", targetManifest)
	}

	// Ensure the 'kind' key is a string
	kindStr, ok := kind.(string)
	if !ok {
		return "", errors.Errorf("value of 'kind' key is not a string in YAML file: %s", targetManifest)
	}

	return kindStr, nil
}

func ExtractKindFromProto(manifestObject proto.Message) (string, error) {
	// Get the protobuf message descriptor for target
	manifestProtoReflect := manifestObject.ProtoReflect()

	// Retrieve the field by name
	field := manifestProtoReflect.Descriptor().Fields().ByName(KindProtoFieldName)
	if field == nil {
		return "", errors.Errorf("Field %s not found in manifest", KindProtoFieldName)
	}

	// Check if the field is of type string
	if field.Kind() != protoreflect.StringKind {
		return "", errors.Errorf("Field %s is not of type string", KindProtoFieldName)
	}

	// Get the value of the field and check if it is nil
	fieldValue := manifestProtoReflect.Get(field)
	if fieldValue.IsValid() == false {
		return "", errors.Errorf("field %s is nil in manifest", KindProtoFieldName)
	}

	// Return the string value of the field
	return fieldValue.String(), nil
}
