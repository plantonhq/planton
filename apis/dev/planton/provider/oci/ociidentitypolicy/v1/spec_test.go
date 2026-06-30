package ociidentitypolicyv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/apis/dev/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/apis/dev/planton/shared/foreignkey/v1"
)

func TestOciIdentityPolicySpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "OciIdentityPolicySpec Validation Tests")
}

func newStringValueOrRef(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
			Value: value,
		},
	}
}

func minimalValidPolicy() *OciIdentityPolicy {
	return &OciIdentityPolicy{
		ApiVersion: "oci.planton.dev/v1",
		Kind:       "OciIdentityPolicy",
		Metadata: &shared.CloudResourceMetadata{
			Name: "test-policy",
		},
		Spec: &OciIdentityPolicySpec{
			CompartmentId: newStringValueOrRef("ocid1.compartment.oc1..example"),
			Description:   "Test policy for validation",
			Statements: []string{
				"Allow group TestGroup to read all-resources in compartment TestCompartment",
			},
		},
	}
}

var _ = ginkgo.Describe("OciIdentityPolicySpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("oci_identity_policy", func() {

			ginkgo.It("should not return a validation error for minimal valid fields", func() {
				input := minimalValidPolicy()
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with a custom name", func() {
				input := minimalValidPolicy()
				input.Spec.Name = "network-admin-policy"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with version_date set", func() {
				input := minimalValidPolicy()
				input.Spec.VersionDate = "2026-02-19"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with multiple statements", func() {
				input := minimalValidPolicy()
				input.Spec.Statements = []string{
					"Allow group Admins to manage all-resources in compartment Production",
					"Allow group Developers to use instances in compartment Development",
					"Allow group Auditors to inspect all-resources in tenancy",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with compartment_id via value_from ref", func() {
				input := minimalValidPolicy()
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

			ginkgo.It("should not return a validation error for fully-specified policy", func() {
				input := &OciIdentityPolicy{
					ApiVersion: "oci.planton.dev/v1",
					Kind:       "OciIdentityPolicy",
					Metadata: &shared.CloudResourceMetadata{
						Name: "full-policy",
						Org:  "acme-corp",
						Env:  "production",
						Labels: map[string]string{
							"team": "platform",
						},
					},
					Spec: &OciIdentityPolicySpec{
						CompartmentId: newStringValueOrRef("ocid1.compartment.oc1..example"),
						Name:          "platform-admin-policy",
						Description:   "Admin policy for the platform team",
						Statements: []string{
							"Allow group PlatformAdmins to manage all-resources in compartment Platform",
						},
						VersionDate: "2026-01-01",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("oci_identity_policy", func() {

			ginkgo.It("should return a validation error when api_version is wrong", func() {
				input := minimalValidPolicy()
				input.ApiVersion = "wrong.planton.dev/v1"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when kind is wrong", func() {
				input := minimalValidPolicy()
				input.Kind = "WrongKind"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when metadata is missing", func() {
				input := minimalValidPolicy()
				input.Metadata = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when spec is missing", func() {
				input := &OciIdentityPolicy{
					ApiVersion: "oci.planton.dev/v1",
					Kind:       "OciIdentityPolicy",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-policy",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when compartment_id is missing", func() {
				input := minimalValidPolicy()
				input.Spec.CompartmentId = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when description is empty", func() {
				input := minimalValidPolicy()
				input.Spec.Description = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when statements is empty", func() {
				input := minimalValidPolicy()
				input.Spec.Statements = []string{}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when statements is nil", func() {
				input := minimalValidPolicy()
				input.Spec.Statements = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})
	})
})
