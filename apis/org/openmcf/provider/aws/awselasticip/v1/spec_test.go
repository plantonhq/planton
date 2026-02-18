package awselasticipv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared"
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
			ApiVersion: "aws.openmcf.org/v1",
			Kind:       "AwsElasticIp",
			Metadata: &shared.CloudResourceMetadata{
				Name: "my-eip",
			},
			Spec: &AwsElasticIpSpec{Region: "us-east-1"},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts an EIP with only network_border_group set", func() {
		input := &AwsElasticIp{
			ApiVersion: "aws.openmcf.org/v1",
			Kind:       "AwsElasticIp",
			Metadata: &shared.CloudResourceMetadata{
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
			ApiVersion: "aws.openmcf.org/v1",
			Kind:       "AwsElasticIp",
			Metadata: &shared.CloudResourceMetadata{
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
			ApiVersion: "aws.openmcf.org/v1",
			Kind:       "AwsElasticIp",
			Metadata: &shared.CloudResourceMetadata{
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
			ApiVersion: "aws.openmcf.org/v1",
			Kind:       "AwsElasticIp",
			Metadata: &shared.CloudResourceMetadata{
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

	// -------------------------------------------------------------------------
	// CEL: address_requires_byoip_pool
	// -------------------------------------------------------------------------

	ginkgo.It("fails when address is set without public_ipv4_pool", func() {
		input := &AwsElasticIp{
			ApiVersion: "aws.openmcf.org/v1",
			Kind:       "AwsElasticIp",
			Metadata: &shared.CloudResourceMetadata{
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
	// api.proto: api_version and kind constants
	// -------------------------------------------------------------------------

	ginkgo.It("fails when api_version is wrong", func() {
		input := &AwsElasticIp{
			ApiVersion: "wrong.openmcf.org/v1",
			Kind:       "AwsElasticIp",
			Metadata: &shared.CloudResourceMetadata{
				Name: "test-eip",
			},
			Spec: &AwsElasticIpSpec{Region: "us-east-1"},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when kind is wrong", func() {
		input := &AwsElasticIp{
			ApiVersion: "aws.openmcf.org/v1",
			Kind:       "WrongKind",
			Metadata: &shared.CloudResourceMetadata{
				Name: "test-eip",
			},
			Spec: &AwsElasticIpSpec{Region: "us-east-1"},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when metadata is missing", func() {
		input := &AwsElasticIp{
			ApiVersion: "aws.openmcf.org/v1",
			Kind:       "AwsElasticIp",
			Spec:       &AwsElasticIpSpec{Region: "us-east-1"},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when spec is missing", func() {
		input := &AwsElasticIp{
			ApiVersion: "aws.openmcf.org/v1",
			Kind:       "AwsElasticIp",
			Metadata: &shared.CloudResourceMetadata{
				Name: "test-eip",
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})
})
