package outputs

import (
	"fmt"
	"strconv"

	"google.golang.org/protobuf/reflect/protoreflect"
)

// convertScalar converts a string value to a protoreflect.Value matching the
// target field descriptor's type. This is the Go equivalent of Java's
// StackOutputsMapToProtoLoader.convertToFieldType().
//
// Supported kinds: string, bool, int32, sint32, sfixed32, int64, sint64,
// sfixed64, uint32, fixed32, uint64, fixed64, float, double, enum.
//
// For enum fields, the value is looked up by name first, then by numeric value.
// Returns an error if the string cannot be parsed into the target type.
func convertScalar(value string, fd protoreflect.FieldDescriptor) (protoreflect.Value, error) {
	switch fd.Kind() {
	case protoreflect.StringKind:
		return protoreflect.ValueOfString(value), nil

	case protoreflect.BoolKind:
		b, err := strconv.ParseBool(value)
		if err != nil {
			return protoreflect.Value{}, fmt.Errorf("field %s: cannot parse %q as bool: %w",
				fd.FullName(), value, err)
		}
		return protoreflect.ValueOfBool(b), nil

	case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Sfixed32Kind:
		n, err := strconv.ParseInt(value, 10, 32)
		if err != nil {
			return protoreflect.Value{}, fmt.Errorf("field %s: cannot parse %q as int32: %w",
				fd.FullName(), value, err)
		}
		return protoreflect.ValueOfInt32(int32(n)), nil

	case protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Sfixed64Kind:
		n, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return protoreflect.Value{}, fmt.Errorf("field %s: cannot parse %q as int64: %w",
				fd.FullName(), value, err)
		}
		return protoreflect.ValueOfInt64(n), nil

	case protoreflect.Uint32Kind, protoreflect.Fixed32Kind:
		n, err := strconv.ParseUint(value, 10, 32)
		if err != nil {
			return protoreflect.Value{}, fmt.Errorf("field %s: cannot parse %q as uint32: %w",
				fd.FullName(), value, err)
		}
		return protoreflect.ValueOfUint32(uint32(n)), nil

	case protoreflect.Uint64Kind, protoreflect.Fixed64Kind:
		n, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return protoreflect.Value{}, fmt.Errorf("field %s: cannot parse %q as uint64: %w",
				fd.FullName(), value, err)
		}
		return protoreflect.ValueOfUint64(n), nil

	case protoreflect.FloatKind:
		f, err := strconv.ParseFloat(value, 32)
		if err != nil {
			return protoreflect.Value{}, fmt.Errorf("field %s: cannot parse %q as float: %w",
				fd.FullName(), value, err)
		}
		return protoreflect.ValueOfFloat32(float32(f)), nil

	case protoreflect.DoubleKind:
		f, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return protoreflect.Value{}, fmt.Errorf("field %s: cannot parse %q as double: %w",
				fd.FullName(), value, err)
		}
		return protoreflect.ValueOfFloat64(f), nil

	case protoreflect.EnumKind:
		enumDesc := fd.Enum()
		// Try lookup by name first, then by numeric value.
		if ev := enumDesc.Values().ByName(protoreflect.Name(value)); ev != nil {
			return protoreflect.ValueOfEnum(ev.Number()), nil
		}
		n, err := strconv.ParseInt(value, 10, 32)
		if err != nil {
			return protoreflect.Value{}, fmt.Errorf("field %s: unknown enum value %q and not a valid number",
				fd.FullName(), value)
		}
		return protoreflect.ValueOfEnum(protoreflect.EnumNumber(n)), nil

	case protoreflect.BytesKind:
		return protoreflect.ValueOfBytes([]byte(value)), nil

	default:
		return protoreflect.Value{}, fmt.Errorf("field %s: unsupported scalar kind %s",
			fd.FullName(), fd.Kind())
	}
}
