package gcpdnsrecordv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/apis/dev/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/apis/dev/planton/shared/foreignkey/v1"
)

func TestGcpDnsRecordSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "GcpDnsRecordSpec Validation Tests")
}

var _ = ginkgo.Describe("GcpDnsRecordSpec Validation Tests", func() {

	ginkgo.Describe("Valid configurations", func() {
		ginkgo.Context("minimal valid configuration", func() {

			ginkgo.It("should not return a validation error for valid A record", func() {
				input := &GcpDnsRecord{
					ApiVersion: "gcp.planton.dev/v1",
					Kind:       "GcpDnsRecord",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-dns-record",
					},
					Spec: &GcpDnsRecordSpec{
						ProjectId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "test-project-123"},
						},
						ManagedZone: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "example-zone"},
						},
						Type:   GcpDnsRecordSpec_A,
						Name:   "www.example.com.",
						Values: []string{"192.0.2.1"},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for valid CNAME record", func() {
				input := &GcpDnsRecord{
					ApiVersion: "gcp.planton.dev/v1",
					Kind:       "GcpDnsRecord",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-cname-record",
					},
					Spec: &GcpDnsRecordSpec{
						ProjectId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "test-project-123"},
						},
						ManagedZone: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "example-zone"},
						},
						Type:   GcpDnsRecordSpec_CNAME,
						Name:   "alias.example.com.",
						Values: []string{"target.example.com."},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for multiple values (round-robin)", func() {
				input := &GcpDnsRecord{
					ApiVersion: "gcp.planton.dev/v1",
					Kind:       "GcpDnsRecord",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-roundrobin-record",
					},
					Spec: &GcpDnsRecordSpec{
						ProjectId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "test-project-123"},
						},
						ManagedZone: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "example-zone"},
						},
						Type:   GcpDnsRecordSpec_A,
						Name:   "api.example.com.",
						Values: []string{"192.0.2.1", "192.0.2.2", "192.0.2.3"},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with custom TTL", func() {
				ttl := int32(3600)
				input := &GcpDnsRecord{
					ApiVersion: "gcp.planton.dev/v1",
					Kind:       "GcpDnsRecord",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-ttl-record",
					},
					Spec: &GcpDnsRecordSpec{
						ProjectId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "test-project-123"},
						},
						ManagedZone: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "example-zone"},
						},
						Type:       GcpDnsRecordSpec_TXT,
						Name:       "example.com.",
						Values:     []string{"v=spf1 include:_spf.google.com ~all"},
						TtlSeconds: &ttl,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for wildcard record", func() {
				input := &GcpDnsRecord{
					ApiVersion: "gcp.planton.dev/v1",
					Kind:       "GcpDnsRecord",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-wildcard-record",
					},
					Spec: &GcpDnsRecordSpec{
						ProjectId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "test-project-123"},
						},
						ManagedZone: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "example-zone"},
						},
						Type:   GcpDnsRecordSpec_A,
						Name:   "*.example.com.",
						Values: []string{"192.0.2.1"},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("Invalid configurations", func() {
		ginkgo.Context("missing required fields", func() {

			ginkgo.It("should return a validation error when project_id is missing", func() {
				input := &GcpDnsRecord{
					ApiVersion: "gcp.planton.dev/v1",
					Kind:       "GcpDnsRecord",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-dns-record",
					},
					Spec: &GcpDnsRecordSpec{
						ManagedZone: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "example-zone"},
						},
						Type:   GcpDnsRecordSpec_A,
						Name:   "www.example.com.",
						Values: []string{"192.0.2.1"},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when managed_zone is missing", func() {
				input := &GcpDnsRecord{
					ApiVersion: "gcp.planton.dev/v1",
					Kind:       "GcpDnsRecord",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-dns-record",
					},
					Spec: &GcpDnsRecordSpec{
						ProjectId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "test-project-123"},
						},
						Type:   GcpDnsRecordSpec_A,
						Name:   "www.example.com.",
						Values: []string{"192.0.2.1"},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when record_type is unspecified", func() {
				input := &GcpDnsRecord{
					ApiVersion: "gcp.planton.dev/v1",
					Kind:       "GcpDnsRecord",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-dns-record",
					},
					Spec: &GcpDnsRecordSpec{
						ProjectId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "test-project-123"},
						},
						ManagedZone: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "example-zone"},
						},
						Type:   GcpDnsRecordSpec_record_type_unspecified,
						Name:   "www.example.com.",
						Values: []string{"192.0.2.1"},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when name is missing", func() {
				input := &GcpDnsRecord{
					ApiVersion: "gcp.planton.dev/v1",
					Kind:       "GcpDnsRecord",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-dns-record",
					},
					Spec: &GcpDnsRecordSpec{
						ProjectId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "test-project-123"},
						},
						ManagedZone: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "example-zone"},
						},
						Type:   GcpDnsRecordSpec_A,
						Values: []string{"192.0.2.1"},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when values is empty", func() {
				input := &GcpDnsRecord{
					ApiVersion: "gcp.planton.dev/v1",
					Kind:       "GcpDnsRecord",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-dns-record",
					},
					Spec: &GcpDnsRecordSpec{
						ProjectId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "test-project-123"},
						},
						ManagedZone: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "example-zone"},
						},
						Type:   GcpDnsRecordSpec_A,
						Name:   "www.example.com.",
						Values: []string{},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("invalid field formats", func() {

			ginkgo.It("should return a validation error for name without trailing dot", func() {
				input := &GcpDnsRecord{
					ApiVersion: "gcp.planton.dev/v1",
					Kind:       "GcpDnsRecord",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-dns-record",
					},
					Spec: &GcpDnsRecordSpec{
						ProjectId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "test-project-123"},
						},
						ManagedZone: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "example-zone"},
						},
						Type:   GcpDnsRecordSpec_A,
						Name:   "www.example.com",
						Values: []string{"192.0.2.1"},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for TTL less than 1", func() {
				ttl := int32(0)
				input := &GcpDnsRecord{
					ApiVersion: "gcp.planton.dev/v1",
					Kind:       "GcpDnsRecord",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-dns-record",
					},
					Spec: &GcpDnsRecordSpec{
						ProjectId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "test-project-123"},
						},
						ManagedZone: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "example-zone"},
						},
						Type:       GcpDnsRecordSpec_A,
						Name:       "www.example.com.",
						Values:     []string{"192.0.2.1"},
						TtlSeconds: &ttl,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for TTL greater than 86400", func() {
				ttl := int32(100000)
				input := &GcpDnsRecord{
					ApiVersion: "gcp.planton.dev/v1",
					Kind:       "GcpDnsRecord",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-dns-record",
					},
					Spec: &GcpDnsRecordSpec{
						ProjectId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "test-project-123"},
						},
						ManagedZone: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "example-zone"},
						},
						Type:       GcpDnsRecordSpec_A,
						Name:       "www.example.com.",
						Values:     []string{"192.0.2.1"},
						TtlSeconds: &ttl,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})
	})
})
