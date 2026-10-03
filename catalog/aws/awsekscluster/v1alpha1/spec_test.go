package awseksclusterv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestAwsEksClusterSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AwsEksClusterSpec Validation Tests")
}

func literalRef(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

// minimalValidCluster is the common case: a two-AZ control plane with a
// dedicated cluster role.
func minimalValidCluster() *AwsEksCluster {
	return &AwsEksCluster{
		ApiVersion: "aws.planton.dev/v1alpha1",
		Kind:       "AwsEksCluster",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "platform",
		},
		Spec: &AwsEksClusterSpec{
			Region: "us-west-2",
			SubnetIds: []*foreignkeyv1.StringValueOrRef{
				literalRef("subnet-0123456789abcdef0"),
				literalRef("subnet-0123456789abcdef1"),
			},
			ClusterRoleArn: literalRef("arn:aws:iam::123456789012:role/EksClusterServiceRole"),
		},
	}
}

var _ = ginkgo.Describe("AwsEksClusterSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("aws_eks_cluster", func() {

			ginkgo.It("should not return a validation error for a minimal cluster", func() {
				err := protovalidate.Validate(minimalValidCluster())
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a fully configured hardened cluster", func() {
				endpointPublic := false
				bootstrapAddons := false
				creatorAdmin := false
				input := minimalValidCluster()
				input.Spec.Version = "1.31"
				input.Spec.SecurityGroupIds = []*foreignkeyv1.StringValueOrRef{
					literalRef("sg-0123456789abcdef0"),
				}
				input.Spec.EndpointPublicAccess = &endpointPublic
				input.Spec.EndpointPrivateAccess = true
				input.Spec.ControlPlaneEgressMode = "CUSTOMER_ROUTED"
				input.Spec.IpFamily = "ipv4"
				input.Spec.ServiceIpv4Cidr = "172.20.0.0/16"
				input.Spec.EnabledClusterLogTypes = []string{"api", "audit", "authenticator"}
				input.Spec.KmsKeyArn = literalRef("arn:aws:kms:us-west-2:123456789012:key/1234abcd-12ab-34cd-56ef-1234567890ab")
				input.Spec.AccessConfig = &AwsEksClusterAccessConfig{
					AuthenticationMode:                      "API",
					BootstrapClusterCreatorAdminPermissions: &creatorAdmin,
				}
				input.Spec.UpgradeSupportType = "STANDARD"
				input.Spec.ZonalShiftEnabled = true
				input.Spec.DeletionProtection = true
				input.Spec.BootstrapSelfManagedAddons = &bootstrapAddons
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept an Auto Mode cluster with built-in pools", func() {
				input := minimalValidCluster()
				input.Spec.AutoMode = &AwsEksClusterAutoMode{
					Enabled:     true,
					NodePools:   []string{"general-purpose", "system"},
					NodeRoleArn: literalRef("arn:aws:iam::123456789012:role/EksAutoNodeRole"),
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a future Kubernetes minor version", func() {
				input := minimalValidCluster()
				input.Spec.Version = "1.40"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept public access CIDRs", func() {
				input := minimalValidCluster()
				input.Spec.PublicAccessCidrs = []string{"203.0.113.0/24", "198.51.100.7/32"}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a provisioned control-plane scaling tier", func() {
				input := minimalValidCluster()
				input.Spec.ControlPlaneScalingTier = "tier-2xl"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept hybrid-node remote networks in every private range", func() {
				input := minimalValidCluster()
				input.Spec.RemoteNetworks = &AwsEksClusterRemoteNetworks{
					NodeCidrs: []string{"10.80.0.0/16", "172.16.10.0/24", "192.168.4.0/22"},
					PodCidrs:  []string{"100.64.10.0/24"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept node-only remote networks", func() {
				input := minimalValidCluster()
				input.Spec.RemoteNetworks = &AwsEksClusterRemoteNetworks{
					NodeCidrs: []string{"10.80.0.0/16"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("aws_eks_cluster", func() {

			ginkgo.It("should reject a missing cluster role", func() {
				input := minimalValidCluster()
				input.Spec.ClusterRoleArn = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject fewer than two subnets", func() {
				input := minimalValidCluster()
				input.Spec.SubnetIds = input.Spec.SubnetIds[:1]
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a pre-1.24 Kubernetes version", func() {
				input := minimalValidCluster()
				input.Spec.Version = "1.23"
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a malformed Kubernetes version", func() {
				input := minimalValidCluster()
				input.Spec.Version = "v1.31"
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject an invalid public access CIDR", func() {
				input := minimalValidCluster()
				input.Spec.PublicAccessCidrs = []string{"not-a-cidr"}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject an invalid control plane egress mode", func() {
				input := minimalValidCluster()
				input.Spec.ControlPlaneEgressMode = "OPEN"
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject an invalid ip family", func() {
				input := minimalValidCluster()
				input.Spec.IpFamily = "dualstack"
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a service CIDR outside the /12-/24 range", func() {
				input := minimalValidCluster()
				input.Spec.ServiceIpv4Cidr = "10.0.0.0/8"
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a service CIDR on an ipv6 cluster", func() {
				input := minimalValidCluster()
				input.Spec.IpFamily = "ipv6"
				input.Spec.ServiceIpv4Cidr = "172.20.0.0/16"
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject an unknown control plane log type", func() {
				input := minimalValidCluster()
				input.Spec.EnabledClusterLogTypes = []string{"api", "kubelet"}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject an invalid upgrade support type", func() {
				input := minimalValidCluster()
				input.Spec.UpgradeSupportType = "LTS"
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject an invalid authentication mode", func() {
				input := minimalValidCluster()
				input.Spec.AccessConfig = &AwsEksClusterAccessConfig{AuthenticationMode: "IAM"}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject an unknown Auto Mode node pool", func() {
				input := minimalValidCluster()
				input.Spec.AutoMode = &AwsEksClusterAutoMode{
					Enabled:     true,
					NodePools:   []string{"gpu"},
					NodeRoleArn: literalRef("arn:aws:iam::123456789012:role/EksAutoNodeRole"),
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject Auto Mode node pools without the mode enabled", func() {
				input := minimalValidCluster()
				input.Spec.AutoMode = &AwsEksClusterAutoMode{
					NodePools:   []string{"general-purpose"},
					NodeRoleArn: literalRef("arn:aws:iam::123456789012:role/EksAutoNodeRole"),
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject Auto Mode node pools without a node role", func() {
				input := minimalValidCluster()
				input.Spec.AutoMode = &AwsEksClusterAutoMode{
					Enabled:   true,
					NodePools: []string{"general-purpose"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject an unknown control-plane scaling tier", func() {
				input := minimalValidCluster()
				input.Spec.ControlPlaneScalingTier = "tier-16xl"
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject an empty remote-networks block", func() {
				input := minimalValidCluster()
				input.Spec.RemoteNetworks = &AwsEksClusterRemoteNetworks{}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a malformed remote node CIDR", func() {
				input := minimalValidCluster()
				input.Spec.RemoteNetworks = &AwsEksClusterRemoteNetworks{
					NodeCidrs: []string{"not-a-cidr"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a public-range remote node CIDR", func() {
				input := minimalValidCluster()
				input.Spec.RemoteNetworks = &AwsEksClusterRemoteNetworks{
					NodeCidrs: []string{"203.0.113.0/24"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a public-range remote pod CIDR", func() {
				input := minimalValidCluster()
				input.Spec.RemoteNetworks = &AwsEksClusterRemoteNetworks{
					NodeCidrs: []string{"10.80.0.0/16"},
					PodCidrs:  []string{"198.51.100.0/24"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})
	})
})
