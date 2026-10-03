package fieldsextractor

import (
	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// IacInputTargetField Field names in the Protobuf message
const IacInputTargetField = "target"
const IacInputTargetSpecField = "spec"

// ExtractApiResourceSpecField extracts the spec field from iac-input using protobuf reflection and returns it.
func ExtractApiResourceSpecField(iacInput proto.Message) (*protoreflect.Message, error) {
	// Check if iacInput is nil
	if iacInput == nil {
		return nil, errors.New("iac-input is nil")
	}

	// Get the protobuf message descriptor
	iacInputProtoReflect := iacInput.ProtoReflect()

	// Retrieve the target field by name
	targetField := iacInputProtoReflect.Descriptor().Fields().ByName(IacInputTargetField)
	if targetField == nil {
		return nil, errors.Errorf("Field %s not found in iac-input", IacInputTargetField)
	}

	// Get the value of the target field and check if it is nil
	targetValue := iacInputProtoReflect.Get(targetField).Message()
	if targetValue.IsValid() == false {
		return nil, errors.Errorf("Field %s is nil in iac-input", IacInputTargetField)
	}

	// Get the protobuf message descriptor for target
	targetProtoReflect := targetValue.Interface().ProtoReflect()

	// Retrieve the spec field by name
	targetSpecField := targetProtoReflect.Descriptor().Fields().ByName(IacInputTargetSpecField)
	if targetSpecField == nil {
		return nil, errors.Errorf("Field %s not found in target", IacInputTargetSpecField)
	}

	// Get the value of the spec field and check if it is nil
	targetSpecValue := targetProtoReflect.Get(targetSpecField).Message()
	if targetSpecValue.IsValid() == false {
		return nil, errors.Errorf("Field %s is nil in target", IacInputTargetSpecField)
	}

	// Return the spec message
	return &targetSpecValue, nil
}
