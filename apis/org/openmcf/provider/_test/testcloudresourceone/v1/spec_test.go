// spec_test.go — Boundary tests for StringValueOrRef's message-level CEL rule.
//
// INSTITUTIONAL KNOWLEDGE — READ THIS BEFORE MODIFYING OR REMOVING:
//
// These tests validate the message-level CEL rule added to StringValueOrRef in
// foreign_key.proto (id: "string_value_or_ref.non_empty"). That rule enforces:
//
//   (has(this.value) && this.value != '') || has(this.value_from)
//
// i.e., whenever a StringValueOrRef message is present, it must carry meaningful
// content — either a non-empty literal string or a ValueFromRef.
//
// WHY THIS RULE EXISTS:
// buf.validate's `required = true` on a message-type field only checks whether the
// message is present (not nil). It does NOT check whether the inner oneof has content.
// This created a false-positive in client-side proto validation: wizard specDefaults
// created `namespace: { value: "" }` which passed `required` (message exists) while
// carrying no meaningful value. Combined with pipeline placeholder injection for image
// and version fields, the entire spec appeared valid — showing a green "validated"
// badge on the wizard review step when required fields were actually empty.
//
// WHY TESTS LIVE ON TestCloudResourceOne (NOT on production resources):
// Production cloud resources (KubernetesDeployment, GcpSecretsManager, AwsEcsService,
// etc.) may be removed or refactored. If these boundary tests lived only on those
// resources, the institutional knowledge would silently vanish with them.
// TestCloudResourceOne is permanent test infrastructure in the `_test/` directory —
// these tests survive regardless of which production resources come and go.
//
// The spec has two StringValueOrRef fields specifically for this purpose:
//   - required_ref (with `required = true`) — mirrors KubernetesDeploymentSpec.namespace
//   - optional_ref (without `required`)     — proves the CEL rule fires on message
//                                              presence, not on the field annotation
//
// A parallel cross-cutting test in foreign_key_test.go imports this package to keep
// the coverage co-located with the message definition. If this test resource is ever
// removed, that import breaks at compile time — flagging the loss of coverage.
//
// RELATED FILES:
//   - apis/org/openmcf/shared/foreignkey/v1/foreign_key.proto  (the CEL rule)
//   - apis/org/openmcf/shared/foreignkey/v1/foreign_key_test.go (message-level + cross-cutting tests)
//   - apis/org/openmcf/provider/_test/testcloudresourceone/v1/spec.proto (field definitions)

package testcloudresourceonev1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared/cloudresourcekind"
	foreignkeyv1 "github.com/plantonhq/openmcf/apis/org/openmcf/shared/foreignkey/v1"
)

func TestTestCloudResourceOneSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "TestCloudResourceOne StringValueOrRef Validation Suite")
}

// validTestCloudResourceOne returns a fully valid TestCloudResourceOne envelope.
// Tests mutate specific fields from this baseline to isolate individual behaviors.
func validTestCloudResourceOne() *TestCloudResourceOne {
	return &TestCloudResourceOne{
		ApiVersion: "_test.openmcf.org/v1",
		Kind:       "TestCloudResourceOne",
		Metadata: &shared.CloudResourceMetadata{
			Name: "test-resource",
		},
		Spec: &TestCloudResourceOneSpec{
			RequiredRef: &foreignkeyv1.StringValueOrRef{
				LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
					Value: "valid-value",
				},
			},
		},
	}
}

var _ = ginkgo.Describe("StringValueOrRef CEL Rule — Consumer-Level Boundary Tests", func() {

	// ─── Required field (required_ref) ──────────────────────────────────────────
	//
	// Tests the interaction of `(buf.validate.field).required = true` with
	// StringValueOrRef's message-level CEL rule. This is the exact combination
	// that caused the wizard validation false-positive: `required` only checked
	// message presence, so `{ value: "" }` passed. The CEL rule now rejects it.

	ginkgo.Describe("required_ref field", func() {

		ginkgo.Context("with a non-empty literal value", func() {
			ginkgo.It("should pass validation", func() {
				input := validTestCloudResourceOne()
				input.Spec.RequiredRef = &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
						Value: "my-namespace",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with a valid ValueFromRef", func() {
			ginkgo.It("should pass validation", func() {
				input := validTestCloudResourceOne()
				input.Spec.RequiredRef = &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
						ValueFrom: &foreignkeyv1.ValueFromRef{
							Kind: cloudresourcekind.CloudResourceKind_TestCloudResourceOne,
							Env:  "dev",
							Name: "my-ref",
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		// THE FALSE-POSITIVE CASE: This is the exact scenario that caused the wizard
		// bug. The message is present (passes `required`), the oneof is set to `value`,
		// but the string is empty. Before the CEL rule, this passed validation silently.
		ginkgo.Context("with an empty literal value", func() {
			ginkgo.It("should return a validation error", func() {
				input := validTestCloudResourceOne()
				input.Spec.RequiredRef = &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
						Value: "",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		// UNSET ONEOF: The message exists (not nil) but no oneof branch is selected.
		// `required` passes (message present), but the CEL rule rejects (no content).
		ginkgo.Context("with unset oneof (empty struct)", func() {
			ginkgo.It("should return a validation error", func() {
				input := validTestCloudResourceOne()
				input.Spec.RequiredRef = &foreignkeyv1.StringValueOrRef{}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		// NIL FIELD: The message is absent. `required` catches this independently
		// of the CEL rule. This test confirms both layers work together.
		ginkgo.Context("with nil (absent)", func() {
			ginkgo.It("should return a validation error", func() {
				input := validTestCloudResourceOne()
				input.Spec.RequiredRef = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})
	})

	// ─── Optional field (optional_ref) ──────────────────────────────────────────
	//
	// These tests prove a subtle but critical property: the CEL rule lives on the
	// StringValueOrRef MESSAGE DEFINITION, not on consumer field annotations. It
	// fires whenever the message is PRESENT, regardless of whether the field has
	// `required = true`. The only way to bypass validation for an optional field
	// is to leave it nil (absent).
	//
	// This distinction matters because many StringValueOrRef fields in the codebase
	// are used without `required` (e.g., environment variable map values). If those
	// map entries exist, their values must have content. To "remove" an optional
	// value, you remove the map entry or leave the field nil — not set it to "".

	ginkgo.Describe("optional_ref field", func() {

		ginkgo.Context("when nil (absent)", func() {
			ginkgo.It("should pass validation — field absent means CEL rule never fires", func() {
				input := validTestCloudResourceOne()
				input.Spec.OptionalRef = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with a non-empty literal value", func() {
			ginkgo.It("should pass validation", func() {
				input := validTestCloudResourceOne()
				input.Spec.OptionalRef = &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
						Value: "optional-but-valid",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with a valid ValueFromRef", func() {
			ginkgo.It("should pass validation", func() {
				input := validTestCloudResourceOne()
				input.Spec.OptionalRef = &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
						ValueFrom: &foreignkeyv1.ValueFromRef{
							Kind: cloudresourcekind.CloudResourceKind_TestCloudResourceOne,
							Env:  "staging",
							Name: "optional-ref",
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		// KEY TEST: The field is optional, but the message IS present with an empty
		// value. The CEL rule still fires and rejects it. This proves the rule is on
		// the message, not the field annotation.
		ginkgo.Context("with an empty literal value", func() {
			ginkgo.It("should return a validation error — CEL fires on message presence, not field annotation", func() {
				input := validTestCloudResourceOne()
				input.Spec.OptionalRef = &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
						Value: "",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("with unset oneof (empty struct)", func() {
			ginkgo.It("should return a validation error — message present but no content", func() {
				input := validTestCloudResourceOne()
				input.Spec.OptionalRef = &foreignkeyv1.StringValueOrRef{}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})
	})
})
