package fieldsextractor

import (
	"testing"

	awslambdav1 "github.com/plantonhq/planton/apis/dev/planton/provider/aws/awslambda/v1"
	"google.golang.org/protobuf/proto"

	"github.com/stretchr/testify/assert"
)

// TestExtractApiResourceSpecField_Success tests the success scenario where fields are valid and messages are correctly structured.
func TestExtractApiResourceSpecField_Success(t *testing.T) {
	// Create mock input
	mockSpec := &awslambdav1.AwsLambdaSpec{}
	mockTarget := &awslambdav1.AwsLambda{Spec: mockSpec}
	mockStackInput := &awslambdav1.AwsLambdaStackInput{Target: mockTarget}

	// Call the function under test
	result, err := ExtractApiResourceSpecField(mockStackInput)

	// Assert no error and check result is as expected
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, proto.Equal((*result).Interface().(proto.Message), mockSpec), "Expected and actual protobuf messages are not equal")
}

// TestExtractApiResourceSpecField_NilInput tests the scenario where stackInput is nil.
func TestExtractApiResourceSpecField_NilInput(t *testing.T) {
	// Call the function under test with nil input
	result, err := ExtractApiResourceSpecField(nil)

	// Assert that an error is returned and result is nil
	assert.Nil(t, result)
	assert.EqualError(t, err, "stack-input is nil")
}

// TestExtractApiResourceSpecField_InvalidTargetField tests the scenario where the target field is missing or invalid.
func TestExtractApiResourceSpecField_InvalidTargetField(t *testing.T) {
	// Create a mock input without a valid PostgresOperatorKubernetes field
	mockStackInput := &awslambdav1.AwsLambdaStackInput{Target: nil}

	// Call the function under test
	result, err := ExtractApiResourceSpecField(mockStackInput)

	// Assert that an error is returned
	assert.Nil(t, result)
	assert.EqualError(t, err, "Field target is nil in stack-input")
}

// TestExtractApiResourceSpecField_InvalidSpecField tests the scenario where the spec field in the target is missing or invalid.
func TestExtractApiResourceSpecField_InvalidSpecField(t *testing.T) {
	// Create a mock input with an invalid Spec field
	mockTarget := &awslambdav1.AwsLambda{Spec: nil}
	mockStackInput := &awslambdav1.AwsLambdaStackInput{Target: mockTarget}

	// Call the function under test
	result, err := ExtractApiResourceSpecField(mockStackInput)

	// Assert that an error is returned
	assert.Nil(t, result)
	assert.EqualError(t, err, "Field spec is nil in target")
}
