package ocidnszonev1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared"
	foreignkeyv1 "github.com/plantonhq/openmcf/apis/org/openmcf/shared/foreignkey/v1"
)

func TestOciDnsZoneSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "OciDnsZoneSpec Validation Tests")
}

func newStringValueOrRef(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
			Value: value,
		},
	}
}

func minimalValidZone() *OciDnsZone {
	return &OciDnsZone{
		ApiVersion: "oci.openmcf.org/v1",
		Kind:       "OciDnsZone",
		Metadata: &shared.CloudResourceMetadata{
			Name: "example-com",
		},
		Spec: &OciDnsZoneSpec{
			CompartmentId: newStringValueOrRef("ocid1.compartment.oc1..example"),
			ZoneType:      OciDnsZoneSpec_primary,
		},
	}
}

var _ = ginkgo.Describe("OciDnsZoneSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("oci_dns_zone", func() {

			ginkgo.It("should not return a validation error for minimal primary global zone", func() {
				input := minimalValidZone()
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with explicit global scope", func() {
				input := minimalValidZone()
				input.Spec.Scope = OciDnsZoneSpec_global
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for primary private zone with view_id", func() {
				input := minimalValidZone()
				input.Spec.Scope = OciDnsZoneSpec_scope_private
				input.Spec.ViewId = newStringValueOrRef("ocid1.dnsview.oc1..example")
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for secondary zone with external masters", func() {
				input := minimalValidZone()
				input.Spec.ZoneType = OciDnsZoneSpec_secondary
				input.Spec.ExternalMasters = []*OciDnsZoneSpec_ExternalServer{
					{Address: "192.168.1.10"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for primary zone with external downstreams", func() {
				input := minimalValidZone()
				input.Spec.ExternalDownstreams = []*OciDnsZoneSpec_ExternalServer{
					{Address: "10.0.0.5"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with DNSSEC enabled", func() {
				input := minimalValidZone()
				dnssec := true
				input.Spec.IsDnssecEnabled = &dnssec
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with DNSSEC explicitly disabled", func() {
				input := minimalValidZone()
				dnssec := false
				input.Spec.IsDnssecEnabled = &dnssec
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with multiple external masters with TSIG keys", func() {
				input := minimalValidZone()
				input.Spec.ZoneType = OciDnsZoneSpec_secondary
				port := int32(53)
				input.Spec.ExternalMasters = []*OciDnsZoneSpec_ExternalServer{
					{
						Address:   "192.168.1.10",
						Port:      &port,
						TsigKeyId: "ocid1.tsigkey.oc1..key1",
					},
					{
						Address:   "192.168.1.11",
						TsigKeyId: "ocid1.tsigkey.oc1..key2",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with both external masters and downstreams on primary", func() {
				input := minimalValidZone()
				input.Spec.ExternalMasters = []*OciDnsZoneSpec_ExternalServer{
					{Address: "10.0.0.1"},
				}
				input.Spec.ExternalDownstreams = []*OciDnsZoneSpec_ExternalServer{
					{Address: "10.0.0.2"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with compartment_id via valueFrom ref", func() {
				input := minimalValidZone()
				input.Spec.CompartmentId = &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
						ValueFrom: &foreignkeyv1.ValueFromRef{
							Name: "my-compartment",
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with view_id via valueFrom ref", func() {
				input := minimalValidZone()
				input.Spec.Scope = OciDnsZoneSpec_scope_private
				input.Spec.ViewId = &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
						ValueFrom: &foreignkeyv1.ValueFromRef{
							Name: "my-dns-view",
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with full configuration", func() {
				input := minimalValidZone()
				input.Spec.Scope = OciDnsZoneSpec_global
				dnssec := true
				input.Spec.IsDnssecEnabled = &dnssec
				port := int32(53)
				input.Spec.ExternalDownstreams = []*OciDnsZoneSpec_ExternalServer{
					{
						Address:   "10.0.0.5",
						Port:      &port,
						TsigKeyId: "ocid1.tsigkey.oc1..downstream",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("oci_dns_zone", func() {

			ginkgo.It("should return a validation error when api_version is wrong", func() {
				input := minimalValidZone()
				input.ApiVersion = "wrong.openmcf.org/v1"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when kind is wrong", func() {
				input := minimalValidZone()
				input.Kind = "WrongKind"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when metadata is missing", func() {
				input := minimalValidZone()
				input.Metadata = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when spec is missing", func() {
				input := &OciDnsZone{
					ApiVersion: "oci.openmcf.org/v1",
					Kind:       "OciDnsZone",
					Metadata:   &shared.CloudResourceMetadata{Name: "test-zone"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when compartment_id is missing", func() {
				input := minimalValidZone()
				input.Spec.CompartmentId = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when zone_type is unspecified", func() {
				input := minimalValidZone()
				input.Spec.ZoneType = OciDnsZoneSpec_unspecified
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for private zone without view_id", func() {
				input := minimalValidZone()
				input.Spec.Scope = OciDnsZoneSpec_scope_private
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for secondary zone with private scope", func() {
				input := minimalValidZone()
				input.Spec.ZoneType = OciDnsZoneSpec_secondary
				input.Spec.Scope = OciDnsZoneSpec_scope_private
				input.Spec.ViewId = newStringValueOrRef("ocid1.dnsview.oc1..example")
				input.Spec.ExternalMasters = []*OciDnsZoneSpec_ExternalServer{
					{Address: "10.0.0.1"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for secondary zone without external masters", func() {
				input := minimalValidZone()
				input.Spec.ZoneType = OciDnsZoneSpec_secondary
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when external master address is empty", func() {
				input := minimalValidZone()
				input.Spec.ZoneType = OciDnsZoneSpec_secondary
				input.Spec.ExternalMasters = []*OciDnsZoneSpec_ExternalServer{
					{Address: ""},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when api_version is empty", func() {
				input := minimalValidZone()
				input.ApiVersion = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when kind is empty", func() {
				input := minimalValidZone()
				input.Kind = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

		})
	})
})
