package awsiamrolev1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/plantonhq/openmcf/apis/org/openmcf/shared"
)

func TestAwsIamRole(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AwsIamRole Suite")
}

var _ = ginkgo.Describe("AwsIamRole Custom Validation Tests", func() {
	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("aws_iam_role", func() {
			ginkgo.It("should not return a validation error", func() {
				input := &AwsIamRole{
					ApiVersion: "aws.openmcf.org/v1",
					Kind:       "AwsIamRole",
					Metadata: &shared.CloudResourceMetadata{
						Name: "valid-name",
					},
					Spec: &AwsIamRoleSpec{
						Region:      "us-west-2",
						TrustPolicy: &structpb.Struct{},
						ManagedPolicyArns: []string{
							"arn:aws:iam::123456789012:policy/testPolicy",
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})
})
