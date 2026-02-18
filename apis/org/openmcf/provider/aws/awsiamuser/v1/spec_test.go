package awsiamuserv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared"
)

func TestAwsIamUserSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AwsIamUserSpec Custom Validation Tests")
}

var _ = ginkgo.Describe("AwsIamUserSpec Custom Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("aws_iam_user", func() {

			ginkgo.It("should not return a validation error for minimal valid fields", func() {
				input := &AwsIamUser{
					ApiVersion: "aws.openmcf.org/v1",
					Kind:       "AwsIamUser",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-iam-user",
					},
					Spec: &AwsIamUserSpec{
						Region:   "us-west-2",
						UserName: "test-ci-user",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})
})
