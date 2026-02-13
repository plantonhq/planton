package azurednsrecordv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared"
	foreignkeyv1 "github.com/plantonhq/openmcf/apis/org/openmcf/shared/foreignkey/v1"
	"google.golang.org/protobuf/proto"
)

func TestAzureDnsRecordSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureDnsRecordSpec Custom Validation Tests")
}

func stringRef(s string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: s}}
}

var _ = ginkgo.Describe("AzureDnsRecordSpec Custom Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("azure_dns_record", func() {

			ginkgo.It("should not return a validation error for minimal valid A record", func() {
				input := &AzureDnsRecord{
					ApiVersion: "azure.openmcf.org/v1",
					Kind:       "AzureDnsRecord",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-a-record",
					},
					Spec: &AzureDnsRecordSpec{
						ResourceGroup: stringRef("test-resource-group"),
						ZoneName: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "example.com",
							},
						},
						Type:   AzureDnsRecordSpec_A,
						Name:   "www",
						Values: []string{"192.0.2.1"},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for zone apex record using @", func() {
				input := &AzureDnsRecord{
					ApiVersion: "azure.openmcf.org/v1",
					Kind:       "AzureDnsRecord",
					Metadata: &shared.CloudResourceMetadata{
						Name: "apex-record",
					},
					Spec: &AzureDnsRecordSpec{
						ResourceGroup: stringRef("test-resource-group"),
						ZoneName: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "example.com",
							},
						},
						Type:   AzureDnsRecordSpec_A,
						Name:   "@",
						Values: []string{"192.0.2.1"},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for wildcard record", func() {
				input := &AzureDnsRecord{
					ApiVersion: "azure.openmcf.org/v1",
					Kind:       "AzureDnsRecord",
					Metadata: &shared.CloudResourceMetadata{
						Name: "wildcard-record",
					},
					Spec: &AzureDnsRecordSpec{
						ResourceGroup: stringRef("test-resource-group"),
						ZoneName: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "example.com",
							},
						},
						Type:   AzureDnsRecordSpec_A,
						Name:   "*",
						Values: []string{"192.0.2.1"},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for MX record with priority", func() {
				input := &AzureDnsRecord{
					ApiVersion: "azure.openmcf.org/v1",
					Kind:       "AzureDnsRecord",
					Metadata: &shared.CloudResourceMetadata{
						Name: "mx-record",
					},
					Spec: &AzureDnsRecordSpec{
						ResourceGroup: stringRef("test-resource-group"),
						ZoneName: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "example.com",
							},
						},
						Type:       AzureDnsRecordSpec_MX,
						Name:       "@",
						Values:     []string{"mail.example.com"},
						MxPriority: proto.Int32(10),
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for CNAME record", func() {
				input := &AzureDnsRecord{
					ApiVersion: "azure.openmcf.org/v1",
					Kind:       "AzureDnsRecord",
					Metadata: &shared.CloudResourceMetadata{
						Name: "cname-record",
					},
					Spec: &AzureDnsRecordSpec{
						ResourceGroup: stringRef("test-resource-group"),
						ZoneName: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "example.com",
							},
						},
						Type:   AzureDnsRecordSpec_CNAME,
						Name:   "alias",
						Values: []string{"target.example.com"},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for TXT record", func() {
				input := &AzureDnsRecord{
					ApiVersion: "azure.openmcf.org/v1",
					Kind:       "AzureDnsRecord",
					Metadata: &shared.CloudResourceMetadata{
						Name: "txt-record",
					},
					Spec: &AzureDnsRecordSpec{
						ResourceGroup: stringRef("test-resource-group"),
						ZoneName: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "example.com",
							},
						},
						Type:   AzureDnsRecordSpec_TXT,
						Name:   "@",
						Values: []string{"v=spf1 include:_spf.google.com ~all"},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for nested subdomain", func() {
				input := &AzureDnsRecord{
					ApiVersion: "azure.openmcf.org/v1",
					Kind:       "AzureDnsRecord",
					Metadata: &shared.CloudResourceMetadata{
						Name: "nested-record",
					},
					Spec: &AzureDnsRecordSpec{
						ResourceGroup: stringRef("test-resource-group"),
						ZoneName: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "example.com",
							},
						},
						Type:   AzureDnsRecordSpec_A,
						Name:   "api.v1",
						Values: []string{"192.0.2.1"},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for zone_name using value_from", func() {
				input := &AzureDnsRecord{
					ApiVersion: "azure.openmcf.org/v1",
					Kind:       "AzureDnsRecord",
					Metadata: &shared.CloudResourceMetadata{
						Name: "ref-record",
					},
					Spec: &AzureDnsRecordSpec{
						ResourceGroup: stringRef("test-resource-group"),
						ZoneName: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
								ValueFrom: &foreignkeyv1.ValueFromRef{
									Name: "my-azure-zone",
								},
							},
						},
						Type:   AzureDnsRecordSpec_A,
						Name:   "www",
						Values: []string{"192.0.2.1"},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("azure_dns_record", func() {

			ginkgo.It("should return a validation error when resource_group is missing", func() {
				input := &AzureDnsRecord{
					ApiVersion: "azure.openmcf.org/v1",
					Kind:       "AzureDnsRecord",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-record",
					},
					Spec: &AzureDnsRecordSpec{
						ZoneName: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "example.com",
							},
						},
						Type:   AzureDnsRecordSpec_A,
						Name:   "www",
						Values: []string{"192.0.2.1"},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when zone_name is missing", func() {
				input := &AzureDnsRecord{
					ApiVersion: "azure.openmcf.org/v1",
					Kind:       "AzureDnsRecord",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-record",
					},
					Spec: &AzureDnsRecordSpec{
						ResourceGroup: stringRef("test-resource-group"),
						Type:          AzureDnsRecordSpec_A,
						Name:          "www",
						Values:        []string{"192.0.2.1"},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when record_type is unspecified", func() {
				input := &AzureDnsRecord{
					ApiVersion: "azure.openmcf.org/v1",
					Kind:       "AzureDnsRecord",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-record",
					},
					Spec: &AzureDnsRecordSpec{
						ResourceGroup: stringRef("test-resource-group"),
						ZoneName: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "example.com",
							},
						},
						Type:   AzureDnsRecordSpec_record_type_unspecified,
						Name:   "www",
						Values: []string{"192.0.2.1"},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when name is missing", func() {
				input := &AzureDnsRecord{
					ApiVersion: "azure.openmcf.org/v1",
					Kind:       "AzureDnsRecord",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-record",
					},
					Spec: &AzureDnsRecordSpec{
						ResourceGroup: stringRef("test-resource-group"),
						ZoneName: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "example.com",
							},
						},
						Type:   AzureDnsRecordSpec_A,
						Values: []string{"192.0.2.1"},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when values is empty", func() {
				input := &AzureDnsRecord{
					ApiVersion: "azure.openmcf.org/v1",
					Kind:       "AzureDnsRecord",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-record",
					},
					Spec: &AzureDnsRecordSpec{
						ResourceGroup: stringRef("test-resource-group"),
						ZoneName: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "example.com",
							},
						},
						Type:   AzureDnsRecordSpec_A,
						Name:   "www",
						Values: []string{},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when mx_priority is set for non-MX record", func() {
				input := &AzureDnsRecord{
					ApiVersion: "azure.openmcf.org/v1",
					Kind:       "AzureDnsRecord",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-record",
					},
					Spec: &AzureDnsRecordSpec{
						ResourceGroup: stringRef("test-resource-group"),
						ZoneName: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "example.com",
							},
						},
						Type:       AzureDnsRecordSpec_A,
						Name:       "www",
						Values:     []string{"192.0.2.1"},
						MxPriority: proto.Int32(10),
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for invalid TTL (0)", func() {
				input := &AzureDnsRecord{
					ApiVersion: "azure.openmcf.org/v1",
					Kind:       "AzureDnsRecord",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-record",
					},
					Spec: &AzureDnsRecordSpec{
						ResourceGroup: stringRef("test-resource-group"),
						ZoneName: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "example.com",
							},
						},
						Type:       AzureDnsRecordSpec_A,
						Name:       "www",
						Values:     []string{"192.0.2.1"},
						TtlSeconds: proto.Int32(0),
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})
	})
})
