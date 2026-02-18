package awseksnodegroupv1

import (
	"testing"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	foreignkeyv1 "github.com/plantonhq/openmcf/apis/org/openmcf/shared/foreignkey/v1"

	"buf.build/go/protovalidate"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared"
)

func TestAwsEksNodeGroupSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AwsEksNodeGroupSpec Custom Validation Tests")
}

var _ = ginkgo.Describe("AwsEksNodeGroupSpec Custom Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("aws_eks_node_group", func() {

			ginkgo.It("should not return a validation error for minimal valid fields", func() {
				input := &AwsEksNodeGroup{
					ApiVersion: "aws.openmcf.org/v1",
					Kind:       "AwsEksNodeGroup",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-node-group",
					},
					Spec: &AwsEksNodeGroupSpec{
						Region: "us-west-2",
						ClusterName: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "test-eks-cluster"},
						},
						NodeRoleArn: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "arn:aws:iam::123456789012:role/EksNodeRole"},
						},
						SubnetIds: []*foreignkeyv1.StringValueOrRef{
							{
								LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "subnet-12345678"},
							},
							{
								LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "subnet-87654321"},
							},
						},
						InstanceType: "t3.small",
						Scaling: &AwsEksNodeGroupScalingConfig{
							MinSize:     1,
							MaxSize:     3,
							DesiredSize: 2,
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})
})
