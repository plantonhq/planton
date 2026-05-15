package protodefaults

import (
	"testing"

	testgenericv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/_test/testcloudresourcegeneric/v1"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestApplyDefaults_AllScalarTypes(t *testing.T) {
	t.Run("applies defaults to unset fields", func(t *testing.T) {
		// Create a message with minimal required fields, leaving fields with defaults unset
		msg := &testgenericv1.TestCloudResourceGeneric{
			ApiVersion: "_test.openmcf.org/v1",
			Kind:       "TestCloudResourceGeneric",
			Metadata: &shared.CloudResourceMetadata{
				Name: "test-resource",
			},
			Spec: &testgenericv1.TestCloudResourceGenericSpec{
				// All fields with defaults are left unset (nil pointers)
			},
		}

		// Apply defaults
		err := ApplyDefaults(msg)
		require.NoError(t, err)

		// Verify all defaults were applied (pointers are non-nil)
		require.NotNil(t, msg.Spec.StringField)
		assert.Equal(t, "default-string", *msg.Spec.StringField)

		require.NotNil(t, msg.Spec.Int32Field)
		assert.Equal(t, int32(42), *msg.Spec.Int32Field)

		require.NotNil(t, msg.Spec.Int64Field)
		assert.Equal(t, int64(9999), *msg.Spec.Int64Field)

		require.NotNil(t, msg.Spec.Uint32Field)
		assert.Equal(t, uint32(100), *msg.Spec.Uint32Field)

		require.NotNil(t, msg.Spec.Uint64Field)
		assert.Equal(t, uint64(50000), *msg.Spec.Uint64Field)

		require.NotNil(t, msg.Spec.FloatField)
		assert.InDelta(t, float32(3.14), *msg.Spec.FloatField, 0.001)

		require.NotNil(t, msg.Spec.DoubleField)
		assert.InDelta(t, 2.718, *msg.Spec.DoubleField, 0.0001)

		require.NotNil(t, msg.Spec.BoolField)
		assert.True(t, *msg.Spec.BoolField)

		// Field without default should remain unset
		assert.Equal(t, "", msg.Spec.StringNoDefault)
	})

	t.Run("preserves existing values when field is already set", func(t *testing.T) {
		// Create a message with custom values (using pointers)
		msg := &testgenericv1.TestCloudResourceGeneric{
			ApiVersion: "_test.openmcf.org/v1",
			Kind:       "TestCloudResourceGeneric",
			Metadata: &shared.CloudResourceMetadata{
				Name: "test-resource",
			},
			Spec: &testgenericv1.TestCloudResourceGenericSpec{
				StringField: proto.String("custom-string"),
				Int32Field:  proto.Int32(999),
				FloatField:  proto.Float32(1.23),
				DoubleField: proto.Float64(4.56),
				BoolField:   proto.Bool(false), // Can now explicitly test false!
			},
		}

		// Apply defaults
		err := ApplyDefaults(msg)
		require.NoError(t, err)

		// Verify existing values were preserved (even zero values!)
		require.NotNil(t, msg.Spec.StringField)
		assert.Equal(t, "custom-string", *msg.Spec.StringField)

		require.NotNil(t, msg.Spec.Int32Field)
		assert.Equal(t, int32(999), *msg.Spec.Int32Field)

		require.NotNil(t, msg.Spec.FloatField)
		assert.InDelta(t, float32(1.23), *msg.Spec.FloatField, 0.001)

		require.NotNil(t, msg.Spec.DoubleField)
		assert.InDelta(t, 4.56, *msg.Spec.DoubleField, 0.0001)

		// CRITICAL: Bool field explicitly set to false should be preserved, not defaulted!
		require.NotNil(t, msg.Spec.BoolField)
		assert.False(t, *msg.Spec.BoolField) // Should stay false, not get default true
	})

	t.Run("handles partial values - some set, some unset", func(t *testing.T) {
		msg := &testgenericv1.TestCloudResourceGeneric{
			ApiVersion: "_test.openmcf.org/v1",
			Kind:       "TestCloudResourceGeneric",
			Metadata: &shared.CloudResourceMetadata{
				Name: "test-resource",
			},
			Spec: &testgenericv1.TestCloudResourceGenericSpec{
				StringField: proto.String("custom-value"),
				// Other fields left unset (nil)
			},
		}

		err := ApplyDefaults(msg)
		require.NoError(t, err)

		// Verify: custom value preserved, defaults applied to unset fields
		require.NotNil(t, msg.Spec.StringField)
		assert.Equal(t, "custom-value", *msg.Spec.StringField)

		require.NotNil(t, msg.Spec.Int32Field)
		assert.Equal(t, int32(42), *msg.Spec.Int32Field)

		require.NotNil(t, msg.Spec.Int64Field)
		assert.Equal(t, int64(9999), *msg.Spec.Int64Field)

		require.NotNil(t, msg.Spec.BoolField)
		assert.True(t, *msg.Spec.BoolField)
	})

	t.Run("handles nil message gracefully", func(t *testing.T) {
		err := ApplyDefaults(nil)
		assert.NoError(t, err)
	})

	t.Run("handles nil spec gracefully", func(t *testing.T) {
		msg := &testgenericv1.TestCloudResourceGeneric{
			ApiVersion: "_test.openmcf.org/v1",
			Kind:       "TestCloudResourceGeneric",
			Metadata: &shared.CloudResourceMetadata{
				Name: "test-resource",
			},
			Spec: nil,
		}

		err := ApplyDefaults(msg)
		assert.NoError(t, err)
	})
}

func TestApplyDefaults_NestedMessages(t *testing.T) {
	t.Run("applies defaults recursively to nested messages", func(t *testing.T) {
		// Create message with nested structure
		msg := &testgenericv1.TestCloudResourceGeneric{
			ApiVersion: "_test.openmcf.org/v1",
			Kind:       "TestCloudResourceGeneric",
			Metadata: &shared.CloudResourceMetadata{
				Name: "test-resource",
			},
			Spec: &testgenericv1.TestCloudResourceGenericSpec{
				// Leave defaults unset at spec level
				Nested: &testgenericv1.TestGenericNestedMessage{
					// Leave nested defaults unset (nil pointers)
				},
			},
		}

		err := ApplyDefaults(msg)
		require.NoError(t, err)

		// Verify defaults were applied at all levels
		require.NotNil(t, msg.Spec.StringField)
		assert.Equal(t, "default-string", *msg.Spec.StringField)

		require.NotNil(t, msg.Spec.Int32Field)
		assert.Equal(t, int32(42), *msg.Spec.Int32Field)

		// Verify nested defaults
		require.NotNil(t, msg.Spec.Nested.NestedString)
		assert.Equal(t, "nested-default", *msg.Spec.Nested.NestedString)

		require.NotNil(t, msg.Spec.Nested.NestedInt)
		assert.Equal(t, int32(99), *msg.Spec.Nested.NestedInt)
	})
}

func TestApplyDefaults_FieldsWithoutDefaults(t *testing.T) {
	t.Run("leaves fields without defaults unchanged", func(t *testing.T) {
		msg := &testgenericv1.TestCloudResourceGeneric{
			ApiVersion: "_test.openmcf.org/v1",
			Kind:       "TestCloudResourceGeneric",
			Metadata: &shared.CloudResourceMetadata{
				Name: "test-resource",
			},
			Spec: &testgenericv1.TestCloudResourceGenericSpec{
				// string_no_default field has no default option
			},
		}

		err := ApplyDefaults(msg)
		require.NoError(t, err)

		// Fields with defaults should be set
		require.NotNil(t, msg.Spec.StringField)
		assert.Equal(t, "default-string", *msg.Spec.StringField)

		require.NotNil(t, msg.Spec.Int32Field)
		assert.Equal(t, int32(42), *msg.Spec.Int32Field)

		// Field without default should remain empty
		assert.Equal(t, "", msg.Spec.StringNoDefault)
	})
}

func TestApplyDefaults_ZeroValuesPreserved(t *testing.T) {
	t.Run("preserves explicitly set zero values for all scalar types", func(t *testing.T) {
		// This is THE critical test that validates the bug fix!
		// With optional fields, we can now distinguish "not set" from "set to zero value"
		msg := &testgenericv1.TestCloudResourceGeneric{
			ApiVersion: "_test.openmcf.org/v1",
			Kind:       "TestCloudResourceGeneric",
			Metadata: &shared.CloudResourceMetadata{
				Name: "test-resource",
			},
			Spec: &testgenericv1.TestCloudResourceGenericSpec{
				// Explicitly set ALL fields to their ZERO values
				StringField: proto.String(""),   // Empty string (zero value for string)
				Int32Field:  proto.Int32(0),     // Zero (zero value for int32)
				Int64Field:  proto.Int64(0),     // Zero (zero value for int64)
				Uint32Field: proto.Uint32(0),    // Zero (zero value for uint32)
				Uint64Field: proto.Uint64(0),    // Zero (zero value for uint64)
				FloatField:  proto.Float32(0.0), // Zero (zero value for float)
				DoubleField: proto.Float64(0.0), // Zero (zero value for double)
				BoolField:   proto.Bool(false),  // False (zero value for bool)
			},
		}

		// Apply defaults
		err := ApplyDefaults(msg)
		require.NoError(t, err)

		// CRITICAL: All zero values should be PRESERVED, NOT replaced with defaults!
		// Before the fix: defaults would be applied because Has() returned false
		// After the fix: zero values are preserved because Has() returns true for non-nil pointers

		require.NotNil(t, msg.Spec.StringField)
		assert.Equal(t, "", *msg.Spec.StringField, "Empty string should be preserved, not defaulted")

		require.NotNil(t, msg.Spec.Int32Field)
		assert.Equal(t, int32(0), *msg.Spec.Int32Field, "Zero int32 should be preserved, not defaulted to 42")

		require.NotNil(t, msg.Spec.Int64Field)
		assert.Equal(t, int64(0), *msg.Spec.Int64Field, "Zero int64 should be preserved, not defaulted to 9999")

		require.NotNil(t, msg.Spec.Uint32Field)
		assert.Equal(t, uint32(0), *msg.Spec.Uint32Field, "Zero uint32 should be preserved, not defaulted to 100")

		require.NotNil(t, msg.Spec.Uint64Field)
		assert.Equal(t, uint64(0), *msg.Spec.Uint64Field, "Zero uint64 should be preserved, not defaulted to 50000")

		require.NotNil(t, msg.Spec.FloatField)
		assert.Equal(t, float32(0.0), *msg.Spec.FloatField, "Zero float should be preserved, not defaulted to 3.14")

		require.NotNil(t, msg.Spec.DoubleField)
		assert.Equal(t, float64(0.0), *msg.Spec.DoubleField, "Zero double should be preserved, not defaulted to 2.718")

		require.NotNil(t, msg.Spec.BoolField)
		assert.False(t, *msg.Spec.BoolField, "False should be preserved, not defaulted to true")
	})

	t.Run("zero values in nested messages are preserved", func(t *testing.T) {
		msg := &testgenericv1.TestCloudResourceGeneric{
			ApiVersion: "_test.openmcf.org/v1",
			Kind:       "TestCloudResourceGeneric",
			Metadata: &shared.CloudResourceMetadata{
				Name: "test-resource",
			},
			Spec: &testgenericv1.TestCloudResourceGenericSpec{
				Nested: &testgenericv1.TestGenericNestedMessage{
					NestedString: proto.String(""), // Empty string
					NestedInt:    proto.Int32(0),   // Zero
				},
			},
		}

		err := ApplyDefaults(msg)
		require.NoError(t, err)

		// Zero values in nested messages should also be preserved
		require.NotNil(t, msg.Spec.Nested.NestedString)
		assert.Equal(t, "", *msg.Spec.Nested.NestedString, "Nested empty string should be preserved")

		require.NotNil(t, msg.Spec.Nested.NestedInt)
		assert.Equal(t, int32(0), *msg.Spec.Nested.NestedInt, "Nested zero int32 should be preserved")
	})
}

func TestApplyDefaults_Idempotency(t *testing.T) {
	t.Run("applying defaults multiple times is idempotent", func(t *testing.T) {
		msg := &testgenericv1.TestCloudResourceGeneric{
			ApiVersion: "_test.openmcf.org/v1",
			Kind:       "TestCloudResourceGeneric",
			Metadata: &shared.CloudResourceMetadata{
				Name: "test-resource",
			},
			Spec: &testgenericv1.TestCloudResourceGenericSpec{},
		}

		// Apply defaults first time
		err := ApplyDefaults(msg)
		require.NoError(t, err)
		firstResult := proto.Clone(msg)

		// Apply defaults second time
		err = ApplyDefaults(msg)
		require.NoError(t, err)
		secondResult := msg

		// Results should be identical
		assert.True(t, proto.Equal(firstResult, secondResult),
			"Applying defaults multiple times should produce identical results")
	})
}

func TestApplyDefaults_UnsetNestedMessageBehavior(t *testing.T) {
	t.Run("unset nested messages remain unset to preserve user intent", func(t *testing.T) {
		// Create message WITHOUT nested message set - the key scenario!
		// This simulates when a YAML manifest has spec.some_field but NOT spec.nested
		// Semantically: "I don't want this optional feature"
		msg := &testgenericv1.TestCloudResourceGeneric{
			ApiVersion: "_test.openmcf.org/v1",
			Kind:       "TestCloudResourceGeneric",
			Metadata: &shared.CloudResourceMetadata{
				Name: "test-resource",
			},
			Spec: &testgenericv1.TestCloudResourceGenericSpec{
				// Nested field is NOT set (nil)
				// Even though TestNestedMessage has fields with defaults,
				// we should NOT auto-initialize it - user didn't request this feature
			},
		}

		// Verify nested is nil before applying defaults
		assert.Nil(t, msg.Spec.Nested, "nested should be nil before applying defaults")

		// Apply defaults
		err := ApplyDefaults(msg)
		require.NoError(t, err)

		// CRITICAL: Nested message should remain nil!
		// This preserves user intent: "I didn't set this optional feature, so don't enable it"
		assert.Nil(t, msg.Spec.Nested, "unset nested message should remain nil to preserve user intent")

		// Top-level defaults should still be applied
		require.NotNil(t, msg.Spec.StringField)
		assert.Equal(t, "default-string", *msg.Spec.StringField)
	})

	t.Run("empty nested message triggers default application", func(t *testing.T) {
		// User explicitly sets nested message to empty: `nested: {}`
		// This signals: "I want this feature with defaults"
		msg := &testgenericv1.TestCloudResourceGeneric{
			ApiVersion: "_test.openmcf.org/v1",
			Kind:       "TestCloudResourceGeneric",
			Metadata: &shared.CloudResourceMetadata{
				Name: "test-resource",
			},
			Spec: &testgenericv1.TestCloudResourceGenericSpec{
				// Empty nested message - user is opting in to defaults
				Nested: &testgenericv1.TestGenericNestedMessage{},
			},
		}

		// Verify nested is set but has no values
		require.NotNil(t, msg.Spec.Nested, "nested should be set before applying defaults")
		assert.Nil(t, msg.Spec.Nested.NestedString, "nested string should be nil before defaults")

		// Apply defaults
		err := ApplyDefaults(msg)
		require.NoError(t, err)

		// Defaults should be applied because user explicitly set the message
		require.NotNil(t, msg.Spec.Nested.NestedString)
		assert.Equal(t, "nested-default", *msg.Spec.Nested.NestedString,
			"nested string should have its default value")

		require.NotNil(t, msg.Spec.Nested.NestedInt)
		assert.Equal(t, int32(99), *msg.Spec.Nested.NestedInt,
			"nested int should have its default value")
	})

	t.Run("unset messages without defaults also remain unset", func(t *testing.T) {
		// TestCloudResourceGeneric has metadata field which is a message without defaults
		// It should NOT be created automatically
		msg := &testgenericv1.TestCloudResourceGeneric{
			ApiVersion: "_test.openmcf.org/v1",
			Kind:       "TestCloudResourceGeneric",
			// Metadata is nil - and CloudResourceMetadata has no fields with defaults
			Spec: &testgenericv1.TestCloudResourceGenericSpec{},
		}

		// Verify metadata is nil before applying defaults
		assert.Nil(t, msg.Metadata, "metadata should be nil before applying defaults")

		// Apply defaults
		err := ApplyDefaults(msg)
		require.NoError(t, err)

		// Metadata should remain nil
		assert.Nil(t, msg.Metadata, "metadata should remain nil")
	})
}
