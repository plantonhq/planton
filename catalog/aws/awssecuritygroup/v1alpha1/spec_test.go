package awssecuritygroupv1alpha1

import (
	"testing"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"

	"buf.build/go/protovalidate"
	"github.com/plantonhq/planton/shared"
)

func TestAwsSecurityGroupSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AwsSecurityGroupSpec Custom Validation Tests")
}

func validMinimalSpec() *AwsSecurityGroup {
	return &AwsSecurityGroup{
		ApiVersion: "aws.planton.dev/v1alpha1",
		Kind:       "AwsSecurityGroup",
		Metadata: &shared.CloudResourceMetadata{
			Name: "test-security-group",
		},
		Spec: &AwsSecurityGroupSpec{
			Region: "us-west-2",
			VpcId: &foreignkeyv1.StringValueOrRef{
				LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "vpc-12345678"},
			},
			Description: "Test security group for validation",
		},
	}
}

var _ = ginkgo.Describe("AwsSecurityGroupSpec Custom Validation Tests", func() {

	// ===== HAPPY PATH TESTS =====

	ginkgo.Describe("When valid input is passed", func() {

		ginkgo.It("should not return a validation error for minimal valid fields", func() {
			input := validMinimalSpec()
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept a group shared into additional VPCs", func() {
			input := validMinimalSpec()
			input.Spec.AdditionalVpcIds = []*foreignkeyv1.StringValueOrRef{
				{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "vpc-87654321"}},
				{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "vpc-abcdef01"}},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept a rules-rich group (self-reference, prefix lists, dual-stack CIDRs)", func() {
			input := validMinimalSpec()
			input.Spec.RevokeRulesOnDelete = true
			input.Spec.Ingress = []*SecurityGroupRule{
				{
					Protocol:      "tcp",
					FromPort:      8080,
					ToPort:        8080,
					SelfReference: true,
					Description:   "service-to-service",
				},
				{
					Protocol:  "tcp",
					FromPort:  443,
					ToPort:    443,
					Ipv4Cidrs: []string{"10.0.0.0/16"},
					Ipv6Cidrs: []string{"::/0"},
				},
			}
			input.Spec.Egress = []*SecurityGroupRule{
				{
					Protocol:      "tcp",
					FromPort:      443,
					ToPort:        443,
					PrefixListIds: []string{"pl-68a54001"},
					Description:   "HTTPS to S3 via the gateway prefix list",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept ICMP type/code overloading with the -1 all-codes sentinel", func() {
			input := validMinimalSpec()
			input.Spec.Ingress = []*SecurityGroupRule{
				{
					Protocol:  "icmp",
					FromPort:  8,
					ToPort:    -1,
					Ipv4Cidrs: []string{"10.0.0.0/16"},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept an all-protocol rule with both ports zero", func() {
			input := validMinimalSpec()
			input.Spec.Egress = []*SecurityGroupRule{
				{
					Protocol:  "-1",
					FromPort:  0,
					ToPort:    0,
					Ipv4Cidrs: []string{"0.0.0.0/0"},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept descriptions using every character AWS allows", func() {
			input := validMinimalSpec()
			allowed := "Web tier 0-9 a_z A.Z :/()#,@[]+=&;{}!$*"
			input.Spec.Description = allowed
			input.Spec.Ingress = []*SecurityGroupRule{
				{
					Protocol:    "tcp",
					FromPort:    443,
					ToPort:      443,
					Ipv4Cidrs:   []string{"10.0.0.0/16"},
					Description: allowed,
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept rules referencing other security groups", func() {
			input := validMinimalSpec()
			input.Spec.Ingress = []*SecurityGroupRule{
				{
					Protocol: "tcp",
					FromPort: 5432,
					ToPort:   5432,
					SourceSecurityGroupIds: []*foreignkeyv1.StringValueOrRef{
						{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "sg-app-tier"}},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})
	})

	// ===== FAILURE TESTS =====

	ginkgo.Describe("When invalid input is passed", func() {

		ginkgo.It("should fail when an all-protocol rule carries a port range", func() {
			input := validMinimalSpec()
			input.Spec.Ingress = []*SecurityGroupRule{
				{
					Protocol:  "-1",
					FromPort:  80,
					ToPort:    80,
					Ipv4Cidrs: []string{"0.0.0.0/0"},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("from_port and to_port must both be 0"))
		})

		ginkgo.It("should fail when a port is below -1", func() {
			input := validMinimalSpec()
			input.Spec.Ingress = []*SecurityGroupRule{
				{Protocol: "tcp", FromPort: -2, ToPort: 80},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when a port exceeds 65535", func() {
			input := validMinimalSpec()
			input.Spec.Ingress = []*SecurityGroupRule{
				{Protocol: "tcp", FromPort: 80, ToPort: 65536},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail for a malformed prefix list ID", func() {
			input := validMinimalSpec()
			input.Spec.Egress = []*SecurityGroupRule{
				{
					Protocol:      "tcp",
					FromPort:      443,
					ToPort:        443,
					PrefixListIds: []string{"not-a-prefix-list"},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when an ingress rule carries the egress-direction group field", func() {
			input := validMinimalSpec()
			input.Spec.Ingress = []*SecurityGroupRule{
				{
					Protocol: "tcp",
					FromPort: 443,
					ToPort:   443,
					DestinationSecurityGroupIds: []*foreignkeyv1.StringValueOrRef{
						{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "sg-dead-config"}},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when an egress rule carries the ingress-direction group field", func() {
			input := validMinimalSpec()
			input.Spec.Egress = []*SecurityGroupRule{
				{
					Protocol: "tcp",
					FromPort: 443,
					ToPort:   443,
					SourceSecurityGroupIds: []*foreignkeyv1.StringValueOrRef{
						{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "sg-dead-config"}},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when description exceeds 255 characters", func() {
			input := validMinimalSpec()
			long := make([]byte, 256)
			for i := range long {
				long[i] = 'a'
			}
			input.Spec.Description = string(long)
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		// AWS refuses the create call for any other character, so an
		// apostrophe that passes validation fails only at apply.
		ginkgo.It("should fail when the group description carries a character AWS refuses", func() {
			input := validMinimalSpec()
			input.Spec.Description = "Team's web tier"
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("in a security group description"))
		})

		ginkgo.It("should fail when a rule description carries a character AWS refuses", func() {
			input := validMinimalSpec()
			input.Spec.Ingress = []*SecurityGroupRule{
				{
					Protocol:    "tcp",
					FromPort:    443,
					ToPort:      443,
					Ipv4Cidrs:   []string{"10.0.0.0/16"},
					Description: "Team's HTTPS",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("in a security group rule description"))
		})

		ginkgo.It("should fail when vpc_id is missing", func() {
			input := validMinimalSpec()
			input.Spec.VpcId = nil
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when description is missing", func() {
			input := validMinimalSpec()
			input.Spec.Description = ""
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})
	})
})
