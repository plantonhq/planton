package ocidevopsprojectv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared"
	foreignkeyv1 "github.com/plantonhq/openmcf/apis/org/openmcf/shared/foreignkey/v1"
)

func TestOciDevopsProjectSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "OciDevopsProjectSpec Validation Tests")
}

func newStringValueOrRef(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
			Value: value,
		},
	}
}

func minimalValidProject() *OciDevopsProject {
	return &OciDevopsProject{
		ApiVersion: "oci.openmcf.org/v1",
		Kind:       "OciDevopsProject",
		Metadata: &shared.CloudResourceMetadata{
			Name: "test-devops-project",
		},
		Spec: &OciDevopsProjectSpec{
			CompartmentId:       newStringValueOrRef("ocid1.compartment.oc1..example"),
			NotificationTopicId: newStringValueOrRef("ocid1.onstopic.oc1..example"),
		},
	}
}

var _ = ginkgo.Describe("OciDevopsProjectSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("oci_devops_project", func() {

			ginkgo.It("should not return a validation error for minimal valid fields", func() {
				input := minimalValidProject()
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with description", func() {
				input := minimalValidProject()
				input.Spec.Description = "Production CI/CD project for microservices"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with compartment_id via valueFrom ref", func() {
				input := minimalValidProject()
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

			ginkgo.It("should not return a validation error with notification_topic_id via valueFrom ref", func() {
				input := minimalValidProject()
				input.Spec.NotificationTopicId = &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
						ValueFrom: &foreignkeyv1.ValueFromRef{
							Name: "my-ons-topic",
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with both refs via valueFrom", func() {
				input := minimalValidProject()
				input.Spec.CompartmentId = &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
						ValueFrom: &foreignkeyv1.ValueFromRef{
							Name: "my-compartment",
						},
					},
				}
				input.Spec.NotificationTopicId = &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
						ValueFrom: &foreignkeyv1.ValueFromRef{
							Name: "my-ons-topic",
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with empty description", func() {
				input := minimalValidProject()
				input.Spec.Description = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("oci_devops_project", func() {

			ginkgo.It("should return a validation error when api_version is wrong", func() {
				input := minimalValidProject()
				input.ApiVersion = "wrong.openmcf.org/v1"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when kind is wrong", func() {
				input := minimalValidProject()
				input.Kind = "WrongKind"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when metadata is missing", func() {
				input := minimalValidProject()
				input.Metadata = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when spec is missing", func() {
				input := &OciDevopsProject{
					ApiVersion: "oci.openmcf.org/v1",
					Kind:       "OciDevopsProject",
					Metadata:   &shared.CloudResourceMetadata{Name: "test-devops-project"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when compartment_id is missing", func() {
				input := minimalValidProject()
				input.Spec.CompartmentId = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when notification_topic_id is missing", func() {
				input := minimalValidProject()
				input.Spec.NotificationTopicId = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when api_version is empty", func() {
				input := minimalValidProject()
				input.ApiVersion = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when kind is empty", func() {
				input := minimalValidProject()
				input.Kind = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

		})
	})
})
