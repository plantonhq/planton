package ocidynamicgroupv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared"
	foreignkeyv1 "github.com/plantonhq/openmcf/apis/org/openmcf/shared/foreignkey/v1"
)

func TestOciDynamicGroupSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "OciDynamicGroupSpec Validation Tests")
}

func newStringValueOrRef(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
			Value: value,
		},
	}
}

func minimalValidDynamicGroup() *OciDynamicGroup {
	return &OciDynamicGroup{
		ApiVersion: "oci.openmcf.org/v1",
		Kind:       "OciDynamicGroup",
		Metadata: &shared.CloudResourceMetadata{
			Name: "test-dynamic-group",
		},
		Spec: &OciDynamicGroupSpec{
			CompartmentId: newStringValueOrRef("ocid1.tenancy.oc1..example"),
			Description:   "Dynamic group for compute instances",
			MatchingRule:  "Any {instance.compartment.id = 'ocid1.compartment.oc1..example'}",
		},
	}
}

var _ = ginkgo.Describe("OciDynamicGroupSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("oci_identity_dynamic_group", func() {

			ginkgo.It("should not return a validation error for minimal valid fields", func() {
				input := minimalValidDynamicGroup()
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with a custom name", func() {
				input := minimalValidDynamicGroup()
				input.Spec.Name = "oke-worker-nodes"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with compartment_id via value_from ref", func() {
				input := minimalValidDynamicGroup()
				input.Spec.CompartmentId = &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
						ValueFrom: &foreignkeyv1.ValueFromRef{
							Name: "tenancy-compartment",
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with Any matching rule syntax", func() {
				input := minimalValidDynamicGroup()
				input.Spec.MatchingRule = "Any {instance.compartment.id = 'ocid1.compartment.oc1..xxx'}"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with All matching rule syntax", func() {
				input := minimalValidDynamicGroup()
				input.Spec.MatchingRule = "All {resource.type = 'fnfunc', resource.compartment.id = 'ocid1.compartment.oc1..xxx'}"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for fully-specified dynamic group", func() {
				input := &OciDynamicGroup{
					ApiVersion: "oci.openmcf.org/v1",
					Kind:       "OciDynamicGroup",
					Metadata: &shared.CloudResourceMetadata{
						Name: "full-dynamic-group",
						Org:  "acme-corp",
						Env:  "production",
						Labels: map[string]string{
							"team": "platform",
						},
					},
					Spec: &OciDynamicGroupSpec{
						CompartmentId: newStringValueOrRef("ocid1.tenancy.oc1..example"),
						Name:          "prod-compute-dynamic-group",
						Description:   "Dynamic group for all production compute instances",
						MatchingRule:  "Any {instance.compartment.id = 'ocid1.compartment.oc1..production'}",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("oci_identity_dynamic_group", func() {

			ginkgo.It("should return a validation error when api_version is wrong", func() {
				input := minimalValidDynamicGroup()
				input.ApiVersion = "wrong.openmcf.org/v1"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when kind is wrong", func() {
				input := minimalValidDynamicGroup()
				input.Kind = "WrongKind"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when metadata is missing", func() {
				input := minimalValidDynamicGroup()
				input.Metadata = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when spec is missing", func() {
				input := &OciDynamicGroup{
					ApiVersion: "oci.openmcf.org/v1",
					Kind:       "OciDynamicGroup",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-dynamic-group",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when compartment_id is missing", func() {
				input := minimalValidDynamicGroup()
				input.Spec.CompartmentId = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when description is empty", func() {
				input := minimalValidDynamicGroup()
				input.Spec.Description = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when matching_rule is empty", func() {
				input := minimalValidDynamicGroup()
				input.Spec.MatchingRule = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})
	})
})
