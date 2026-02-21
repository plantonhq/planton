package alicloudprivatednszonev1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared"
	fkv1 "github.com/plantonhq/openmcf/apis/org/openmcf/shared/foreignkey/v1"
)

func TestAliCloudPrivateDnsZoneSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AliCloudPrivateDnsZoneSpec Validation Tests")
}

var _ = ginkgo.Describe("AliCloudPrivateDnsZoneSpec Validation Tests", func() {

	ginkgo.Describe("valid input", func() {

		ginkgo.It("should pass with minimal required fields", func() {
			input := &AliCloudPrivateDnsZone{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AliCloudPrivateDnsZone",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test-zone",
				},
				Spec: &AliCloudPrivateDnsZoneSpec{
					Region:   "cn-hangzhou",
					ZoneName: "internal.example.com",
					VpcAttachments: []*AliCloudPrivateDnsZoneVpcAttachment{
						{
							VpcId: &fkv1.StringValueOrRef{
								LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vpc-abc123"},
							},
						},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with all optional fields populated", func() {
			input := &AliCloudPrivateDnsZone{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AliCloudPrivateDnsZone",
				Metadata: &shared.CloudResourceMetadata{
					Name: "full-zone",
					Org:  "my-org",
					Env:  "production",
				},
				Spec: &AliCloudPrivateDnsZoneSpec{
					Region:          "cn-shanghai",
					ZoneName:        "db.internal.corp",
					Remark:          "Database service discovery zone for production",
					ResourceGroupId: "rg-prod-456",
					VpcAttachments: []*AliCloudPrivateDnsZoneVpcAttachment{
						{
							VpcId: &fkv1.StringValueOrRef{
								LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vpc-app"},
							},
						},
						{
							VpcId: &fkv1.StringValueOrRef{
								LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vpc-mgmt"},
							},
							RegionId: "cn-hangzhou",
						},
					},
					Records: []*AliCloudPrivateDnsZoneRecord{
						{
							Rr:    "master",
							Type:  "A",
							Value: "10.0.1.100",
							Ttl:   120,
						},
						{
							Rr:    "replica",
							Type:  "A",
							Value: "10.0.2.100",
						},
					},
					Tags: map[string]string{"team": "platform", "service": "database"},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with multiple VPC attachments including cross-region", func() {
			input := &AliCloudPrivateDnsZone{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AliCloudPrivateDnsZone",
				Metadata: &shared.CloudResourceMetadata{
					Name: "multi-vpc-zone",
				},
				Spec: &AliCloudPrivateDnsZoneSpec{
					Region:   "cn-hangzhou",
					ZoneName: "services.internal",
					VpcAttachments: []*AliCloudPrivateDnsZoneVpcAttachment{
						{
							VpcId: &fkv1.StringValueOrRef{
								LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vpc-hangzhou"},
							},
						},
						{
							VpcId: &fkv1.StringValueOrRef{
								LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vpc-shanghai"},
							},
							RegionId: "cn-shanghai",
						},
						{
							VpcId: &fkv1.StringValueOrRef{
								LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vpc-singapore"},
							},
							RegionId: "ap-southeast-1",
						},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with MX record and priority", func() {
			input := &AliCloudPrivateDnsZone{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AliCloudPrivateDnsZone",
				Metadata: &shared.CloudResourceMetadata{
					Name: "mx-zone",
				},
				Spec: &AliCloudPrivateDnsZoneSpec{
					Region:   "cn-hangzhou",
					ZoneName: "mail.internal",
					VpcAttachments: []*AliCloudPrivateDnsZoneVpcAttachment{
						{
							VpcId: &fkv1.StringValueOrRef{
								LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vpc-123"},
							},
						},
					},
					Records: []*AliCloudPrivateDnsZoneRecord{
						{
							Rr:       "@",
							Type:     "MX",
							Value:    "mail.internal",
							Priority: 10,
						},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with all supported record types", func() {
			input := &AliCloudPrivateDnsZone{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AliCloudPrivateDnsZone",
				Metadata: &shared.CloudResourceMetadata{
					Name: "all-types",
				},
				Spec: &AliCloudPrivateDnsZoneSpec{
					Region:   "cn-hangzhou",
					ZoneName: "test.internal",
					VpcAttachments: []*AliCloudPrivateDnsZoneVpcAttachment{
						{
							VpcId: &fkv1.StringValueOrRef{
								LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vpc-123"},
							},
						},
					},
					Records: []*AliCloudPrivateDnsZoneRecord{
						{Rr: "web", Type: "A", Value: "10.0.1.1"},
						{Rr: "alias", Type: "CNAME", Value: "web.test.internal"},
						{Rr: "@", Type: "MX", Value: "mail.test.internal", Priority: 5},
						{Rr: "100.1.0.10.in-addr.arpa", Type: "PTR", Value: "web.test.internal"},
						{Rr: "_sip._tcp", Type: "SRV", Value: "10 60 5060 sip.test.internal"},
						{Rr: "@", Type: "TXT", Value: "v=spf1 -all"},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with records that have remarks", func() {
			input := &AliCloudPrivateDnsZone{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AliCloudPrivateDnsZone",
				Metadata: &shared.CloudResourceMetadata{
					Name: "remarked-zone",
				},
				Spec: &AliCloudPrivateDnsZoneSpec{
					Region:   "us-west-1",
					ZoneName: "svc.internal",
					VpcAttachments: []*AliCloudPrivateDnsZoneVpcAttachment{
						{
							VpcId: &fkv1.StringValueOrRef{
								LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vpc-456"},
							},
						},
					},
					Records: []*AliCloudPrivateDnsZoneRecord{
						{
							Rr:     "api",
							Type:   "A",
							Value:  "10.0.3.50",
							Ttl:    300,
							Remark: "API gateway internal endpoint",
						},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("invalid input", func() {

		ginkgo.It("should fail when region is missing", func() {
			input := &AliCloudPrivateDnsZone{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AliCloudPrivateDnsZone",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test",
				},
				Spec: &AliCloudPrivateDnsZoneSpec{
					ZoneName: "internal.example.com",
					VpcAttachments: []*AliCloudPrivateDnsZoneVpcAttachment{
						{
							VpcId: &fkv1.StringValueOrRef{
								LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vpc-123"},
							},
						},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when zone_name is missing", func() {
			input := &AliCloudPrivateDnsZone{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AliCloudPrivateDnsZone",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test",
				},
				Spec: &AliCloudPrivateDnsZoneSpec{
					Region: "cn-hangzhou",
					VpcAttachments: []*AliCloudPrivateDnsZoneVpcAttachment{
						{
							VpcId: &fkv1.StringValueOrRef{
								LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vpc-123"},
							},
						},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when vpc_attachments is empty", func() {
			input := &AliCloudPrivateDnsZone{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AliCloudPrivateDnsZone",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test",
				},
				Spec: &AliCloudPrivateDnsZoneSpec{
					Region:         "cn-hangzhou",
					ZoneName:       "internal.example.com",
					VpcAttachments: []*AliCloudPrivateDnsZoneVpcAttachment{},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when vpc_id in attachment is missing", func() {
			input := &AliCloudPrivateDnsZone{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AliCloudPrivateDnsZone",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test",
				},
				Spec: &AliCloudPrivateDnsZoneSpec{
					Region:   "cn-hangzhou",
					ZoneName: "internal.example.com",
					VpcAttachments: []*AliCloudPrivateDnsZoneVpcAttachment{
						{},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when record type is invalid", func() {
			input := &AliCloudPrivateDnsZone{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AliCloudPrivateDnsZone",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test",
				},
				Spec: &AliCloudPrivateDnsZoneSpec{
					Region:   "cn-hangzhou",
					ZoneName: "internal.example.com",
					VpcAttachments: []*AliCloudPrivateDnsZoneVpcAttachment{
						{
							VpcId: &fkv1.StringValueOrRef{
								LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vpc-123"},
							},
						},
					},
					Records: []*AliCloudPrivateDnsZoneRecord{
						{
							Rr:    "www",
							Type:  "AAAA",
							Value: "2001:db8::1",
						},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when record rr is empty", func() {
			input := &AliCloudPrivateDnsZone{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AliCloudPrivateDnsZone",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test",
				},
				Spec: &AliCloudPrivateDnsZoneSpec{
					Region:   "cn-hangzhou",
					ZoneName: "internal.example.com",
					VpcAttachments: []*AliCloudPrivateDnsZoneVpcAttachment{
						{
							VpcId: &fkv1.StringValueOrRef{
								LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vpc-123"},
							},
						},
					},
					Records: []*AliCloudPrivateDnsZoneRecord{
						{
							Type:  "A",
							Value: "10.0.1.1",
						},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when record value is empty", func() {
			input := &AliCloudPrivateDnsZone{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AliCloudPrivateDnsZone",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test",
				},
				Spec: &AliCloudPrivateDnsZoneSpec{
					Region:   "cn-hangzhou",
					ZoneName: "internal.example.com",
					VpcAttachments: []*AliCloudPrivateDnsZoneVpcAttachment{
						{
							VpcId: &fkv1.StringValueOrRef{
								LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vpc-123"},
							},
						},
					},
					Records: []*AliCloudPrivateDnsZoneRecord{
						{
							Rr:   "web",
							Type: "A",
						},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when api_version is wrong", func() {
			input := &AliCloudPrivateDnsZone{
				ApiVersion: "wrong/v1",
				Kind:       "AliCloudPrivateDnsZone",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test",
				},
				Spec: &AliCloudPrivateDnsZoneSpec{
					Region:   "cn-hangzhou",
					ZoneName: "internal.example.com",
					VpcAttachments: []*AliCloudPrivateDnsZoneVpcAttachment{
						{
							VpcId: &fkv1.StringValueOrRef{
								LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vpc-123"},
							},
						},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when kind is wrong", func() {
			input := &AliCloudPrivateDnsZone{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "WrongKind",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test",
				},
				Spec: &AliCloudPrivateDnsZoneSpec{
					Region:   "cn-hangzhou",
					ZoneName: "internal.example.com",
					VpcAttachments: []*AliCloudPrivateDnsZoneVpcAttachment{
						{
							VpcId: &fkv1.StringValueOrRef{
								LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vpc-123"},
							},
						},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when metadata is missing", func() {
			input := &AliCloudPrivateDnsZone{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AliCloudPrivateDnsZone",
				Spec: &AliCloudPrivateDnsZoneSpec{
					Region:   "cn-hangzhou",
					ZoneName: "internal.example.com",
					VpcAttachments: []*AliCloudPrivateDnsZoneVpcAttachment{
						{
							VpcId: &fkv1.StringValueOrRef{
								LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vpc-123"},
							},
						},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when spec is missing", func() {
			input := &AliCloudPrivateDnsZone{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AliCloudPrivateDnsZone",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when zone_name exceeds max length", func() {
			longName := ""
			for i := 0; i < 254; i++ {
				longName += "a"
			}
			input := &AliCloudPrivateDnsZone{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AliCloudPrivateDnsZone",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test",
				},
				Spec: &AliCloudPrivateDnsZoneSpec{
					Region:   "cn-hangzhou",
					ZoneName: longName,
					VpcAttachments: []*AliCloudPrivateDnsZoneVpcAttachment{
						{
							VpcId: &fkv1.StringValueOrRef{
								LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vpc-123"},
							},
						},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})
	})
})
