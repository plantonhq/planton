package awselasticipv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestAwsElasticIpSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AwsElasticIpSpec Validation Suite")
}

var _ = ginkgo.Describe("AwsElasticIpSpec validations", func() {

	// -------------------------------------------------------------------------
	// Happy path
	// -------------------------------------------------------------------------

	ginkgo.It("accepts a minimal EIP with no spec fields (the 95% use case)", func() {
		input := &AwsElasticIp{
			ApiVersion: "aws.planton.dev/v1alpha1",
			Kind:       "AwsElasticIp",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "my-eip",
			},
			Spec: &AwsElasticIpSpec{Region: "us-east-1"},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts an EIP with only network_border_group set", func() {
		input := &AwsElasticIp{
			ApiVersion: "aws.planton.dev/v1alpha1",
			Kind:       "AwsElasticIp",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "wavelength-eip",
			},
			Spec: &AwsElasticIpSpec{
				Region:             "us-east-1",
				NetworkBorderGroup: "us-east-1-wl1-bos-wlz-1",
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts an EIP with BYOIP pool only (no specific address)", func() {
		input := &AwsElasticIp{
			ApiVersion: "aws.planton.dev/v1alpha1",
			Kind:       "AwsElasticIp",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "byoip-eip",
			},
			Spec: &AwsElasticIpSpec{
				Region:         "us-east-1",
				PublicIpv4Pool: "ipv4pool-ec2-0123456789abcdef0",
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts an EIP with BYOIP pool and specific address", func() {
		input := &AwsElasticIp{
			ApiVersion: "aws.planton.dev/v1alpha1",
			Kind:       "AwsElasticIp",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "byoip-specific-eip",
			},
			Spec: &AwsElasticIpSpec{
				Region:         "us-east-1",
				PublicIpv4Pool: "ipv4pool-ec2-0123456789abcdef0",
				Address:        "198.51.100.10",
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts an EIP with all optional fields set", func() {
		input := &AwsElasticIp{
			ApiVersion: "aws.planton.dev/v1alpha1",
			Kind:       "AwsElasticIp",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "full-config-eip",
			},
			Spec: &AwsElasticIpSpec{
				Region:             "us-east-1",
				PublicIpv4Pool:     "ipv4pool-ec2-0123456789abcdef0",
				Address:            "198.51.100.10",
				NetworkBorderGroup: "us-east-1",
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts an EIP allocated from an IPAM pool", func() {
		input := &AwsElasticIp{
			ApiVersion: "aws.planton.dev/v1alpha1",
			Kind:       "AwsElasticIp",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "ipam-eip",
			},
			Spec: &AwsElasticIpSpec{
				Region:     "us-east-1",
				IpamPoolId: "ipam-pool-07ccc86aa41bef7ce",
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts a specific address recovered from an IPAM pool (no BYOIP pool)", func() {
		input := &AwsElasticIp{
			ApiVersion: "aws.planton.dev/v1alpha1",
			Kind:       "AwsElasticIp",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "ipam-recover-eip",
			},
			Spec: &AwsElasticIpSpec{
				Region:     "us-east-1",
				IpamPoolId: "ipam-pool-07ccc86aa41bef7ce",
				Address:    "198.51.100.10",
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts an EIP associated with an instance", func() {
		input := &AwsElasticIp{
			ApiVersion: "aws.planton.dev/v1alpha1",
			Kind:       "AwsElasticIp",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "instance-eip",
			},
			Spec: &AwsElasticIpSpec{
				Region: "us-east-1",
				Instance: &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "i-0123456789abcdef0"},
				},
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts an EIP associated with an ENI on a specific private IP", func() {
		input := &AwsElasticIp{
			ApiVersion: "aws.planton.dev/v1alpha1",
			Kind:       "AwsElasticIp",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "eni-eip",
			},
			Spec: &AwsElasticIpSpec{
				Region: "us-east-1",
				NetworkInterface: &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "eni-0123456789abcdef0"},
				},
				AssociateWithPrivateIp: "10.0.0.12",
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts a reverse DNS domain name (with and without trailing dot)", func() {
		for _, domain := range []string{"mail.example.com", "mail.example.com."} {
			input := &AwsElasticIp{
				ApiVersion: "aws.planton.dev/v1alpha1",
				Kind:       "AwsElasticIp",
				Metadata: &shared.CatalogObjectMetadata{
					Name: "rdns-eip",
				},
				Spec: &AwsElasticIpSpec{
					Region:               "us-east-1",
					ReverseDnsDomainName: domain,
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil(), "domain %q should be accepted", domain)
		}
	})

	// -------------------------------------------------------------------------
	// CEL: address_requires_pool
	// -------------------------------------------------------------------------

	ginkgo.It("fails when address is set without any pool", func() {
		input := &AwsElasticIp{
			ApiVersion: "aws.planton.dev/v1alpha1",
			Kind:       "AwsElasticIp",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "bad-address-eip",
			},
			Spec: &AwsElasticIpSpec{
				Region:  "us-east-1",
				Address: "198.51.100.10",
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// CEL: at_most_one_association_target / private_ip_requires_association_target
	// -------------------------------------------------------------------------

	ginkgo.It("fails when both instance and network_interface are set", func() {
		input := &AwsElasticIp{
			ApiVersion: "aws.planton.dev/v1alpha1",
			Kind:       "AwsElasticIp",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "double-target-eip",
			},
			Spec: &AwsElasticIpSpec{
				Region: "us-east-1",
				Instance: &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "i-0123456789abcdef0"},
				},
				NetworkInterface: &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "eni-0123456789abcdef0"},
				},
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when associate_with_private_ip is set without a target", func() {
		input := &AwsElasticIp{
			ApiVersion: "aws.planton.dev/v1alpha1",
			Kind:       "AwsElasticIp",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "dangling-private-ip-eip",
			},
			Spec: &AwsElasticIpSpec{
				Region:                 "us-east-1",
				AssociateWithPrivateIp: "10.0.0.12",
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when associate_with_private_ip is not a valid IPv4 address", func() {
		input := &AwsElasticIp{
			ApiVersion: "aws.planton.dev/v1alpha1",
			Kind:       "AwsElasticIp",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "bad-private-ip-eip",
			},
			Spec: &AwsElasticIpSpec{
				Region: "us-east-1",
				Instance: &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "i-0123456789abcdef0"},
				},
				AssociateWithPrivateIp: "not-an-ip",
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// reverse_dns_domain_name format
	// -------------------------------------------------------------------------

	ginkgo.It("fails when reverse_dns_domain_name is not a plausible FQDN", func() {
		for _, domain := range []string{"not a domain", "-bad.example.com", "example"} {
			input := &AwsElasticIp{
				ApiVersion: "aws.planton.dev/v1alpha1",
				Kind:       "AwsElasticIp",
				Metadata: &shared.CatalogObjectMetadata{
					Name: "bad-rdns-eip",
				},
				Spec: &AwsElasticIpSpec{
					Region:               "us-east-1",
					ReverseDnsDomainName: domain,
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil(), "domain %q should be rejected", domain)
		}
	})

	// -------------------------------------------------------------------------
	// api.proto: api_version and kind constants
	// -------------------------------------------------------------------------

	ginkgo.It("fails when api_version is wrong", func() {
		input := &AwsElasticIp{
			ApiVersion: "wrong.planton.dev/v1",
			Kind:       "AwsElasticIp",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "test-eip",
			},
			Spec: &AwsElasticIpSpec{Region: "us-east-1"},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when kind is wrong", func() {
		input := &AwsElasticIp{
			ApiVersion: "aws.planton.dev/v1alpha1",
			Kind:       "WrongKind",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "test-eip",
			},
			Spec: &AwsElasticIpSpec{Region: "us-east-1"},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when metadata is missing", func() {
		input := &AwsElasticIp{
			ApiVersion: "aws.planton.dev/v1alpha1",
			Kind:       "AwsElasticIp",
			Spec:       &AwsElasticIpSpec{Region: "us-east-1"},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when spec is missing", func() {
		input := &AwsElasticIp{
			ApiVersion: "aws.planton.dev/v1alpha1",
			Kind:       "AwsElasticIp",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "test-eip",
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})
})
