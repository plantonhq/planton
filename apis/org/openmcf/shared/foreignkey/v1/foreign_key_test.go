// foreign_key_test.go — Message-level isolation tests for StringValueOrRef's CEL rule.
//
// These test StringValueOrRef directly (no consumer wrapper message) to validate that
// the CEL rule (id: "string_value_or_ref.non_empty") works at the message level.
//
// Cross-cutting consumer tests that import TestCloudResourceGeneric live in a
// separate file (foreign_key_consumer_test.go) using Go's external test package
// convention (`package foreignkeyv1_test`). This avoids Go's import cycle
// restriction: foreignkeyv1 -> testcloudresourcegenericv1 -> foreignkeyv1 would
// form a cycle in the same package, but the external test package breaks it.
//
// RELATED FILES:
//   - foreign_key.proto                              (the CEL rule)
//   - foreign_key_consumer_test.go                   (cross-cutting consumer tests with TestCloudResourceGeneric)
//   - _test/testcloudresourcegeneric/v1/spec_test.go (comprehensive boundary tests)

package foreignkeyv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared/cloudresourcekind"
)

func TestForeignKey(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "ForeignKey Suite")
}

var _ = ginkgo.Describe("StringValueOrRef Validation", func() {

	// ─── Message-level isolation tests ──────────────────────────────────────────
	//
	// These test StringValueOrRef directly, without any consumer wrapper message.
	// They validate that the CEL rule (id: "string_value_or_ref.non_empty") works
	// at the message level before testing its behavior inside consumer fields.

	ginkgo.Describe("message-level CEL rule", func() {

		ginkgo.Context("with a non-empty literal value", func() {
			ginkgo.It("should pass validation", func() {
				input := &StringValueOrRef{
					LiteralOrRef: &StringValueOrRef_Value{
						Value: "my-string",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with a valid ValueFromRef", func() {
			ginkgo.It("should pass validation", func() {
				input := &StringValueOrRef{
					LiteralOrRef: &StringValueOrRef_ValueFrom{
						ValueFrom: &ValueFromRef{
							Kind:      cloudresourcekind.CloudResourceKind_TestCloudResourceGeneric,
							Env:       "dev",
							Name:      "my-cert",
							FieldPath: "status.outputs.cert_arn",
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("when proto oneof overwrites (last-write wins)", func() {
			ginkgo.It("should validate the final oneof branch", func() {
				input := &StringValueOrRef{
					LiteralOrRef: &StringValueOrRef_Value{
						Value: "my-string",
					},
				}
				input.LiteralOrRef = &StringValueOrRef_ValueFrom{
					ValueFrom: &ValueFromRef{
						Kind: cloudresourcekind.CloudResourceKind_TestCloudResourceGeneric,
						Env:  "dev",
						Name: "overwrites-literal",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())

				gomega.Expect(input.GetValue()).To(gomega.Equal(""))
				gomega.Expect(input.GetValueFrom().GetName()).To(gomega.Equal("overwrites-literal"))
			})
		})

		// EMPTY VALUE: The oneof is set to `value` with an empty string. This is the
		// core false-positive scenario. Before the CEL rule, `required = true` on the
		// consumer field passed because the message was present. Now rejected.
		ginkgo.Context("with an empty literal value", func() {
			ginkgo.It("should return a validation error", func() {
				input := &StringValueOrRef{
					LiteralOrRef: &StringValueOrRef_Value{
						Value: "",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		// UNSET ONEOF: The message exists but no oneof branch is selected.
		// Neither `has(this.value)` nor `has(this.value_from)` is true.
		ginkgo.Context("with unset oneof (empty struct)", func() {
			ginkgo.It("should return a validation error", func() {
				input := &StringValueOrRef{}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})
	})
})
