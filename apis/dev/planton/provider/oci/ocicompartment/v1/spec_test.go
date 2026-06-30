package ocicompartmentv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/apis/dev/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/apis/dev/planton/shared/foreignkey/v1"
)

func TestOciCompartmentSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "OciCompartmentSpec Validation Tests")
}

func newStringValueOrRef(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
			Value: value,
		},
	}
}

func minimalValidCompartment() *OciCompartment {
	return &OciCompartment{
		ApiVersion: "oci.planton.dev/v1",
		Kind:       "OciCompartment",
		Metadata: &shared.CloudResourceMetadata{
			Name: "test-compartment",
		},
		Spec: &OciCompartmentSpec{
			CompartmentId: newStringValueOrRef("ocid1.compartment.oc1..example"),
			Description:   "Test compartment for validation",
		},
	}
}

var _ = ginkgo.Describe("OciCompartmentSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("oci_identity_compartment", func() {

			ginkgo.It("should not return a validation error for minimal valid fields", func() {
				input := minimalValidCompartment()
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with a custom name", func() {
				input := minimalValidCompartment()
				input.Spec.Name = "my-network-compartment"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with enable_delete set to true", func() {
				input := minimalValidCompartment()
				input.Spec.EnableDelete = true
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with compartment_id via value_from ref", func() {
				input := minimalValidCompartment()
				input.Spec.CompartmentId = &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
						ValueFrom: &foreignkeyv1.ValueFromRef{
							Name: "parent-compartment",
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for fully-specified compartment", func() {
				input := &OciCompartment{
					ApiVersion: "oci.planton.dev/v1",
					Kind:       "OciCompartment",
					Metadata: &shared.CloudResourceMetadata{
						Name: "full-compartment",
						Org:  "acme-corp",
						Env:  "production",
						Labels: map[string]string{
							"team": "platform",
						},
					},
					Spec: &OciCompartmentSpec{
						CompartmentId: newStringValueOrRef("ocid1.compartment.oc1..example"),
						Name:          "platform-production",
						Description:   "Production compartment for the platform team",
						EnableDelete:  false,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("oci_identity_compartment", func() {

			ginkgo.It("should return a validation error when api_version is wrong", func() {
				input := minimalValidCompartment()
				input.ApiVersion = "wrong.planton.dev/v1"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when kind is wrong", func() {
				input := minimalValidCompartment()
				input.Kind = "WrongKind"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when metadata is missing", func() {
				input := minimalValidCompartment()
				input.Metadata = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when spec is missing", func() {
				input := &OciCompartment{
					ApiVersion: "oci.planton.dev/v1",
					Kind:       "OciCompartment",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-compartment",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when compartment_id is missing", func() {
				input := minimalValidCompartment()
				input.Spec.CompartmentId = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when description is empty", func() {
				input := minimalValidCompartment()
				input.Spec.Description = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})
	})
})
