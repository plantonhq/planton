package awsrdsclusterv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/apis/dev/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/apis/dev/planton/shared/foreignkey/v1"
)

func TestAwsRdsClusterSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AwsRdsClusterSpec Custom Validation Tests")
}

var _ = ginkgo.Describe("AwsRdsClusterSpec Custom Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("aws_rds_cluster", func() {

			ginkgo.It("should not return a validation error for minimal valid fields", func() {
				input := &AwsRdsCluster{
					ApiVersion: "aws.planton.dev/v1",
					Kind:       "AwsRdsCluster",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-rds-cluster",
					},
					Spec: &AwsRdsClusterSpec{
						Region: "us-west-2",
						SubnetIds: []*foreignkeyv1.StringValueOrRef{
							{
								LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "subnet-12345678"},
							},
							{
								LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "subnet-87654321"},
							},
						},
						Engine:                     "aurora-mysql",
						EngineVersion:              "8.0.mysql_aurora.3.05.2",
						SkipFinalSnapshot:          true,                  // Skip final snapshot to avoid requiring final_snapshot_identifier
						PreferredMaintenanceWindow: "mon:03:00-mon:04:00", // Valid format
						PreferredBackupWindow:      "05:00-06:00",         // Valid backup window format
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})
})
