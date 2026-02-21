package aliclouddnsrecordv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared"
)

func TestAliCloudDnsRecordSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AliCloudDnsRecordSpec Validation Tests")
}

var _ = ginkgo.Describe("AliCloudDnsRecordSpec Validation Tests", func() {

	ginkgo.Describe("valid input", func() {

		ginkgo.It("should pass with minimal required fields (A record)", func() {
			input := &AliCloudDnsRecord{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AliCloudDnsRecord",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test-a-record",
				},
				Spec: &AliCloudDnsRecordSpec{
					Region:     "cn-hangzhou",
					DomainName: "example.com",
					Rr:         "www",
					Type:       "A",
					Value:      "1.2.3.4",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with all optional fields populated", func() {
			input := &AliCloudDnsRecord{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AliCloudDnsRecord",
				Metadata: &shared.CloudResourceMetadata{
					Name: "full-config-record",
					Org:  "my-org",
					Env:  "production",
				},
				Spec: &AliCloudDnsRecordSpec{
					Region:     "cn-shanghai",
					DomainName: "platform.example.com",
					Rr:         "mail",
					Type:       "MX",
					Value:      "mx1.example.com",
					Ttl:        300,
					Priority:   5,
					Line:       "default",
					Status:     "ENABLE",
					Remark:     "Primary mail exchange record",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with CNAME record", func() {
			input := &AliCloudDnsRecord{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AliCloudDnsRecord",
				Metadata: &shared.CloudResourceMetadata{
					Name: "cname-record",
				},
				Spec: &AliCloudDnsRecordSpec{
					Region:     "ap-southeast-1",
					DomainName: "example.com",
					Rr:         "cdn",
					Type:       "CNAME",
					Value:      "cdn.provider.com",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with apex record using @", func() {
			input := &AliCloudDnsRecord{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AliCloudDnsRecord",
				Metadata: &shared.CloudResourceMetadata{
					Name: "apex-record",
				},
				Spec: &AliCloudDnsRecordSpec{
					Region:     "cn-hangzhou",
					DomainName: "example.com",
					Rr:         "@",
					Type:       "A",
					Value:      "203.0.113.10",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with wildcard record", func() {
			input := &AliCloudDnsRecord{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AliCloudDnsRecord",
				Metadata: &shared.CloudResourceMetadata{
					Name: "wildcard-record",
				},
				Spec: &AliCloudDnsRecordSpec{
					Region:     "cn-hangzhou",
					DomainName: "example.com",
					Rr:         "*",
					Type:       "CNAME",
					Value:      "fallback.example.com",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with TXT record for SPF", func() {
			input := &AliCloudDnsRecord{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AliCloudDnsRecord",
				Metadata: &shared.CloudResourceMetadata{
					Name: "spf-record",
				},
				Spec: &AliCloudDnsRecordSpec{
					Region:     "cn-hangzhou",
					DomainName: "example.com",
					Rr:         "@",
					Type:       "TXT",
					Value:      "v=spf1 include:example.com ~all",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with DISABLE status", func() {
			input := &AliCloudDnsRecord{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AliCloudDnsRecord",
				Metadata: &shared.CloudResourceMetadata{
					Name: "disabled-record",
				},
				Spec: &AliCloudDnsRecordSpec{
					Region:     "cn-hangzhou",
					DomainName: "example.com",
					Rr:         "old-service",
					Type:       "A",
					Value:      "1.2.3.4",
					Status:     "DISABLE",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with CAA record", func() {
			input := &AliCloudDnsRecord{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AliCloudDnsRecord",
				Metadata: &shared.CloudResourceMetadata{
					Name: "caa-record",
				},
				Spec: &AliCloudDnsRecordSpec{
					Region:     "cn-hangzhou",
					DomainName: "example.com",
					Rr:         "@",
					Type:       "CAA",
					Value:      `0 issue "letsencrypt.org"`,
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("invalid input", func() {

		ginkgo.It("should fail when region is missing", func() {
			input := &AliCloudDnsRecord{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AliCloudDnsRecord",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test",
				},
				Spec: &AliCloudDnsRecordSpec{
					DomainName: "example.com",
					Rr:         "www",
					Type:       "A",
					Value:      "1.2.3.4",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when domain_name is missing", func() {
			input := &AliCloudDnsRecord{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AliCloudDnsRecord",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test",
				},
				Spec: &AliCloudDnsRecordSpec{
					Region: "cn-hangzhou",
					Rr:     "www",
					Type:   "A",
					Value:  "1.2.3.4",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when rr is missing", func() {
			input := &AliCloudDnsRecord{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AliCloudDnsRecord",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test",
				},
				Spec: &AliCloudDnsRecordSpec{
					Region:     "cn-hangzhou",
					DomainName: "example.com",
					Type:       "A",
					Value:      "1.2.3.4",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when type is missing", func() {
			input := &AliCloudDnsRecord{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AliCloudDnsRecord",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test",
				},
				Spec: &AliCloudDnsRecordSpec{
					Region:     "cn-hangzhou",
					DomainName: "example.com",
					Rr:         "www",
					Value:      "1.2.3.4",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when type is invalid", func() {
			input := &AliCloudDnsRecord{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AliCloudDnsRecord",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test",
				},
				Spec: &AliCloudDnsRecordSpec{
					Region:     "cn-hangzhou",
					DomainName: "example.com",
					Rr:         "www",
					Type:       "INVALID",
					Value:      "1.2.3.4",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when value is missing", func() {
			input := &AliCloudDnsRecord{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AliCloudDnsRecord",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test",
				},
				Spec: &AliCloudDnsRecordSpec{
					Region:     "cn-hangzhou",
					DomainName: "example.com",
					Rr:         "www",
					Type:       "A",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when api_version is wrong", func() {
			input := &AliCloudDnsRecord{
				ApiVersion: "wrong/v1",
				Kind:       "AliCloudDnsRecord",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test",
				},
				Spec: &AliCloudDnsRecordSpec{
					Region:     "cn-hangzhou",
					DomainName: "example.com",
					Rr:         "www",
					Type:       "A",
					Value:      "1.2.3.4",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when kind is wrong", func() {
			input := &AliCloudDnsRecord{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "WrongKind",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test",
				},
				Spec: &AliCloudDnsRecordSpec{
					Region:     "cn-hangzhou",
					DomainName: "example.com",
					Rr:         "www",
					Type:       "A",
					Value:      "1.2.3.4",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when metadata is missing", func() {
			input := &AliCloudDnsRecord{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AliCloudDnsRecord",
				Spec: &AliCloudDnsRecordSpec{
					Region:     "cn-hangzhou",
					DomainName: "example.com",
					Rr:         "www",
					Type:       "A",
					Value:      "1.2.3.4",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when spec is missing", func() {
			input := &AliCloudDnsRecord{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AliCloudDnsRecord",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test",
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
			input := &AliCloudDnsRecord{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AliCloudDnsRecord",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test",
				},
				Spec: &AliCloudDnsRecordSpec{
					Region:     "cn-hangzhou",
					DomainName: longDomain,
					Rr:         "www",
					Type:       "A",
					Value:      "1.2.3.4",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when status is invalid", func() {
			input := &AliCloudDnsRecord{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AliCloudDnsRecord",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test",
				},
				Spec: &AliCloudDnsRecordSpec{
					Region:     "cn-hangzhou",
					DomainName: "example.com",
					Rr:         "www",
					Type:       "A",
					Value:      "1.2.3.4",
					Status:     "INVALID_STATUS",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})
	})
})
