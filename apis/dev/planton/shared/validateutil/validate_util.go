package validateutil

import (
	"buf.build/gen/go/bufbuild/protovalidate/protocolbuffers/go/buf/validate"
	"github.com/onsi/gomega"
)

type ExpectedViolation struct {
	FieldPath    string
	ConstraintId string
	Message      string
}

const (
	RequiredConstraint        = "required"
	RequiredViolationMessage  = "value is required"
	StringInConstraint        = "string.in"
	StringMinLengthConstraint = "string.min_len"
	StringMaxLengthConstraint = "string.max_len"
)

// Match is a helper to compare an actual violation
// against an expected one. This way, your test code is clean, and
// all the pointer-taking logic is in one place.
func Match(actual *validate.Violation, expected *ExpectedViolation) {
	// Convert the expected FieldPath to a pointer for a BeEquivalentTo check
	gomega.Expect(actual.Field).To(gomega.BeEquivalentTo(&expected.FieldPath))

	// Same pattern for constraint ID
	var wantConstraintId = expected.ConstraintId
	gomega.Expect(actual.RuleId).To(gomega.BeEquivalentTo(&wantConstraintId))

	// And again for the violation message
	var wantMessage = expected.Message
	gomega.Expect(actual.Message).To(gomega.BeEquivalentTo(&wantMessage))
}
