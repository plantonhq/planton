package alicloudcdndomainv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared"
)

func TestAlicloudCdnDomainSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AlicloudCdnDomainSpec Validation Tests")
}

func validSource() *AlicloudCdnDomainSource {
	return &AlicloudCdnDomainSource{
		Type:    "ipaddr",
		Content: "1.2.3.4",
	}
}

var _ = ginkgo.Describe("AlicloudCdnDomainSpec Validation Tests", func() {

	ginkgo.Describe("valid input", func() {

		ginkgo.It("should pass with minimal required fields", func() {
			input := &AlicloudCdnDomain{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudCdnDomain",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test-cdn",
				},
				Spec: &AlicloudCdnDomainSpec{
					Region:     "cn-hangzhou",
					DomainName: "cdn.example.com",
					CdnType:    "web",
					Sources:    []*AlicloudCdnDomainSource{validSource()},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with all optional fields populated", func() {
			input := &AlicloudCdnDomain{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudCdnDomain",
				Metadata: &shared.CloudResourceMetadata{
					Name: "full-cdn",
					Org:  "my-org",
					Env:  "production",
				},
				Spec: &AlicloudCdnDomainSpec{
					Region:     "cn-shanghai",
					DomainName: "static.example.com",
					CdnType:    "download",
					Scope:      "global",
					Sources: []*AlicloudCdnDomainSource{
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
					CertificateConfig: &AlicloudCdnDomainCertificateConfig{
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
			input := &AlicloudCdnDomain{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudCdnDomain",
				Metadata: &shared.CloudResourceMetadata{
					Name: "video-cdn",
				},
				Spec: &AlicloudCdnDomainSpec{
					Region:     "ap-southeast-1",
					DomainName: "video.example.com",
					CdnType:    "video",
					Sources: []*AlicloudCdnDomainSource{
						{Type: "oss", Content: "my-bucket.oss-cn-hangzhou.aliyuncs.com"},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with source type common", func() {
			input := &AlicloudCdnDomain{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudCdnDomain",
				Metadata: &shared.CloudResourceMetadata{
					Name: "common-cdn",
				},
				Spec: &AlicloudCdnDomainSpec{
					Region:     "cn-hangzhou",
					DomainName: "l2.example.com",
					CdnType:    "web",
					Sources: []*AlicloudCdnDomainSource{
						{Type: "common", Content: "l2-origin.example.com"},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with scope domestic", func() {
			input := &AlicloudCdnDomain{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudCdnDomain",
				Metadata: &shared.CloudResourceMetadata{
					Name: "domestic-cdn",
				},
				Spec: &AlicloudCdnDomainSpec{
					Region:     "cn-hangzhou",
					DomainName: "cn.example.com",
					CdnType:    "web",
					Scope:      "domestic",
					Sources:    []*AlicloudCdnDomainSource{validSource()},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with scope overseas", func() {
			input := &AlicloudCdnDomain{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudCdnDomain",
				Metadata: &shared.CloudResourceMetadata{
					Name: "overseas-cdn",
				},
				Spec: &AlicloudCdnDomainSpec{
					Region:     "us-west-1",
					DomainName: "global.example.com",
					CdnType:    "web",
					Scope:      "overseas",
					Sources:    []*AlicloudCdnDomainSource{validSource()},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with upload certificate", func() {
			input := &AlicloudCdnDomain{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudCdnDomain",
				Metadata: &shared.CloudResourceMetadata{
					Name: "https-cdn",
				},
				Spec: &AlicloudCdnDomainSpec{
					Region:     "cn-hangzhou",
					DomainName: "secure.example.com",
					CdnType:    "web",
					Sources:    []*AlicloudCdnDomainSource{validSource()},
					CertificateConfig: &AlicloudCdnDomainCertificateConfig{
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
			input := &AlicloudCdnDomain{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudCdnDomain",
				Metadata: &shared.CloudResourceMetadata{
					Name: "no-https-cdn",
				},
				Spec: &AlicloudCdnDomainSpec{
					Region:     "cn-hangzhou",
					DomainName: "http.example.com",
					CdnType:    "web",
					Sources:    []*AlicloudCdnDomainSource{validSource()},
					CertificateConfig: &AlicloudCdnDomainCertificateConfig{
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
			input := &AlicloudCdnDomain{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudCdnDomain",
				Metadata:   &shared.CloudResourceMetadata{Name: "test"},
				Spec: &AlicloudCdnDomainSpec{
					DomainName: "cdn.example.com",
					CdnType:    "web",
					Sources:    []*AlicloudCdnDomainSource{validSource()},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when domain_name is missing", func() {
			input := &AlicloudCdnDomain{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudCdnDomain",
				Metadata:   &shared.CloudResourceMetadata{Name: "test"},
				Spec: &AlicloudCdnDomainSpec{
					Region:  "cn-hangzhou",
					CdnType: "web",
					Sources: []*AlicloudCdnDomainSource{validSource()},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when cdn_type is missing", func() {
			input := &AlicloudCdnDomain{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudCdnDomain",
				Metadata:   &shared.CloudResourceMetadata{Name: "test"},
				Spec: &AlicloudCdnDomainSpec{
					Region:     "cn-hangzhou",
					DomainName: "cdn.example.com",
					Sources:    []*AlicloudCdnDomainSource{validSource()},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when cdn_type is invalid", func() {
			input := &AlicloudCdnDomain{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudCdnDomain",
				Metadata:   &shared.CloudResourceMetadata{Name: "test"},
				Spec: &AlicloudCdnDomainSpec{
					Region:     "cn-hangzhou",
					DomainName: "cdn.example.com",
					CdnType:    "invalid",
					Sources:    []*AlicloudCdnDomainSource{validSource()},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when sources is empty", func() {
			input := &AlicloudCdnDomain{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudCdnDomain",
				Metadata:   &shared.CloudResourceMetadata{Name: "test"},
				Spec: &AlicloudCdnDomainSpec{
					Region:     "cn-hangzhou",
					DomainName: "cdn.example.com",
					CdnType:    "web",
					Sources:    []*AlicloudCdnDomainSource{},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when source type is invalid", func() {
			input := &AlicloudCdnDomain{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudCdnDomain",
				Metadata:   &shared.CloudResourceMetadata{Name: "test"},
				Spec: &AlicloudCdnDomainSpec{
					Region:     "cn-hangzhou",
					DomainName: "cdn.example.com",
					CdnType:    "web",
					Sources: []*AlicloudCdnDomainSource{
						{Type: "invalid", Content: "1.2.3.4"},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when source content is empty", func() {
			input := &AlicloudCdnDomain{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudCdnDomain",
				Metadata:   &shared.CloudResourceMetadata{Name: "test"},
				Spec: &AlicloudCdnDomainSpec{
					Region:     "cn-hangzhou",
					DomainName: "cdn.example.com",
					CdnType:    "web",
					Sources: []*AlicloudCdnDomainSource{
						{Type: "ipaddr", Content: ""},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when scope is invalid", func() {
			input := &AlicloudCdnDomain{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudCdnDomain",
				Metadata:   &shared.CloudResourceMetadata{Name: "test"},
				Spec: &AlicloudCdnDomainSpec{
					Region:     "cn-hangzhou",
					DomainName: "cdn.example.com",
					CdnType:    "web",
					Scope:      "invalid",
					Sources:    []*AlicloudCdnDomainSource{validSource()},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when cert_type is invalid", func() {
			input := &AlicloudCdnDomain{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudCdnDomain",
				Metadata:   &shared.CloudResourceMetadata{Name: "test"},
				Spec: &AlicloudCdnDomainSpec{
					Region:     "cn-hangzhou",
					DomainName: "cdn.example.com",
					CdnType:    "web",
					Sources:    []*AlicloudCdnDomainSource{validSource()},
					CertificateConfig: &AlicloudCdnDomainCertificateConfig{
						CertType: "invalid",
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when server_certificate_status is invalid", func() {
			input := &AlicloudCdnDomain{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudCdnDomain",
				Metadata:   &shared.CloudResourceMetadata{Name: "test"},
				Spec: &AlicloudCdnDomainSpec{
					Region:     "cn-hangzhou",
					DomainName: "cdn.example.com",
					CdnType:    "web",
					Sources:    []*AlicloudCdnDomainSource{validSource()},
					CertificateConfig: &AlicloudCdnDomainCertificateConfig{
						ServerCertificateStatus: "invalid",
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when api_version is wrong", func() {
			input := &AlicloudCdnDomain{
				ApiVersion: "wrong/v1",
				Kind:       "AlicloudCdnDomain",
				Metadata:   &shared.CloudResourceMetadata{Name: "test"},
				Spec: &AlicloudCdnDomainSpec{
					Region:     "cn-hangzhou",
					DomainName: "cdn.example.com",
					CdnType:    "web",
					Sources:    []*AlicloudCdnDomainSource{validSource()},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when kind is wrong", func() {
			input := &AlicloudCdnDomain{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "WrongKind",
				Metadata:   &shared.CloudResourceMetadata{Name: "test"},
				Spec: &AlicloudCdnDomainSpec{
					Region:     "cn-hangzhou",
					DomainName: "cdn.example.com",
					CdnType:    "web",
					Sources:    []*AlicloudCdnDomainSource{validSource()},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when metadata is missing", func() {
			input := &AlicloudCdnDomain{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudCdnDomain",
				Spec: &AlicloudCdnDomainSpec{
					Region:     "cn-hangzhou",
					DomainName: "cdn.example.com",
					CdnType:    "web",
					Sources:    []*AlicloudCdnDomainSource{validSource()},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when spec is missing", func() {
			input := &AlicloudCdnDomain{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudCdnDomain",
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
			input := &AlicloudCdnDomain{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudCdnDomain",
				Metadata:   &shared.CloudResourceMetadata{Name: "test"},
				Spec: &AlicloudCdnDomainSpec{
					Region:     "cn-hangzhou",
					DomainName: longDomain,
					CdnType:    "web",
					Sources:    []*AlicloudCdnDomainSource{validSource()},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})
	})
})
