// foreign_key_consumer_test.go — Cross-cutting consumer tests for StringValueOrRef's CEL rule.
//
// WHY THIS IS A SEPARATE FILE (not in foreign_key_test.go):
// Go's import cycle restriction prevents `package foreignkeyv1` from importing
// testcatalogkindgenericv1alpha1, because that generated code already imports foreignkeyv1
// (it uses StringValueOrRef). The standard Go solution is an external test package:
// `package foreignkeyv1_test` is treated as a separate package by the Go toolchain,
// breaking the cycle while still being in the same directory.
//
// WHY THESE TESTS EXIST HERE (co-located with StringValueOrRef):
// The comprehensive boundary tests live in testcatalogkindgeneric/v1/spec_test.go.
// These tests serve a different purpose: they are a STRUCTURAL TRIPWIRE. By importing
// testcatalogkindgenericv1alpha1 in the foreignkey package's test directory, we create a
// compile-time dependency. If TestCatalogKindGeneric is ever removed or its package
// path changes, this file fails to compile — immediately flagging the loss of
// consumer-level StringValueOrRef validation coverage.
//
// The tests here are intentionally minimal (4 cases) because they exist for structural
// integrity, not exhaustive coverage. spec_test.go in testcatalogkindgeneric has the
// full boundary matrix.
//
// RELATED FILES:
//   - foreign_key.proto                              (the CEL rule)
//   - foreign_key_test.go                            (message-level isolation tests)
//   - _test/testcatalogkindgeneric/v1/spec.proto   (required_ref + optional_ref fields)
//   - _test/testcatalogkindgeneric/v1/spec_test.go (comprehensive boundary tests)

package foreignkeyv1_test

import (
	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"

	// STRUCTURAL TRIPWIRE: This import creates a compile-time dependency on
	// TestCatalogKindGeneric. If that package is ever removed, this file fails
	// to compile — alerting maintainers that the cross-cutting consumer tests
	// below (and the comprehensive boundary tests in spec_test.go) need a new home.
	testresource "github.com/plantonhq/planton/catalog/_test/testcatalogkindgeneric/v1alpha1"
)

// validTestResource returns a valid TestCatalogKindGeneric envelope for mutation-based testing.
func validTestResource() *testresource.TestCatalogKindGeneric {
	return &testresource.TestCatalogKindGeneric{
		ApiVersion: "_test.planton.dev/v1alpha1",
		Kind:       "TestCatalogKindGeneric",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-resource",
		},
		Spec: &testresource.TestCatalogKindGenericSpec{
			RequiredRef: &foreignkeyv1.StringValueOrRef{
				LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
					Value: "valid-value",
				},
			},
		},
	}
}

var _ = ginkgo.Describe("StringValueOrRef — Cross-Cutting Consumer Validation", func() {

	// These tests validate that the message-level CEL rule propagates correctly
	// through protovalidate's recursive validation when StringValueOrRef is used
	// as a field inside a full infra component envelope.

	ginkgo.Describe("required_ref on TestCatalogKindGeneric", func() {

		ginkgo.Context("with empty struct", func() {
			ginkgo.It("should fail — CEL rule rejects empty message inside required field", func() {
				input := validTestResource()
				input.Spec.RequiredRef = &foreignkeyv1.StringValueOrRef{}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("with valid value", func() {
			ginkgo.It("should pass", func() {
				input := validTestResource()
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("optional_ref on TestCatalogKindGeneric", func() {

		ginkgo.Context("with empty struct", func() {
			ginkgo.It("should fail — CEL fires on message presence, not field annotation", func() {
				input := validTestResource()
				input.Spec.OptionalRef = &foreignkeyv1.StringValueOrRef{}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("when nil (absent)", func() {
			ginkgo.It("should pass — absent message means CEL rule never fires", func() {
				input := validTestResource()
				input.Spec.OptionalRef = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})
})
