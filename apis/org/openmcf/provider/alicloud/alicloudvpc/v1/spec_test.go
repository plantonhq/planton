package alicloudvpcv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared"
)

func TestAliCloudVpcSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AliCloudVpcSpec Validation Tests")
}

var _ = ginkgo.Describe("AliCloudVpcSpec Validation Tests", func() {

	ginkgo.Describe("valid input", func() {

		ginkgo.It("should pass with minimal required fields", func() {
			input := &AliCloudVpc{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AliCloudVpc",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test-vpc",
				},
				Spec: &AliCloudVpcSpec{
					Region:    "cn-hangzhou",
					VpcName:   "my-vpc",
					CidrBlock: "10.0.0.0/8",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with all optional fields populated", func() {
			input := &AliCloudVpc{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AliCloudVpc",
				Metadata: &shared.CloudResourceMetadata{
					Name: "full-config-vpc",
					Org:  "my-org",
					Env:  "production",
				},
				Spec: &AliCloudVpcSpec{
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
			input := &AliCloudVpc{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AliCloudVpc",
				Metadata: &shared.CloudResourceMetadata{
					Name: "small-vpc",
				},
				Spec: &AliCloudVpcSpec{
					Region:    "ap-southeast-1",
					VpcName:   "dev-vpc",
					CidrBlock: "192.168.0.0/16",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with IPv6 enabled", func() {
			input := &AliCloudVpc{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AliCloudVpc",
				Metadata: &shared.CloudResourceMetadata{
					Name: "ipv6-vpc",
				},
				Spec: &AliCloudVpcSpec{
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
			input := &AliCloudVpc{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AliCloudVpc",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test",
				},
				Spec: &AliCloudVpcSpec{
					VpcName:   "my-vpc",
					CidrBlock: "10.0.0.0/8",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when vpc_name is missing", func() {
			input := &AliCloudVpc{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AliCloudVpc",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test",
				},
				Spec: &AliCloudVpcSpec{
					Region:    "cn-hangzhou",
					CidrBlock: "10.0.0.0/8",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when cidr_block is missing", func() {
			input := &AliCloudVpc{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AliCloudVpc",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test",
				},
				Spec: &AliCloudVpcSpec{
					Region:  "cn-hangzhou",
					VpcName: "my-vpc",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when api_version is wrong", func() {
			input := &AliCloudVpc{
				ApiVersion: "wrong/v1",
				Kind:       "AliCloudVpc",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test",
				},
				Spec: &AliCloudVpcSpec{
					Region:    "cn-hangzhou",
					VpcName:   "my-vpc",
					CidrBlock: "10.0.0.0/8",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when kind is wrong", func() {
			input := &AliCloudVpc{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "WrongKind",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test",
				},
				Spec: &AliCloudVpcSpec{
					Region:    "cn-hangzhou",
					VpcName:   "my-vpc",
					CidrBlock: "10.0.0.0/8",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when metadata is missing", func() {
			input := &AliCloudVpc{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AliCloudVpc",
				Spec: &AliCloudVpcSpec{
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
			input := &AliCloudVpc{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AliCloudVpc",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test",
				},
				Spec: &AliCloudVpcSpec{
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
