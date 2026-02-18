package awslambdav1

import (
	"testing"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	foreignkeyv1 "github.com/plantonhq/openmcf/apis/org/openmcf/shared/foreignkey/v1"

	"buf.build/go/protovalidate"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared"
)

func TestAwsLambdaSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AwsLambdaSpec Custom Validation Tests")
}

var _ = ginkgo.Describe("AwsLambdaSpec Custom Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("aws_lambda", func() {

			ginkgo.It("should not return a validation error for minimal valid fields", func() {
				input := &AwsLambda{
					ApiVersion: "aws.openmcf.org/v1",
					Kind:       "AwsLambda",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-lambda-function",
					},
					Spec: &AwsLambdaSpec{
						Region:       "us-west-2",
						FunctionName: "test-function",
						RoleArn: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "arn:aws:iam::123456789012:role/lambda-execution-role"},
						},
						CodeSourceType: CodeSourceType_CODE_SOURCE_TYPE_S3,
						S3: &S3Code{
							Bucket: "my-lambda-bucket",
							Key:    "functions/test-function.zip",
						},
						Runtime: "python3.11",
						Handler: "main.handler",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})
})
