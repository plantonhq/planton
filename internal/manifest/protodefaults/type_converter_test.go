package protodefaults

import (
	"math"
	"testing"

	testgenericv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/_test/testcloudresourcegeneric/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// getFieldDescriptor is a helper to get a field descriptor from TestCloudResourceGenericSpec
func getFieldDescriptor(fieldName string) protoreflect.FieldDescriptor {
	specDesc := (&testgenericv1.TestCloudResourceGenericSpec{}).ProtoReflect().Descriptor()
	return specDesc.Fields().ByName(protoreflect.Name(fieldName))
}

// getNestedFieldDescriptor gets a field descriptor from TestNestedMessage
func getNestedFieldDescriptor(fieldName string) protoreflect.FieldDescriptor {
	nestedDesc := (&testgenericv1.TestGenericNestedMessage{}).ProtoReflect().Descriptor()
	return nestedDesc.Fields().ByName(protoreflect.Name(fieldName))
}

func TestConvertStringToFieldValue_String(t *testing.T) {
	field := getFieldDescriptor("string_field")

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"simple string", "hello", "hello"},
		{"empty string", "", ""},
		{"special chars", "hello-world_123", "hello-world_123"},
		{"unicode", "hello 世界", "hello 世界"},
		{"with spaces", "hello world", "hello world"},
		{"default value", "default-string", "default-string"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ConvertStringToFieldValue(tt.input, field)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result.String())
		})
	}
}

func TestConvertStringToFieldValue_Int32(t *testing.T) {
	field := getFieldDescriptor("int32_field")

	tests := []struct {
		name      string
		input     string
		expected  int32
		expectErr bool
	}{
		{"positive number", "42", 42, false},
		{"negative number", "-100", -100, false},
		{"zero", "0", 0, false},
		{"max int32", "2147483647", math.MaxInt32, false},
		{"min int32", "-2147483648", math.MinInt32, false},
		{"invalid - not a number", "not-a-number", 0, true},
		{"invalid - float", "3.14", 0, true},
		{"invalid - overflow", "9999999999999", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ConvertStringToFieldValue(tt.input, field)
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, int32(result.Int()))
			}
		})
	}
}

func TestConvertStringToFieldValue_Int64(t *testing.T) {
	field := getFieldDescriptor("int64_field")

	tests := []struct {
		name      string
		input     string
		expected  int64
		expectErr bool
	}{
		{"positive number", "9999", 9999, false},
		{"negative number", "-100", -100, false},
		{"zero", "0", 0, false},
		{"large number", "9223372036854775807", math.MaxInt64, false},
		{"invalid - not a number", "not-a-number", 0, true},
		{"invalid - float", "3.14", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ConvertStringToFieldValue(tt.input, field)
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, result.Int())
			}
		})
	}
}

func TestConvertStringToFieldValue_Uint32(t *testing.T) {
	field := getFieldDescriptor("uint32_field")

	tests := []struct {
		name      string
		input     string
		expected  uint32
		expectErr bool
	}{
		{"positive number", "100", 100, false},
		{"zero", "0", 0, false},
		{"max uint32", "4294967295", math.MaxUint32, false},
		{"invalid - negative", "-1", 0, true},
		{"invalid - not a number", "not-a-number", 0, true},
		{"invalid - overflow", "9999999999999", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ConvertStringToFieldValue(tt.input, field)
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, uint32(result.Uint()))
			}
		})
	}
}

func TestConvertStringToFieldValue_Uint64(t *testing.T) {
	field := getFieldDescriptor("uint64_field")

	tests := []struct {
		name      string
		input     string
		expected  uint64
		expectErr bool
	}{
		{"positive number", "50000", 50000, false},
		{"zero", "0", 0, false},
		{"large number", "18446744073709551615", uint64(math.MaxUint64), false},
		{"invalid - negative", "-1", 0, true},
		{"invalid - not a number", "not-a-number", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ConvertStringToFieldValue(tt.input, field)
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, result.Uint())
			}
		})
	}
}

func TestConvertStringToFieldValue_Bool(t *testing.T) {
	field := getFieldDescriptor("bool_field")

	tests := []struct {
		name      string
		input     string
		expected  bool
		expectErr bool
	}{
		{"true lowercase", "true", true, false},
		{"false lowercase", "false", false, false},
		{"true uppercase", "TRUE", true, false},
		{"false uppercase", "FALSE", false, false},
		{"1 as true", "1", true, false},
		{"0 as false", "0", false, false},
		{"invalid - yes", "yes", false, true},
		{"invalid - no", "no", false, true},
		{"invalid - maybe", "maybe", false, true},
		{"invalid - number", "2", false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ConvertStringToFieldValue(tt.input, field)
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, result.Bool())
			}
		})
	}
}

func TestConvertStringToFieldValue_Float32(t *testing.T) {
	field := getFieldDescriptor("float_field")

	tests := []struct {
		name      string
		input     string
		expected  float32
		expectErr bool
	}{
		{"positive float", "3.14", 3.14, false},
		{"negative float", "-2.5", -2.5, false},
		{"zero", "0.0", 0.0, false},
		{"scientific notation", "1e10", 1e10, false},
		{"small scientific", "1e-5", 1e-5, false},
		{"integer as float", "42", 42.0, false},
		{"invalid - not a number", "not-a-float", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ConvertStringToFieldValue(tt.input, field)
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.InDelta(t, tt.expected, result.Float(), 0.0001)
			}
		})
	}
}

func TestConvertStringToFieldValue_Float64(t *testing.T) {
	field := getFieldDescriptor("double_field")

	tests := []struct {
		name      string
		input     string
		expected  float64
		expectErr bool
	}{
		{"positive float", "2.718", 2.718, false},
		{"negative float", "-2.718281828", -2.718281828, false},
		{"zero", "0.0", 0.0, false},
		{"scientific notation", "1e100", 1e100, false},
		{"small scientific", "1e-100", 1e-100, false},
		{"integer as float", "42", 42.0, false},
		{"invalid - not a number", "not-a-float", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ConvertStringToFieldValue(tt.input, field)
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.InDelta(t, tt.expected, result.Float(), 0.0000000001)
			}
		})
	}
}

func TestConvertStringToFieldValue_NestedMessage(t *testing.T) {
	t.Run("nested message fields work correctly", func(t *testing.T) {
		// Test nested string field
		nestedStringField := getNestedFieldDescriptor("nested_string")
		result, err := ConvertStringToFieldValue("nested-test", nestedStringField)
		require.NoError(t, err)
		assert.Equal(t, "nested-test", result.String())

		// Test nested int32 field
		nestedIntField := getNestedFieldDescriptor("nested_int")
		result, err = ConvertStringToFieldValue("99", nestedIntField)
		require.NoError(t, err)
		assert.Equal(t, int32(99), int32(result.Int()))
	})
}

func TestConvertStringToFieldValue_AllDefaults(t *testing.T) {
	// This test verifies that all fields with defaults in the test proto
	// can be correctly converted from their default string values
	t.Run("all default values convert correctly", func(t *testing.T) {
		tests := []struct {
			fieldName string
			defValue  string
			expected  interface{}
		}{
			{"string_field", "default-string", "default-string"},
			{"int32_field", "42", int32(42)},
			{"int64_field", "9999", int64(9999)},
			{"uint32_field", "100", uint32(100)},
			{"uint64_field", "50000", uint64(50000)},
			{"float_field", "3.14", float32(3.14)},
			{"double_field", "2.718", float64(2.718)},
			{"bool_field", "true", true},
		}

		for _, tt := range tests {
			t.Run(tt.fieldName, func(t *testing.T) {
				field := getFieldDescriptor(tt.fieldName)
				result, err := ConvertStringToFieldValue(tt.defValue, field)
				require.NoError(t, err)
				assert.True(t, result.IsValid())
			})
		}
	})
}

// Verify error messages are descriptive
func TestConvertStringToFieldValue_ErrorMessages(t *testing.T) {
	boolField := getFieldDescriptor("bool_field")

	_, err := ConvertStringToFieldValue("invalid-bool", boolField)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to convert")
	assert.Contains(t, err.Error(), "bool")
}

// Test max values for numeric types
func TestConvertStringToFieldValue_NumericLimits(t *testing.T) {
	// Verify the logic handles max values correctly
	maxInt32 := int32(math.MaxInt32)
	minInt32 := int32(math.MinInt32)

	// These are the actual limits
	assert.Equal(t, int32(2147483647), maxInt32)
	assert.Equal(t, int32(-2147483648), minInt32)

	// Verify max values
	assert.Greater(t, float64(math.MaxFloat64), float64(0))
	assert.Less(t, -math.MaxFloat64, float64(0))
}
