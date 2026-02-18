package awsrdsinstancev1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared"
	foreignkeyv1 "github.com/plantonhq/openmcf/apis/org/openmcf/shared/foreignkey/v1"
)

func TestAwsRdsInstanceSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AwsRdsInstanceSpec Custom Validation Tests")
}

var _ = ginkgo.Describe("AwsRdsInstanceSpec Custom Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("aws_rds_instance", func() {

			ginkgo.It("should not return a validation error for minimal valid fields", func() {
				input := &AwsRdsInstance{
					ApiVersion: "aws.openmcf.org/v1",
					Kind:       "AwsRdsInstance",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-rds-instance",
					},
					Spec: &AwsRdsInstanceSpec{
						Region: "us-west-2",
						SubnetIds: []*foreignkeyv1.StringValueOrRef{
							{
								LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "subnet-12345678"},
							},
							{
								LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "subnet-87654321"},
							},
						},
						Engine:             "postgres",
						EngineVersion:      "14.10",
						InstanceClass:      "db.t3.micro",
						AllocatedStorageGb: 20,
						Username:           "dbmaster",
						Password:           "mypassword123",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})
})
