package aliclouddnszonev1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared"
)

func TestAlicloudDnsZoneSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AlicloudDnsZoneSpec Validation Tests")
}

var _ = ginkgo.Describe("AlicloudDnsZoneSpec Validation Tests", func() {

	ginkgo.Describe("valid input", func() {

		ginkgo.It("should pass with minimal required fields", func() {
			input := &AlicloudDnsZone{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudDnsZone",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test-domain",
				},
				Spec: &AlicloudDnsZoneSpec{
					Region:     "cn-hangzhou",
					DomainName: "example.com",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with all optional fields populated", func() {
			input := &AlicloudDnsZone{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudDnsZone",
				Metadata: &shared.CloudResourceMetadata{
					Name: "full-config-domain",
					Org:  "my-org",
					Env:  "production",
				},
				Spec: &AlicloudDnsZoneSpec{
					Region:          "cn-shanghai",
					DomainName:      "platform.example.com",
					GroupId:         "group-abc123",
					Remark:          "Primary platform domain for production",
					ResourceGroupId: "rg-prod-456",
					Tags:            map[string]string{"team": "platform", "cost-center": "eng"},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with a subdomain", func() {
			input := &AlicloudDnsZone{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudDnsZone",
				Metadata: &shared.CloudResourceMetadata{
					Name: "subdomain",
				},
				Spec: &AlicloudDnsZoneSpec{
					Region:     "ap-southeast-1",
					DomainName: "api.services.example.com",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with tags only", func() {
			input := &AlicloudDnsZone{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudDnsZone",
				Metadata: &shared.CloudResourceMetadata{
					Name: "tagged-domain",
				},
				Spec: &AlicloudDnsZoneSpec{
					Region:     "us-west-1",
					DomainName: "tagged.example.com",
					Tags:       map[string]string{"env": "staging"},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("invalid input", func() {

		ginkgo.It("should fail when region is missing", func() {
			input := &AlicloudDnsZone{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudDnsZone",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test",
				},
				Spec: &AlicloudDnsZoneSpec{
					DomainName: "example.com",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when domain_name is missing", func() {
			input := &AlicloudDnsZone{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudDnsZone",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test",
				},
				Spec: &AlicloudDnsZoneSpec{
					Region: "cn-hangzhou",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when api_version is wrong", func() {
			input := &AlicloudDnsZone{
				ApiVersion: "wrong/v1",
				Kind:       "AlicloudDnsZone",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test",
				},
				Spec: &AlicloudDnsZoneSpec{
					Region:     "cn-hangzhou",
					DomainName: "example.com",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when kind is wrong", func() {
			input := &AlicloudDnsZone{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "WrongKind",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test",
				},
				Spec: &AlicloudDnsZoneSpec{
					Region:     "cn-hangzhou",
					DomainName: "example.com",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when metadata is missing", func() {
			input := &AlicloudDnsZone{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudDnsZone",
				Spec: &AlicloudDnsZoneSpec{
					Region:     "cn-hangzhou",
					DomainName: "example.com",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when domain_name exceeds max length", func() {
			longDomain := ""
			for i := 0; i < 254; i++ {
				longDomain += "a"
			}
			input := &AlicloudDnsZone{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudDnsZone",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test",
				},
				Spec: &AlicloudDnsZoneSpec{
					Region:     "cn-hangzhou",
					DomainName: longDomain,
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when spec is missing", func() {
			input := &AlicloudDnsZone{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudDnsZone",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})
	})
})
