package ocivcnv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared"
	foreignkeyv1 "github.com/plantonhq/openmcf/apis/org/openmcf/shared/foreignkey/v1"
)

func TestOciVcnSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "OciVcnSpec Validation Tests")
}

func newStringValueOrRef(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
			Value: value,
		},
	}
}

func minimalValidVcn() *OciVcn {
	return &OciVcn{
		ApiVersion: "oci.openmcf.org/v1",
		Kind:       "OciVcn",
		Metadata: &shared.CloudResourceMetadata{
			Name: "test-vcn",
		},
		Spec: &OciVcnSpec{
			CompartmentId: newStringValueOrRef("ocid1.compartment.oc1..example"),
			CidrBlocks:    []string{"10.0.0.0/16"},
		},
	}
}

var _ = ginkgo.Describe("OciVcnSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("oci_vcn", func() {

			ginkgo.It("should not return a validation error for minimal valid fields", func() {
				input := minimalValidVcn()
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with all gateways enabled", func() {
				input := minimalValidVcn()
				input.Spec.IsInternetGatewayEnabled = true
				input.Spec.IsNatGatewayEnabled = true
				input.Spec.IsServiceGatewayEnabled = true
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with multiple CIDR blocks", func() {
				input := minimalValidVcn()
				input.Spec.CidrBlocks = []string{"10.0.0.0/16", "172.16.0.0/16"}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with dns_label and display_name", func() {
				input := minimalValidVcn()
				input.Spec.DnsLabel = "myvcn"
				input.Spec.DisplayName = "My Production VCN"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with IPv6 enabled", func() {
				input := minimalValidVcn()
				input.Spec.IsIpv6Enabled = true
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with compartment_id via value_from ref", func() {
				input := minimalValidVcn()
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

			ginkgo.It("should not return a validation error for fully-specified vcn", func() {
				input := &OciVcn{
					ApiVersion: "oci.openmcf.org/v1",
					Kind:       "OciVcn",
					Metadata: &shared.CloudResourceMetadata{
						Name: "full-vcn",
						Org:  "acme-corp",
						Env:  "production",
						Labels: map[string]string{
							"team": "platform",
						},
					},
					Spec: &OciVcnSpec{
						CompartmentId:            newStringValueOrRef("ocid1.compartment.oc1..example"),
						CidrBlocks:               []string{"10.0.0.0/16", "172.16.0.0/16"},
						DisplayName:              "Production VCN",
						DnsLabel:                 "prodvcn",
						IsIpv6Enabled:            true,
						IsInternetGatewayEnabled: true,
						IsNatGatewayEnabled:      true,
						IsServiceGatewayEnabled:  true,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("oci_vcn", func() {

			ginkgo.It("should return a validation error when api_version is wrong", func() {
				input := minimalValidVcn()
				input.ApiVersion = "wrong.openmcf.org/v1"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when kind is wrong", func() {
				input := minimalValidVcn()
				input.Kind = "WrongKind"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when metadata is missing", func() {
				input := minimalValidVcn()
				input.Metadata = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when spec is missing", func() {
				input := &OciVcn{
					ApiVersion: "oci.openmcf.org/v1",
					Kind:       "OciVcn",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-vcn",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when compartment_id is missing", func() {
				input := minimalValidVcn()
				input.Spec.CompartmentId = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when cidr_blocks is empty", func() {
				input := minimalValidVcn()
				input.Spec.CidrBlocks = []string{}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when cidr_blocks is nil", func() {
				input := minimalValidVcn()
				input.Spec.CidrBlocks = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})
	})
})
