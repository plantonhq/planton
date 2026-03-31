package alicloudcdndomainv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared"
)

func TestAliCloudCdnDomainSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AliCloudCdnDomainSpec Validation Tests")
}

func validSource() *AliCloudCdnDomainSource {
	return &AliCloudCdnDomainSource{
		Type:    "ipaddr",
		Content: "1.2.3.4",
	}
}

var _ = ginkgo.Describe("AliCloudCdnDomainSpec Validation Tests", func() {

	ginkgo.Describe("valid input", func() {

		ginkgo.It("should pass with minimal required fields", func() {
			input := &AliCloudCdnDomain{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AliCloudCdnDomain",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test-cdn",
				},
				Spec: &AliCloudCdnDomainSpec{
					Region:     "cn-hangzhou",
					DomainName: "cdn.example.com",
					CdnType:    "web",
					Sources:    []*AliCloudCdnDomainSource{validSource()},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with all optional fields populated", func() {
			input := &AliCloudCdnDomain{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AliCloudCdnDomain",
				Metadata: &shared.CloudResourceMetadata{
					Name: "full-cdn",
					Org:  "my-org",
					Env:  "production",
				},
				Spec: &AliCloudCdnDomainSpec{
					Region:     "cn-shanghai",
					DomainName: "static.example.com",
					CdnType:    "download",
					Scope:      "global",
					Sources: []*AliCloudCdnDomainSource{
						{
							Type:     "domain",
							Content:  "origin.example.com",
							Port:     443,
							Priority: 20,
							Weight:   10,
						},
						{
							Type:     "ipaddr",
							Content:  "10.0.0.1",
							Port:     80,
							Priority: 30,
							Weight:   5,
						},
					},
					CertificateConfig: &AliCloudCdnDomainCertificateConfig{
						CertName:                "my-cert",
						CertType:                "cas",
						CertId:                  "cas-12345",
						CertRegion:              "cn-hangzhou",
						ServerCertificateStatus: "on",
					},
					CheckUrl:        "http://origin.example.com/health",
					ResourceGroupId: "rg-prod-123",
					Tags:            map[string]string{"team": "platform", "env": "prod"},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with cdn_type video", func() {
			input := &AliCloudCdnDomain{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AliCloudCdnDomain",
				Metadata: &shared.CloudResourceMetadata{
					Name: "video-cdn",
				},
				Spec: &AliCloudCdnDomainSpec{
					Region:     "ap-southeast-1",
					DomainName: "video.example.com",
					CdnType:    "video",
					Sources: []*AliCloudCdnDomainSource{
						{Type: "oss", Content: "my-bucket.oss-cn-hangzhou.aliyuncs.com"},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with source type common", func() {
			input := &AliCloudCdnDomain{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AliCloudCdnDomain",
				Metadata: &shared.CloudResourceMetadata{
					Name: "common-cdn",
				},
				Spec: &AliCloudCdnDomainSpec{
					Region:     "cn-hangzhou",
					DomainName: "l2.example.com",
					CdnType:    "web",
					Sources: []*AliCloudCdnDomainSource{
						{Type: "common", Content: "l2-origin.example.com"},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with scope domestic", func() {
			input := &AliCloudCdnDomain{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AliCloudCdnDomain",
				Metadata: &shared.CloudResourceMetadata{
					Name: "domestic-cdn",
				},
				Spec: &AliCloudCdnDomainSpec{
					Region:     "cn-hangzhou",
					DomainName: "cn.example.com",
					CdnType:    "web",
					Scope:      "domestic",
					Sources:    []*AliCloudCdnDomainSource{validSource()},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with scope overseas", func() {
			input := &AliCloudCdnDomain{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AliCloudCdnDomain",
				Metadata: &shared.CloudResourceMetadata{
					Name: "overseas-cdn",
				},
				Spec: &AliCloudCdnDomainSpec{
					Region:     "us-west-1",
					DomainName: "global.example.com",
					CdnType:    "web",
					Scope:      "overseas",
					Sources:    []*AliCloudCdnDomainSource{validSource()},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with upload certificate", func() {
			input := &AliCloudCdnDomain{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AliCloudCdnDomain",
				Metadata: &shared.CloudResourceMetadata{
					Name: "https-cdn",
				},
				Spec: &AliCloudCdnDomainSpec{
					Region:     "cn-hangzhou",
					DomainName: "secure.example.com",
					CdnType:    "web",
					Sources:    []*AliCloudCdnDomainSource{validSource()},
					CertificateConfig: &AliCloudCdnDomainCertificateConfig{
						CertName:                "my-upload-cert",
						CertType:                "upload",
						ServerCertificate:       "-----BEGIN CERTIFICATE-----\ntest\n-----END CERTIFICATE-----",
						PrivateKey:              "-----BEGIN RSA PRIVATE KEY-----\ntest\n-----END RSA PRIVATE KEY-----",
						ServerCertificateStatus: "on",
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with certificate status off", func() {
			input := &AliCloudCdnDomain{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AliCloudCdnDomain",
				Metadata: &shared.CloudResourceMetadata{
					Name: "no-https-cdn",
				},
				Spec: &AliCloudCdnDomainSpec{
					Region:     "cn-hangzhou",
					DomainName: "http.example.com",
					CdnType:    "web",
					Sources:    []*AliCloudCdnDomainSource{validSource()},
					CertificateConfig: &AliCloudCdnDomainCertificateConfig{
						ServerCertificateStatus: "off",
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("invalid input", func() {

		ginkgo.It("should fail when region is missing", func() {
			input := &AliCloudCdnDomain{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AliCloudCdnDomain",
				Metadata:   &shared.CloudResourceMetadata{Name: "test"},
				Spec: &AliCloudCdnDomainSpec{
					DomainName: "cdn.example.com",
					CdnType:    "web",
					Sources:    []*AliCloudCdnDomainSource{validSource()},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when domain_name is missing", func() {
			input := &AliCloudCdnDomain{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AliCloudCdnDomain",
				Metadata:   &shared.CloudResourceMetadata{Name: "test"},
				Spec: &AliCloudCdnDomainSpec{
					Region:  "cn-hangzhou",
					CdnType: "web",
					Sources: []*AliCloudCdnDomainSource{validSource()},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when cdn_type is missing", func() {
			input := &AliCloudCdnDomain{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AliCloudCdnDomain",
				Metadata:   &shared.CloudResourceMetadata{Name: "test"},
				Spec: &AliCloudCdnDomainSpec{
					Region:     "cn-hangzhou",
					DomainName: "cdn.example.com",
					Sources:    []*AliCloudCdnDomainSource{validSource()},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when cdn_type is invalid", func() {
			input := &AliCloudCdnDomain{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AliCloudCdnDomain",
				Metadata:   &shared.CloudResourceMetadata{Name: "test"},
				Spec: &AliCloudCdnDomainSpec{
					Region:     "cn-hangzhou",
					DomainName: "cdn.example.com",
					CdnType:    "invalid",
					Sources:    []*AliCloudCdnDomainSource{validSource()},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when sources is empty", func() {
			input := &AliCloudCdnDomain{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AliCloudCdnDomain",
				Metadata:   &shared.CloudResourceMetadata{Name: "test"},
				Spec: &AliCloudCdnDomainSpec{
					Region:     "cn-hangzhou",
					DomainName: "cdn.example.com",
					CdnType:    "web",
					Sources:    []*AliCloudCdnDomainSource{},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when source type is invalid", func() {
			input := &AliCloudCdnDomain{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AliCloudCdnDomain",
				Metadata:   &shared.CloudResourceMetadata{Name: "test"},
				Spec: &AliCloudCdnDomainSpec{
					Region:     "cn-hangzhou",
					DomainName: "cdn.example.com",
					CdnType:    "web",
					Sources: []*AliCloudCdnDomainSource{
						{Type: "invalid", Content: "1.2.3.4"},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when source content is empty", func() {
			input := &AliCloudCdnDomain{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AliCloudCdnDomain",
				Metadata:   &shared.CloudResourceMetadata{Name: "test"},
				Spec: &AliCloudCdnDomainSpec{
					Region:     "cn-hangzhou",
					DomainName: "cdn.example.com",
					CdnType:    "web",
					Sources: []*AliCloudCdnDomainSource{
						{Type: "ipaddr", Content: ""},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when scope is invalid", func() {
			input := &AliCloudCdnDomain{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AliCloudCdnDomain",
				Metadata:   &shared.CloudResourceMetadata{Name: "test"},
				Spec: &AliCloudCdnDomainSpec{
					Region:     "cn-hangzhou",
					DomainName: "cdn.example.com",
					CdnType:    "web",
					Scope:      "invalid",
					Sources:    []*AliCloudCdnDomainSource{validSource()},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when cert_type is invalid", func() {
			input := &AliCloudCdnDomain{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AliCloudCdnDomain",
				Metadata:   &shared.CloudResourceMetadata{Name: "test"},
				Spec: &AliCloudCdnDomainSpec{
					Region:     "cn-hangzhou",
					DomainName: "cdn.example.com",
					CdnType:    "web",
					Sources:    []*AliCloudCdnDomainSource{validSource()},
					CertificateConfig: &AliCloudCdnDomainCertificateConfig{
						CertType: "invalid",
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when server_certificate_status is invalid", func() {
			input := &AliCloudCdnDomain{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AliCloudCdnDomain",
				Metadata:   &shared.CloudResourceMetadata{Name: "test"},
				Spec: &AliCloudCdnDomainSpec{
					Region:     "cn-hangzhou",
					DomainName: "cdn.example.com",
					CdnType:    "web",
					Sources:    []*AliCloudCdnDomainSource{validSource()},
					CertificateConfig: &AliCloudCdnDomainCertificateConfig{
						ServerCertificateStatus: "invalid",
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when api_version is wrong", func() {
			input := &AliCloudCdnDomain{
				ApiVersion: "wrong/v1",
				Kind:       "AliCloudCdnDomain",
				Metadata:   &shared.CloudResourceMetadata{Name: "test"},
				Spec: &AliCloudCdnDomainSpec{
					Region:     "cn-hangzhou",
					DomainName: "cdn.example.com",
					CdnType:    "web",
					Sources:    []*AliCloudCdnDomainSource{validSource()},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when kind is wrong", func() {
			input := &AliCloudCdnDomain{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "WrongKind",
				Metadata:   &shared.CloudResourceMetadata{Name: "test"},
				Spec: &AliCloudCdnDomainSpec{
					Region:     "cn-hangzhou",
					DomainName: "cdn.example.com",
					CdnType:    "web",
					Sources:    []*AliCloudCdnDomainSource{validSource()},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when metadata is missing", func() {
			input := &AliCloudCdnDomain{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AliCloudCdnDomain",
				Spec: &AliCloudCdnDomainSpec{
					Region:     "cn-hangzhou",
					DomainName: "cdn.example.com",
					CdnType:    "web",
					Sources:    []*AliCloudCdnDomainSource{validSource()},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when spec is missing", func() {
			input := &AliCloudCdnDomain{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AliCloudCdnDomain",
				Metadata:   &shared.CloudResourceMetadata{Name: "test"},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when domain_name exceeds max length", func() {
			longDomain := ""
			for i := 0; i < 64; i++ {
				longDomain += "a"
			}
			input := &AliCloudCdnDomain{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AliCloudCdnDomain",
				Metadata:   &shared.CloudResourceMetadata{Name: "test"},
				Spec: &AliCloudCdnDomainSpec{
					Region:     "cn-hangzhou",
					DomainName: longDomain,
					CdnType:    "web",
					Sources:    []*AliCloudCdnDomainSource{validSource()},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})
	})
})
