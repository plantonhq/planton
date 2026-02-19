package alicloudvpcv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared"
)

func TestAlicloudVpcSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AlicloudVpcSpec Validation Tests")
}

var _ = ginkgo.Describe("AlicloudVpcSpec Validation Tests", func() {

	ginkgo.Describe("valid input", func() {

		ginkgo.It("should pass with minimal required fields", func() {
			input := &AlicloudVpc{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudVpc",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test-vpc",
				},
				Spec: &AlicloudVpcSpec{
					Region:    "cn-hangzhou",
					VpcName:   "my-vpc",
					CidrBlock: "10.0.0.0/8",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with all optional fields populated", func() {
			input := &AlicloudVpc{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudVpc",
				Metadata: &shared.CloudResourceMetadata{
					Name: "full-config-vpc",
					Org:  "my-org",
					Env:  "production",
				},
				Spec: &AlicloudVpcSpec{
					Region:          "cn-shanghai",
					VpcName:         "prod-vpc",
					CidrBlock:       "172.16.0.0/12",
					Description:     "Production VPC for the platform",
					EnableIpv6:      true,
					ResourceGroupId: "rg-abc123",
					Tags:            map[string]string{"team": "platform", "cost-center": "eng"},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with 192.168 CIDR range", func() {
			input := &AlicloudVpc{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudVpc",
				Metadata: &shared.CloudResourceMetadata{
					Name: "small-vpc",
				},
				Spec: &AlicloudVpcSpec{
					Region:    "ap-southeast-1",
					VpcName:   "dev-vpc",
					CidrBlock: "192.168.0.0/16",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with IPv6 enabled", func() {
			input := &AlicloudVpc{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudVpc",
				Metadata: &shared.CloudResourceMetadata{
					Name: "ipv6-vpc",
				},
				Spec: &AlicloudVpcSpec{
					Region:     "us-west-1",
					VpcName:    "ipv6-enabled-vpc",
					CidrBlock:  "10.0.0.0/16",
					EnableIpv6: true,
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("invalid input", func() {

		ginkgo.It("should fail when region is missing", func() {
			input := &AlicloudVpc{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudVpc",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test",
				},
				Spec: &AlicloudVpcSpec{
					VpcName:   "my-vpc",
					CidrBlock: "10.0.0.0/8",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when vpc_name is missing", func() {
			input := &AlicloudVpc{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudVpc",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test",
				},
				Spec: &AlicloudVpcSpec{
					Region:    "cn-hangzhou",
					CidrBlock: "10.0.0.0/8",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when cidr_block is missing", func() {
			input := &AlicloudVpc{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudVpc",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test",
				},
				Spec: &AlicloudVpcSpec{
					Region:  "cn-hangzhou",
					VpcName: "my-vpc",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when api_version is wrong", func() {
			input := &AlicloudVpc{
				ApiVersion: "wrong/v1",
				Kind:       "AlicloudVpc",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test",
				},
				Spec: &AlicloudVpcSpec{
					Region:    "cn-hangzhou",
					VpcName:   "my-vpc",
					CidrBlock: "10.0.0.0/8",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when kind is wrong", func() {
			input := &AlicloudVpc{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "WrongKind",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test",
				},
				Spec: &AlicloudVpcSpec{
					Region:    "cn-hangzhou",
					VpcName:   "my-vpc",
					CidrBlock: "10.0.0.0/8",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when metadata is missing", func() {
			input := &AlicloudVpc{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudVpc",
				Spec: &AlicloudVpcSpec{
					Region:    "cn-hangzhou",
					VpcName:   "my-vpc",
					CidrBlock: "10.0.0.0/8",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when vpc_name exceeds max length", func() {
			longName := ""
			for i := 0; i < 129; i++ {
				longName += "a"
			}
			input := &AlicloudVpc{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudVpc",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test",
				},
				Spec: &AlicloudVpcSpec{
					Region:    "cn-hangzhou",
					VpcName:   longName,
					CidrBlock: "10.0.0.0/8",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})
	})
})
