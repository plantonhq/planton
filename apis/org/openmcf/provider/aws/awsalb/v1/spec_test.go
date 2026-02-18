package awsalbv1

import (
	"testing"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	foreignkeyv1 "github.com/plantonhq/openmcf/apis/org/openmcf/shared/foreignkey/v1"

	"buf.build/go/protovalidate"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared"
)

func TestAwsAlbSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AwsAlbSpec Custom Validation Tests")
}

var _ = ginkgo.Describe("AwsAlbSpec Custom Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("aws_alb", func() {

			ginkgo.It("should not return a validation error for minimal valid fields", func() {
				input := &AwsAlb{
					ApiVersion: "aws.openmcf.org/v1",
					Kind:       "AwsAlb",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-alb-resource",
					},
					Spec: &AwsAlbSpec{
						Region: "us-west-2",
						Subnets: []*foreignkeyv1.StringValueOrRef{
							{
								LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "subnet-12345678"},
							},
							{
								LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "subnet-12345679"},
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			// Removed: SSL/certificate_arn test cases,
			// since we've temporarily removed that validation rule.
		})
	})

	// Removed: The test block that checks for an error if SSL is enabled but certificate ARN is missing,
	// because that validation no longer exists (we'll re-add these tests later).
})
