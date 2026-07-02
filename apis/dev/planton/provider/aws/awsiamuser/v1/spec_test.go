package awsiamuserv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/apis/dev/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/apis/dev/planton/shared/foreignkey/v1"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestAwsIamUserSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AwsIamUserSpec Validation Tests")
}

func literalArn(arn string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: arn},
	}
}

// minimalValidUser is the common case: a region and a user name.
func minimalValidUser() *AwsIamUser {
	return &AwsIamUser{
		ApiVersion: "aws.planton.dev/v1",
		Kind:       "AwsIamUser",
		Metadata: &shared.CloudResourceMetadata{
			Name: "test-iam-user",
		},
		Spec: &AwsIamUserSpec{
			Region:   "us-west-2",
			UserName: "test-ci-user",
		},
	}
}

var _ = ginkgo.Describe("AwsIamUserSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("aws_iam_user", func() {

			ginkgo.It("should not return a validation error for a minimal user", func() {
				err := protovalidate.Validate(minimalValidUser())
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for a fully-specified user", func() {
				input := minimalValidUser()
				input.Spec.Path = "/ci/"
				input.Spec.ManagedPolicyArns = []*foreignkeyv1.StringValueOrRef{
					literalArn("arn:aws:iam::aws:policy/ReadOnlyAccess"),
				}
				input.Spec.PermissionsBoundary = literalArn("arn:aws:iam::123456789012:policy/boundaries/ci")
				input.Spec.DisableAccessKeys = true
				input.Spec.ForceDestroy = true
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error when managed policies are references", func() {
				input := minimalValidUser()
				input.Spec.ManagedPolicyArns = []*foreignkeyv1.StringValueOrRef{
					{
						LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
							ValueFrom: &foreignkeyv1.ValueFromRef{Name: "s3-read-only"},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with inline policies", func() {
				input := minimalValidUser()
				inline, err := structpb.NewStruct(map[string]interface{}{
					"Version": "2012-10-17",
					"Statement": []interface{}{
						map[string]interface{}{
							"Effect":   "Allow",
							"Action":   "s3:ListBucket",
							"Resource": "arn:aws:s3:::demo-bucket",
						},
					},
				})
				gomega.Expect(err).To(gomega.BeNil())
				input.Spec.InlinePolicies = map[string]*structpb.Struct{"bucketAccess": inline}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("aws_iam_user", func() {

			ginkgo.It("should return a validation error when user_name is missing", func() {
				input := minimalValidUser()
				input.Spec.UserName = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when user_name has illegal characters", func() {
				input := minimalValidUser()
				input.Spec.UserName = "bad name!"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when region is empty", func() {
				input := minimalValidUser()
				input.Spec.Region = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when path does not begin and end with '/'", func() {
				input := minimalValidUser()
				input.Spec.Path = "ci"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when path contains an empty segment", func() {
				input := minimalValidUser()
				input.Spec.Path = "/ci//bots/"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when a managed policy entry is empty", func() {
				input := minimalValidUser()
				input.Spec.ManagedPolicyArns = []*foreignkeyv1.StringValueOrRef{
					{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: ""}},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when an inline policy name exceeds 128 characters", func() {
				input := minimalValidUser()
				long := make([]byte, 129)
				for i := range long {
					long[i] = 'a'
				}
				doc, err := structpb.NewStruct(map[string]interface{}{"Version": "2012-10-17"})
				gomega.Expect(err).To(gomega.BeNil())
				input.Spec.InlinePolicies = map[string]*structpb.Struct{string(long): doc}
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})
		})
	})
})
