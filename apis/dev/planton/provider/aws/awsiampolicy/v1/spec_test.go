package awsiampolicyv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/apis/dev/planton/shared"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestAwsIamPolicySpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AwsIamPolicySpec Validation Tests")
}

// s3ReadOnlyDocument is a representative permission document: the common
// "read objects from one bucket" grant.
func s3ReadOnlyDocument() *structpb.Struct {
	doc, err := structpb.NewStruct(map[string]interface{}{
		"Version": "2012-10-17",
		"Statement": []interface{}{
			map[string]interface{}{
				"Effect":   "Allow",
				"Action":   []interface{}{"s3:GetObject", "s3:ListBucket"},
				"Resource": []interface{}{"arn:aws:s3:::demo-bucket", "arn:aws:s3:::demo-bucket/*"},
			},
		},
	})
	if err != nil {
		panic(err)
	}
	return doc
}

// minimalValidPolicy is the common case: a region and a permission document.
func minimalValidPolicy() *AwsIamPolicy {
	return &AwsIamPolicy{
		ApiVersion: "aws.planton.dev/v1",
		Kind:       "AwsIamPolicy",
		Metadata: &shared.CloudResourceMetadata{
			Name: "s3-read-only",
		},
		Spec: &AwsIamPolicySpec{
			Region:         "us-west-2",
			PolicyDocument: s3ReadOnlyDocument(),
		},
	}
}

var _ = ginkgo.Describe("AwsIamPolicySpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("aws_iam_policy", func() {

			ginkgo.It("should not return a validation error for a minimal policy", func() {
				err := protovalidate.Validate(minimalValidPolicy())
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with a description and a path", func() {
				input := minimalValidPolicy()
				input.Spec.Description = "Read-only access to the demo bucket"
				input.Spec.Path = "/service-policies/"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for the root path", func() {
				input := minimalValidPolicy()
				input.Spec.Path = "/"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for a nested path", func() {
				input := minimalValidPolicy()
				input.Spec.Path = "/org/team-a/boundaries/"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with full metadata set", func() {
				input := minimalValidPolicy()
				input.Metadata = &shared.CloudResourceMetadata{
					Name:   "s3-read-only",
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
		ginkgo.Context("aws_iam_policy", func() {

			ginkgo.It("should return a validation error when api_version is wrong", func() {
				input := minimalValidPolicy()
				input.ApiVersion = "wrong.planton.dev/v1"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when kind is wrong", func() {
				input := minimalValidPolicy()
				input.Kind = "WrongKind"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when metadata is missing", func() {
				input := minimalValidPolicy()
				input.Metadata = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when spec is missing", func() {
				input := &AwsIamPolicy{
					ApiVersion: "aws.planton.dev/v1",
					Kind:       "AwsIamPolicy",
					Metadata:   &shared.CloudResourceMetadata{Name: "s3-read-only"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when region is empty", func() {
				input := minimalValidPolicy()
				input.Spec.Region = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when the policy document is missing", func() {
				input := minimalValidPolicy()
				input.Spec.PolicyDocument = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when description exceeds 1000 characters", func() {
				input := minimalValidPolicy()
				long := make([]byte, 1001)
				for i := range long {
					long[i] = 'a'
				}
				input.Spec.Description = string(long)
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when path does not begin with '/'", func() {
				input := minimalValidPolicy()
				input.Spec.Path = "service-policies/"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when path does not end with '/'", func() {
				input := minimalValidPolicy()
				input.Spec.Path = "/service-policies"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when path contains an empty segment", func() {
				input := minimalValidPolicy()
				input.Spec.Path = "/service//policies/"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})
	})
})
