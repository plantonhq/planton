package ocibastionv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/apis/dev/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/apis/dev/planton/shared/foreignkey/v1"
	"google.golang.org/protobuf/proto"
)

func TestOciBastionSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "OciBastionSpec Validation Tests")
}

func newStringValueOrRef(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
			Value: value,
		},
	}
}

func minimalValidBastion() *OciBastion {
	return &OciBastion{
		ApiVersion: "oci.planton.dev/v1",
		Kind:       "OciBastion",
		Metadata: &shared.CloudResourceMetadata{
			Name: "test-bastion",
		},
		Spec: &OciBastionSpec{
			CompartmentId:  newStringValueOrRef("ocid1.compartment.oc1..example"),
			TargetSubnetId: newStringValueOrRef("ocid1.subnet.oc1..example"),
		},
	}
}

var _ = ginkgo.Describe("OciBastionSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("oci_bastion_bastion", func() {

			ginkgo.It("should not return a validation error for minimal bastion", func() {
				input := minimalValidBastion()
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with display_name set", func() {
				input := minimalValidBastion()
				input.Spec.DisplayName = "my-private-bastion"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with client CIDR allow list", func() {
				input := minimalValidBastion()
				input.Spec.ClientCidrBlockAllowList = []string{"10.0.0.0/16", "192.168.1.0/24"}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with max session TTL", func() {
				input := minimalValidBastion()
				input.Spec.MaxSessionTtlInSeconds = proto.Int32(10800)
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with DNS proxy enabled", func() {
				input := minimalValidBastion()
				input.Spec.IsDnsProxyEnabled = proto.Bool(true)
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with DNS proxy disabled", func() {
				input := minimalValidBastion()
				input.Spec.IsDnsProxyEnabled = proto.Bool(false)
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with compartment_id via valueFrom ref", func() {
				input := minimalValidBastion()
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

			ginkgo.It("should not return a validation error with target_subnet_id via valueFrom ref", func() {
				input := minimalValidBastion()
				input.Spec.TargetSubnetId = &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
						ValueFrom: &foreignkeyv1.ValueFromRef{
							Name: "my-subnet",
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with full configuration", func() {
				input := minimalValidBastion()
				input.Spec.DisplayName = "production-bastion"
				input.Spec.ClientCidrBlockAllowList = []string{"10.0.0.0/8"}
				input.Spec.MaxSessionTtlInSeconds = proto.Int32(7200)
				input.Spec.IsDnsProxyEnabled = proto.Bool(true)
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with empty CIDR list", func() {
				input := minimalValidBastion()
				input.Spec.ClientCidrBlockAllowList = []string{}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("oci_bastion_bastion", func() {

			ginkgo.It("should return a validation error when api_version is wrong", func() {
				input := minimalValidBastion()
				input.ApiVersion = "wrong.planton.dev/v1"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when kind is wrong", func() {
				input := minimalValidBastion()
				input.Kind = "WrongKind"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when metadata is missing", func() {
				input := minimalValidBastion()
				input.Metadata = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when spec is missing", func() {
				input := &OciBastion{
					ApiVersion: "oci.planton.dev/v1",
					Kind:       "OciBastion",
					Metadata:   &shared.CloudResourceMetadata{Name: "test-bastion"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when compartment_id is missing", func() {
				input := minimalValidBastion()
				input.Spec.CompartmentId = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when target_subnet_id is missing", func() {
				input := minimalValidBastion()
				input.Spec.TargetSubnetId = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})
	})
})
