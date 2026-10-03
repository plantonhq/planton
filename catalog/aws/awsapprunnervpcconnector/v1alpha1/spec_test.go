package awsapprunnervpcconnectorv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestAwsAppRunnerVpcConnectorSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AwsAppRunnerVpcConnector Validation Suite")
}

// helper to create a StringValueOrRef with a literal value.
func strRef(val string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: val},
	}
}

func validEnvelope(spec *AwsAppRunnerVpcConnectorSpec) *AwsAppRunnerVpcConnector {
	return &AwsAppRunnerVpcConnector{
		ApiVersion: "aws.planton.dev/v1alpha1",
		Kind:       "AwsAppRunnerVpcConnector",
		Metadata:   &shared.CatalogObjectMetadata{Name: "test-connector"},
		Spec:       spec,
	}
}

var _ = ginkgo.Describe("AwsAppRunnerVpcConnectorSpec validations", func() {
	var spec *AwsAppRunnerVpcConnectorSpec

	ginkgo.BeforeEach(func() {
		// Minimal valid spec: region + one subnet + one security group.
		spec = &AwsAppRunnerVpcConnectorSpec{
			Region: "us-west-2",
			SubnetIds: []*foreignkeyv1.StringValueOrRef{
				strRef("subnet-0abc123"),
			},
			SecurityGroupIds: []*foreignkeyv1.StringValueOrRef{
				strRef("sg-0abc123"),
			},
		}
	})

	// -------------------------------------------------------------------------
	// Happy path
	// -------------------------------------------------------------------------

	ginkgo.It("accepts a minimal spec (region + one subnet + one security group)", func() {
		err := protovalidate.Validate(validEnvelope(spec))
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts a two-AZ subnet spread with multiple security groups", func() {
		spec.SubnetIds = []*foreignkeyv1.StringValueOrRef{
			strRef("subnet-0abc123"),
			strRef("subnet-0def456"),
		}
		spec.SecurityGroupIds = []*foreignkeyv1.StringValueOrRef{
			strRef("sg-0abc123"),
			strRef("sg-0def456"),
		}
		err := protovalidate.Validate(validEnvelope(spec))
		gomega.Expect(err).To(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// Required field validations
	// -------------------------------------------------------------------------

	ginkgo.It("fails when region is empty", func() {
		spec.Region = ""
		err := protovalidate.Validate(validEnvelope(spec))
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when subnet_ids is empty", func() {
		spec.SubnetIds = nil
		err := protovalidate.Validate(validEnvelope(spec))
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when security_group_ids is empty", func() {
		spec.SecurityGroupIds = nil
		err := protovalidate.Validate(validEnvelope(spec))
		gomega.Expect(err).NotTo(gomega.BeNil())
	})
})
