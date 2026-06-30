package testcloudresourcegenericv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/apis/dev/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/apis/dev/planton/shared/foreignkey/v1"
)

func TestTestCloudResourceGenericSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "TestCloudResourceGenericSpec Validation Suite")
}

// validResource returns a minimal valid TestCloudResourceGeneric for mutation-based testing.
func validResource() *TestCloudResourceGeneric {
	return &TestCloudResourceGeneric{
		ApiVersion: "_test.planton.dev/v1",
		Kind:       "TestCloudResourceGeneric",
		Metadata: &shared.CloudResourceMetadata{
			Name: "test-resource",
		},
		Spec: &TestCloudResourceGenericSpec{
			RequiredRef: &foreignkeyv1.StringValueOrRef{
				LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
					Value: "valid-value",
				},
			},
		},
	}
}

var _ = ginkgo.Describe("TestCloudResourceGeneric Validation", func() {

	ginkgo.Describe("required_ref (StringValueOrRef with required=true)", func() {

		ginkgo.Context("with a valid literal value", func() {
			ginkgo.It("should pass validation", func() {
				input := validResource()
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with nil (absent)", func() {
			ginkgo.It("should fail — required field is missing", func() {
				input := validResource()
				input.Spec.RequiredRef = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("with empty struct", func() {
			ginkgo.It("should fail — CEL rule rejects empty StringValueOrRef", func() {
				input := validResource()
				input.Spec.RequiredRef = &foreignkeyv1.StringValueOrRef{}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("with empty string value", func() {
			ginkgo.It("should fail — CEL rule rejects empty value", func() {
				input := validResource()
				input.Spec.RequiredRef = &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
						Value: "",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("optional_ref (StringValueOrRef without required)", func() {

		ginkgo.Context("when nil (absent)", func() {
			ginkgo.It("should pass — absent optional field is valid", func() {
				input := validResource()
				input.Spec.OptionalRef = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with a valid literal value", func() {
			ginkgo.It("should pass", func() {
				input := validResource()
				input.Spec.OptionalRef = &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
						Value: "optional-value",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with empty struct", func() {
			ginkgo.It("should fail — CEL fires on message presence, not field annotation", func() {
				input := validResource()
				input.Spec.OptionalRef = &foreignkeyv1.StringValueOrRef{}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("envelope validation", func() {

		ginkgo.Context("with wrong api_version", func() {
			ginkgo.It("should fail", func() {
				input := validResource()
				input.ApiVersion = "wrong/v1"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("with wrong kind", func() {
			ginkgo.It("should fail", func() {
				input := validResource()
				input.Kind = "WrongKind"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("with nil metadata", func() {
			ginkgo.It("should fail", func() {
				input := validResource()
				input.Metadata = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})
	})
})
