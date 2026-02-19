package alicloudvswitchv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	foreignkeyv1 "github.com/plantonhq/openmcf/apis/org/openmcf/shared/foreignkey/v1"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared"
)

func TestAlicloudVswitchSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AlicloudVswitchSpec Validation Tests")
}

func stringValueOrRef(val string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: val},
	}
}

var _ = ginkgo.Describe("AlicloudVswitchSpec Validation Tests", func() {

	ginkgo.Describe("valid input", func() {

		ginkgo.It("should pass with minimal required fields", func() {
			input := &AlicloudVswitch{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudVswitch",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test-vswitch",
				},
				Spec: &AlicloudVswitchSpec{
					Region:      "cn-hangzhou",
					VpcId:       stringValueOrRef("vpc-abc123"),
					ZoneId:      "cn-hangzhou-a",
					CidrBlock:   "10.0.0.0/24",
					VswitchName: "dev-vswitch",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with all optional fields populated", func() {
			input := &AlicloudVswitch{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudVswitch",
				Metadata: &shared.CloudResourceMetadata{
					Name: "full-config-vswitch",
					Org:  "my-org",
					Env:  "production",
				},
				Spec: &AlicloudVswitchSpec{
					Region:             "cn-shanghai",
					VpcId:              stringValueOrRef("vpc-prod-001"),
					ZoneId:             "cn-shanghai-b",
					CidrBlock:          "172.16.1.0/24",
					VswitchName:        "prod-app-vswitch",
					Description:        "Production application tier VSwitch",
					EnableIpv6:         true,
					Ipv6CidrBlockMask:  42,
					Tags:               map[string]string{"team": "platform", "tier": "application"},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with 192.168 CIDR range", func() {
			input := &AlicloudVswitch{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudVswitch",
				Metadata: &shared.CloudResourceMetadata{
					Name: "small-vswitch",
				},
				Spec: &AlicloudVswitchSpec{
					Region:      "ap-southeast-1",
					VpcId:       stringValueOrRef("vpc-dev-001"),
					ZoneId:      "ap-southeast-1a",
					CidrBlock:   "192.168.0.0/24",
					VswitchName: "dev-vswitch",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with ipv6_cidr_block_mask at lower bound", func() {
			input := &AlicloudVswitch{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudVswitch",
				Metadata: &shared.CloudResourceMetadata{
					Name: "ipv6-lower-bound",
				},
				Spec: &AlicloudVswitchSpec{
					Region:             "cn-hangzhou",
					VpcId:              stringValueOrRef("vpc-abc123"),
					ZoneId:             "cn-hangzhou-a",
					CidrBlock:          "10.0.1.0/24",
					VswitchName:        "ipv6-vswitch",
					EnableIpv6:         true,
					Ipv6CidrBlockMask:  1,
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with ipv6_cidr_block_mask at upper bound", func() {
			input := &AlicloudVswitch{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudVswitch",
				Metadata: &shared.CloudResourceMetadata{
					Name: "ipv6-upper-bound",
				},
				Spec: &AlicloudVswitchSpec{
					Region:             "cn-hangzhou",
					VpcId:              stringValueOrRef("vpc-abc123"),
					ZoneId:             "cn-hangzhou-a",
					CidrBlock:          "10.0.2.0/24",
					VswitchName:        "ipv6-max-vswitch",
					EnableIpv6:         true,
					Ipv6CidrBlockMask:  255,
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with vpc_id as value_from reference", func() {
			input := &AlicloudVswitch{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudVswitch",
				Metadata: &shared.CloudResourceMetadata{
					Name: "ref-vswitch",
				},
				Spec: &AlicloudVswitchSpec{
					Region: "cn-hangzhou",
					VpcId: &foreignkeyv1.StringValueOrRef{
						LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
							ValueFrom: &foreignkeyv1.ValueFromRef{
								Name: "my-vpc",
							},
						},
					},
					ZoneId:      "cn-hangzhou-a",
					CidrBlock:   "10.0.0.0/24",
					VswitchName: "ref-vswitch",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("invalid input", func() {

		ginkgo.It("should fail when region is missing", func() {
			input := &AlicloudVswitch{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudVswitch",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test",
				},
				Spec: &AlicloudVswitchSpec{
					VpcId:       stringValueOrRef("vpc-abc123"),
					ZoneId:      "cn-hangzhou-a",
					CidrBlock:   "10.0.0.0/24",
					VswitchName: "my-vswitch",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when vpc_id is missing", func() {
			input := &AlicloudVswitch{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudVswitch",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test",
				},
				Spec: &AlicloudVswitchSpec{
					Region:      "cn-hangzhou",
					ZoneId:      "cn-hangzhou-a",
					CidrBlock:   "10.0.0.0/24",
					VswitchName: "my-vswitch",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when zone_id is missing", func() {
			input := &AlicloudVswitch{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudVswitch",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test",
				},
				Spec: &AlicloudVswitchSpec{
					Region:      "cn-hangzhou",
					VpcId:       stringValueOrRef("vpc-abc123"),
					CidrBlock:   "10.0.0.0/24",
					VswitchName: "my-vswitch",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when cidr_block is missing", func() {
			input := &AlicloudVswitch{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudVswitch",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test",
				},
				Spec: &AlicloudVswitchSpec{
					Region:      "cn-hangzhou",
					VpcId:       stringValueOrRef("vpc-abc123"),
					ZoneId:      "cn-hangzhou-a",
					VswitchName: "my-vswitch",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when vswitch_name is missing", func() {
			input := &AlicloudVswitch{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudVswitch",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test",
				},
				Spec: &AlicloudVswitchSpec{
					Region:    "cn-hangzhou",
					VpcId:     stringValueOrRef("vpc-abc123"),
					ZoneId:    "cn-hangzhou-a",
					CidrBlock: "10.0.0.0/24",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when vswitch_name exceeds max length", func() {
			longName := ""
			for i := 0; i < 129; i++ {
				longName += "a"
			}
			input := &AlicloudVswitch{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudVswitch",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test",
				},
				Spec: &AlicloudVswitchSpec{
					Region:      "cn-hangzhou",
					VpcId:       stringValueOrRef("vpc-abc123"),
					ZoneId:      "cn-hangzhou-a",
					CidrBlock:   "10.0.0.0/24",
					VswitchName: longName,
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when api_version is wrong", func() {
			input := &AlicloudVswitch{
				ApiVersion: "wrong/v1",
				Kind:       "AlicloudVswitch",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test",
				},
				Spec: &AlicloudVswitchSpec{
					Region:      "cn-hangzhou",
					VpcId:       stringValueOrRef("vpc-abc123"),
					ZoneId:      "cn-hangzhou-a",
					CidrBlock:   "10.0.0.0/24",
					VswitchName: "my-vswitch",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when kind is wrong", func() {
			input := &AlicloudVswitch{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "WrongKind",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test",
				},
				Spec: &AlicloudVswitchSpec{
					Region:      "cn-hangzhou",
					VpcId:       stringValueOrRef("vpc-abc123"),
					ZoneId:      "cn-hangzhou-a",
					CidrBlock:   "10.0.0.0/24",
					VswitchName: "my-vswitch",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when metadata is missing", func() {
			input := &AlicloudVswitch{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudVswitch",
				Spec: &AlicloudVswitchSpec{
					Region:      "cn-hangzhou",
					VpcId:       stringValueOrRef("vpc-abc123"),
					ZoneId:      "cn-hangzhou-a",
					CidrBlock:   "10.0.0.0/24",
					VswitchName: "my-vswitch",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when ipv6_cidr_block_mask exceeds upper bound", func() {
			input := &AlicloudVswitch{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudVswitch",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test",
				},
				Spec: &AlicloudVswitchSpec{
					Region:             "cn-hangzhou",
					VpcId:              stringValueOrRef("vpc-abc123"),
					ZoneId:             "cn-hangzhou-a",
					CidrBlock:          "10.0.0.0/24",
					VswitchName:        "my-vswitch",
					EnableIpv6:         true,
					Ipv6CidrBlockMask:  256,
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when ipv6_cidr_block_mask is negative", func() {
			input := &AlicloudVswitch{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudVswitch",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test",
				},
				Spec: &AlicloudVswitchSpec{
					Region:             "cn-hangzhou",
					VpcId:              stringValueOrRef("vpc-abc123"),
					ZoneId:             "cn-hangzhou-a",
					CidrBlock:          "10.0.0.0/24",
					VswitchName:        "my-vswitch",
					EnableIpv6:         true,
					Ipv6CidrBlockMask:  -1,
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})
	})
})
