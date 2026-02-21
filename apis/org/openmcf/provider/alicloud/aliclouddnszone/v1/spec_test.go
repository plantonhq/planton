package aliclouddnszonev1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared"
)

func TestAliCloudDnsZoneSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AliCloudDnsZoneSpec Validation Tests")
}

var _ = ginkgo.Describe("AliCloudDnsZoneSpec Validation Tests", func() {

	ginkgo.Describe("valid input", func() {

		ginkgo.It("should pass with minimal required fields", func() {
			input := &AliCloudDnsZone{
				ApiVersion: "ali-cloud.openmcf.org/v1",
				Kind:       "AliCloudDnsZone",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test-domain",
				},
				Spec: &AliCloudDnsZoneSpec{
					Region:     "cn-hangzhou",
					DomainName: "example.com",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with all optional fields populated", func() {
			input := &AliCloudDnsZone{
				ApiVersion: "ali-cloud.openmcf.org/v1",
				Kind:       "AliCloudDnsZone",
				Metadata: &shared.CloudResourceMetadata{
					Name: "full-config-domain",
					Org:  "my-org",
					Env:  "production",
				},
				Spec: &AliCloudDnsZoneSpec{
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
			input := &AliCloudDnsZone{
				ApiVersion: "ali-cloud.openmcf.org/v1",
				Kind:       "AliCloudDnsZone",
				Metadata: &shared.CloudResourceMetadata{
					Name: "subdomain",
				},
				Spec: &AliCloudDnsZoneSpec{
					Region:     "ap-southeast-1",
					DomainName: "api.services.example.com",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with tags only", func() {
			input := &AliCloudDnsZone{
				ApiVersion: "ali-cloud.openmcf.org/v1",
				Kind:       "AliCloudDnsZone",
				Metadata: &shared.CloudResourceMetadata{
					Name: "tagged-domain",
				},
				Spec: &AliCloudDnsZoneSpec{
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
			input := &AliCloudDnsZone{
				ApiVersion: "ali-cloud.openmcf.org/v1",
				Kind:       "AliCloudDnsZone",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test",
				},
				Spec: &AliCloudDnsZoneSpec{
					DomainName: "example.com",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when domain_name is missing", func() {
			input := &AliCloudDnsZone{
				ApiVersion: "ali-cloud.openmcf.org/v1",
				Kind:       "AliCloudDnsZone",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test",
				},
				Spec: &AliCloudDnsZoneSpec{
					Region: "cn-hangzhou",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when api_version is wrong", func() {
			input := &AliCloudDnsZone{
				ApiVersion: "wrong/v1",
				Kind:       "AliCloudDnsZone",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test",
				},
				Spec: &AliCloudDnsZoneSpec{
					Region:     "cn-hangzhou",
					DomainName: "example.com",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when kind is wrong", func() {
			input := &AliCloudDnsZone{
				ApiVersion: "ali-cloud.openmcf.org/v1",
				Kind:       "WrongKind",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test",
				},
				Spec: &AliCloudDnsZoneSpec{
					Region:     "cn-hangzhou",
					DomainName: "example.com",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when metadata is missing", func() {
			input := &AliCloudDnsZone{
				ApiVersion: "ali-cloud.openmcf.org/v1",
				Kind:       "AliCloudDnsZone",
				Spec: &AliCloudDnsZoneSpec{
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
			input := &AliCloudDnsZone{
				ApiVersion: "ali-cloud.openmcf.org/v1",
				Kind:       "AliCloudDnsZone",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test",
				},
				Spec: &AliCloudDnsZoneSpec{
					Region:     "cn-hangzhou",
					DomainName: longDomain,
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when spec is missing", func() {
			input := &AliCloudDnsZone{
				ApiVersion: "ali-cloud.openmcf.org/v1",
				Kind:       "AliCloudDnsZone",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})
	})
})
