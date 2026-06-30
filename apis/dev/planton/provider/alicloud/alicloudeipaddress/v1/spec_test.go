package alicloudeipaddressv1

import (
	"strings"
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/apis/dev/planton/shared"
	"google.golang.org/protobuf/proto"
)

func TestAliCloudEipAddressSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AliCloudEipAddressSpec Validation Tests")
}

var _ = ginkgo.Describe("AliCloudEipAddressSpec Validation Tests", func() {

	ginkgo.Describe("valid input", func() {

		ginkgo.It("should pass with minimal required fields", func() {
			input := &AliCloudEipAddress{
				ApiVersion: "alicloud.planton.dev/v1",
				Kind:       "AliCloudEipAddress",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test-eip",
				},
				Spec: &AliCloudEipAddressSpec{
					Region: "cn-hangzhou",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with all optional fields populated", func() {
			input := &AliCloudEipAddress{
				ApiVersion: "alicloud.planton.dev/v1",
				Kind:       "AliCloudEipAddress",
				Metadata: &shared.CloudResourceMetadata{
					Name: "full-eip",
					Org:  "my-org",
					Env:  "production",
				},
				Spec: &AliCloudEipAddressSpec{
					Region:             "cn-shanghai",
					AddressName:        "prod-nat-eip",
					Description:        "EIP for production NAT gateway",
					Bandwidth:          proto.Int32(100),
					InternetChargeType: proto.String("PayByBandwidth"),
					Isp:                proto.String("BGP_PRO"),
					ResourceGroupId:    "rg-abc123",
					Tags:               map[string]string{"team": "platform", "purpose": "nat"},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with PayByTraffic charge type", func() {
			input := &AliCloudEipAddress{
				ApiVersion: "alicloud.planton.dev/v1",
				Kind:       "AliCloudEipAddress",
				Metadata: &shared.CloudResourceMetadata{
					Name: "traffic-eip",
				},
				Spec: &AliCloudEipAddressSpec{
					Region:             "us-west-1",
					InternetChargeType: proto.String("PayByTraffic"),
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with China-specific ISP values", func() {
			for _, isp := range []string{"ChinaTelecom", "ChinaUnicom", "ChinaMobile"} {
				input := &AliCloudEipAddress{
					ApiVersion: "alicloud.planton.dev/v1",
					Kind:       "AliCloudEipAddress",
					Metadata: &shared.CloudResourceMetadata{
						Name: "isp-eip",
					},
					Spec: &AliCloudEipAddressSpec{
						Region: "cn-hangzhou",
						Isp:    proto.String(isp),
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			}
		})

		ginkgo.It("should pass with L2 and special ISP values", func() {
			for _, isp := range []string{"ChinaTelecom_L2", "ChinaUnicom_L2", "ChinaMobile_L2", "BGP_FinanceCloud", "BGP_International"} {
				input := &AliCloudEipAddress{
					ApiVersion: "alicloud.planton.dev/v1",
					Kind:       "AliCloudEipAddress",
					Metadata: &shared.CloudResourceMetadata{
						Name: "special-isp-eip",
					},
					Spec: &AliCloudEipAddressSpec{
						Region: "cn-hangzhou",
						Isp:    proto.String(isp),
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			}
		})

		ginkgo.It("should pass with bandwidth at boundary values", func() {
			input := &AliCloudEipAddress{
				ApiVersion: "alicloud.planton.dev/v1",
				Kind:       "AliCloudEipAddress",
				Metadata: &shared.CloudResourceMetadata{
					Name: "min-bw-eip",
				},
				Spec: &AliCloudEipAddressSpec{
					Region:    "cn-hangzhou",
					Bandwidth: proto.Int32(1),
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())

			input.Spec.Bandwidth = proto.Int32(1000)
			err = protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("invalid input", func() {

		ginkgo.It("should fail when region is missing", func() {
			input := &AliCloudEipAddress{
				ApiVersion: "alicloud.planton.dev/v1",
				Kind:       "AliCloudEipAddress",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test",
				},
				Spec: &AliCloudEipAddressSpec{},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when api_version is wrong", func() {
			input := &AliCloudEipAddress{
				ApiVersion: "wrong/v1",
				Kind:       "AliCloudEipAddress",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test",
				},
				Spec: &AliCloudEipAddressSpec{
					Region: "cn-hangzhou",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when kind is wrong", func() {
			input := &AliCloudEipAddress{
				ApiVersion: "alicloud.planton.dev/v1",
				Kind:       "WrongKind",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test",
				},
				Spec: &AliCloudEipAddressSpec{
					Region: "cn-hangzhou",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when metadata is missing", func() {
			input := &AliCloudEipAddress{
				ApiVersion: "alicloud.planton.dev/v1",
				Kind:       "AliCloudEipAddress",
				Spec: &AliCloudEipAddressSpec{
					Region: "cn-hangzhou",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when internet_charge_type has invalid value", func() {
			input := &AliCloudEipAddress{
				ApiVersion: "alicloud.planton.dev/v1",
				Kind:       "AliCloudEipAddress",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test",
				},
				Spec: &AliCloudEipAddressSpec{
					Region:             "cn-hangzhou",
					InternetChargeType: proto.String("PayByRequest"),
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when isp has invalid value", func() {
			input := &AliCloudEipAddress{
				ApiVersion: "alicloud.planton.dev/v1",
				Kind:       "AliCloudEipAddress",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test",
				},
				Spec: &AliCloudEipAddressSpec{
					Region: "cn-hangzhou",
					Isp:    proto.String("InvalidISP"),
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when bandwidth is below minimum", func() {
			input := &AliCloudEipAddress{
				ApiVersion: "alicloud.planton.dev/v1",
				Kind:       "AliCloudEipAddress",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test",
				},
				Spec: &AliCloudEipAddressSpec{
					Region:    "cn-hangzhou",
					Bandwidth: proto.Int32(0),
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when bandwidth exceeds maximum", func() {
			input := &AliCloudEipAddress{
				ApiVersion: "alicloud.planton.dev/v1",
				Kind:       "AliCloudEipAddress",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test",
				},
				Spec: &AliCloudEipAddressSpec{
					Region:    "cn-hangzhou",
					Bandwidth: proto.Int32(1001),
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when address_name exceeds max length", func() {
			input := &AliCloudEipAddress{
				ApiVersion: "alicloud.planton.dev/v1",
				Kind:       "AliCloudEipAddress",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test",
				},
				Spec: &AliCloudEipAddressSpec{
					Region:      "cn-hangzhou",
					AddressName: strings.Repeat("a", 129),
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})
	})
})
