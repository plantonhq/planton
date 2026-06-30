package ocipublicipv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/apis/dev/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/apis/dev/planton/shared/foreignkey/v1"
)

func TestOciPublicIpSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "OciPublicIpSpec Validation Tests")
}

func newStringValueOrRef(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
			Value: value,
		},
	}
}

func minimalValidReservedPublicIp() *OciPublicIp {
	return &OciPublicIp{
		ApiVersion: "oci.planton.dev/v1",
		Kind:       "OciPublicIp",
		Metadata: &shared.CloudResourceMetadata{
			Name: "test-public-ip",
		},
		Spec: &OciPublicIpSpec{
			CompartmentId: newStringValueOrRef("ocid1.compartment.oc1..example"),
			Lifetime:      "RESERVED",
		},
	}
}

func minimalValidEphemeralPublicIp() *OciPublicIp {
	return &OciPublicIp{
		ApiVersion: "oci.planton.dev/v1",
		Kind:       "OciPublicIp",
		Metadata: &shared.CloudResourceMetadata{
			Name: "test-ephemeral-ip",
		},
		Spec: &OciPublicIpSpec{
			CompartmentId: newStringValueOrRef("ocid1.compartment.oc1..example"),
			Lifetime:      "EPHEMERAL",
			PrivateIpId:   newStringValueOrRef("ocid1.privateip.oc1..example"),
		},
	}
}

var _ = ginkgo.Describe("OciPublicIpSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("reserved public ip", func() {

			ginkgo.It("should not return a validation error for minimal reserved IP", func() {
				input := minimalValidReservedPublicIp()
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for reserved IP with display_name", func() {
				input := minimalValidReservedPublicIp()
				input.Spec.DisplayName = "my-static-ip"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for reserved IP assigned to a private IP", func() {
				input := minimalValidReservedPublicIp()
				input.Spec.PrivateIpId = newStringValueOrRef("ocid1.privateip.oc1..example")
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for reserved IP from a BYOIP pool", func() {
				input := minimalValidReservedPublicIp()
				input.Spec.PublicIpPoolId = newStringValueOrRef("ocid1.publicippool.oc1..example")
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with compartment_id via value_from ref", func() {
				input := minimalValidReservedPublicIp()
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

			ginkgo.It("should not return a validation error with private_ip_id via value_from ref", func() {
				input := minimalValidEphemeralPublicIp()
				input.Spec.PrivateIpId = &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
						ValueFrom: &foreignkeyv1.ValueFromRef{
							Name: "my-instance",
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("ephemeral public ip", func() {

			ginkgo.It("should not return a validation error for minimal ephemeral IP with private_ip_id", func() {
				input := minimalValidEphemeralPublicIp()
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("fully specified public ip", func() {

			ginkgo.It("should not return a validation error for fully-specified reserved public IP", func() {
				input := &OciPublicIp{
					ApiVersion: "oci.planton.dev/v1",
					Kind:       "OciPublicIp",
					Metadata: &shared.CloudResourceMetadata{
						Name: "full-public-ip",
						Org:  "acme-corp",
						Env:  "production",
						Labels: map[string]string{
							"team": "networking",
						},
					},
					Spec: &OciPublicIpSpec{
						CompartmentId:  newStringValueOrRef("ocid1.compartment.oc1..example"),
						Lifetime:       "RESERVED",
						DisplayName:    "prod-ingress-ip",
						PrivateIpId:    newStringValueOrRef("ocid1.privateip.oc1..example"),
						PublicIpPoolId: newStringValueOrRef("ocid1.publicippool.oc1..example"),
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("api envelope", func() {

			ginkgo.It("should return a validation error when api_version is wrong", func() {
				input := minimalValidReservedPublicIp()
				input.ApiVersion = "wrong.planton.dev/v1"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when kind is wrong", func() {
				input := minimalValidReservedPublicIp()
				input.Kind = "WrongKind"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when metadata is missing", func() {
				input := minimalValidReservedPublicIp()
				input.Metadata = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when spec is missing", func() {
				input := &OciPublicIp{
					ApiVersion: "oci.planton.dev/v1",
					Kind:       "OciPublicIp",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-public-ip",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("spec fields", func() {

			ginkgo.It("should return a validation error when compartment_id is missing", func() {
				input := minimalValidReservedPublicIp()
				input.Spec.CompartmentId = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when lifetime is empty", func() {
				input := minimalValidReservedPublicIp()
				input.Spec.Lifetime = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when lifetime is invalid", func() {
				input := minimalValidReservedPublicIp()
				input.Spec.Lifetime = "INVALID"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when ephemeral IP has no private_ip_id", func() {
				input := &OciPublicIp{
					ApiVersion: "oci.planton.dev/v1",
					Kind:       "OciPublicIp",
					Metadata: &shared.CloudResourceMetadata{
						Name: "bad-ephemeral-ip",
					},
					Spec: &OciPublicIpSpec{
						CompartmentId: newStringValueOrRef("ocid1.compartment.oc1..example"),
						Lifetime:      "EPHEMERAL",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})
	})
})
