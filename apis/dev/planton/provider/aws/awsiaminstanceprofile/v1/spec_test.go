package awsiaminstanceprofilev1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/apis/dev/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/apis/dev/planton/shared/foreignkey/v1"
)

func TestAwsIamInstanceProfileSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AwsIamInstanceProfileSpec Validation Tests")
}

func literalRole(name string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: name},
	}
}

// minimalValidProfile is the common case: a region and a role reference.
func minimalValidProfile() *AwsIamInstanceProfile {
	return &AwsIamInstanceProfile{
		ApiVersion: "aws.planton.dev/v1",
		Kind:       "AwsIamInstanceProfile",
		Metadata: &shared.CloudResourceMetadata{
			Name: "web-server-profile",
		},
		Spec: &AwsIamInstanceProfileSpec{
			Region: "us-west-2",
			Role:   literalRole("web-server-role"),
		},
	}
}

var _ = ginkgo.Describe("AwsIamInstanceProfileSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("aws_iam_instance_profile", func() {

			ginkgo.It("should not return a validation error for a minimal profile", func() {
				err := protovalidate.Validate(minimalValidProfile())
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error when the role is a reference", func() {
				input := minimalValidProfile()
				input.Spec.Role = &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
						ValueFrom: &foreignkeyv1.ValueFromRef{
							Name: "web-server-role",
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with a path", func() {
				input := minimalValidProfile()
				input.Spec.Path = "/compute/"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for the root path", func() {
				input := minimalValidProfile()
				input.Spec.Path = "/"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with full metadata set", func() {
				input := minimalValidProfile()
				input.Metadata = &shared.CloudResourceMetadata{
					Name:   "web-server-profile",
					Org:    "acme-corp",
					Env:    "production",
					Labels: map[string]string{"team": "platform"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("aws_iam_instance_profile", func() {

			ginkgo.It("should return a validation error when api_version is wrong", func() {
				input := minimalValidProfile()
				input.ApiVersion = "wrong.planton.dev/v1"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when kind is wrong", func() {
				input := minimalValidProfile()
				input.Kind = "WrongKind"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when metadata is missing", func() {
				input := minimalValidProfile()
				input.Metadata = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when spec is missing", func() {
				input := &AwsIamInstanceProfile{
					ApiVersion: "aws.planton.dev/v1",
					Kind:       "AwsIamInstanceProfile",
					Metadata:   &shared.CloudResourceMetadata{Name: "web-server-profile"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when region is empty", func() {
				input := minimalValidProfile()
				input.Spec.Region = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when role is missing", func() {
				input := minimalValidProfile()
				input.Spec.Role = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when path does not begin with '/'", func() {
				input := minimalValidProfile()
				input.Spec.Path = "compute/"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when path does not end with '/'", func() {
				input := minimalValidProfile()
				input.Spec.Path = "/compute"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when path contains an empty segment", func() {
				input := minimalValidProfile()
				input.Spec.Path = "/compute//profiles/"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})
	})
})
