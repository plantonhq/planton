package ociblockvolumev1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/apis/dev/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/apis/dev/planton/shared/foreignkey/v1"
)

func TestOciBlockVolumeSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "OciBlockVolumeSpec Validation Tests")
}

func newStringValueOrRef(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
			Value: value,
		},
	}
}

func int32Ptr(v int32) *int32 {
	return &v
}

func minimalValidBlockVolume() *OciBlockVolume {
	return &OciBlockVolume{
		ApiVersion: "oci.planton.dev/v1",
		Kind:       "OciBlockVolume",
		Metadata: &shared.CloudResourceMetadata{
			Name: "test-volume",
		},
		Spec: &OciBlockVolumeSpec{
			CompartmentId:      newStringValueOrRef("ocid1.compartment.oc1..example"),
			AvailabilityDomain: "Uocm:US-ASHBURN-AD-1",
			SizeInGbs:          50,
		},
	}
}

var _ = ginkgo.Describe("OciBlockVolumeSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("oci_core_volume", func() {

			ginkgo.It("should not return a validation error for minimal valid fields", func() {
				input := minimalValidBlockVolume()
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with display_name set", func() {
				input := minimalValidBlockVolume()
				input.Spec.DisplayName = "data-volume-01"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with size_in_gbs at 50 (minimum)", func() {
				input := minimalValidBlockVolume()
				input.Spec.SizeInGbs = 50
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with size_in_gbs at 32768 (maximum)", func() {
				input := minimalValidBlockVolume()
				input.Spec.SizeInGbs = 32768
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with vpus_per_gb = 0 (Lower Cost)", func() {
				input := minimalValidBlockVolume()
				input.Spec.VpusPerGb = int32Ptr(0)
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with vpus_per_gb = 10 (Balanced)", func() {
				input := minimalValidBlockVolume()
				input.Spec.VpusPerGb = int32Ptr(10)
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with vpus_per_gb = 20 (Higher Performance)", func() {
				input := minimalValidBlockVolume()
				input.Spec.VpusPerGb = int32Ptr(20)
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with kms_key_id for encryption", func() {
				input := minimalValidBlockVolume()
				input.Spec.KmsKeyId = newStringValueOrRef("ocid1.key.oc1..example")
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with is_reservations_enabled = true", func() {
				input := minimalValidBlockVolume()
				input.Spec.IsReservationsEnabled = true
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with detached_volume autotune policy", func() {
				input := minimalValidBlockVolume()
				input.Spec.AutotunePolicies = []*OciBlockVolumeSpec_AutotunePolicy{
					{AutotuneType: OciBlockVolumeSpec_AutotunePolicy_detached_volume},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with performance_based autotune policy", func() {
				input := minimalValidBlockVolume()
				input.Spec.AutotunePolicies = []*OciBlockVolumeSpec_AutotunePolicy{
					{
						AutotuneType: OciBlockVolumeSpec_AutotunePolicy_performance_based,
						MaxVpusPerGb: 40,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with block_volume_replicas", func() {
				input := minimalValidBlockVolume()
				input.Spec.BlockVolumeReplicas = []*OciBlockVolumeSpec_BlockVolumeReplica{
					{
						AvailabilityDomain: "Uocm:US-PHOENIX-AD-1",
						DisplayName:        "dr-replica",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with backup_policy_id", func() {
				input := minimalValidBlockVolume()
				input.Spec.BackupPolicyId = newStringValueOrRef("ocid1.volumebackuppolicy.oc1..example")
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with xrc_kms_key_id", func() {
				input := minimalValidBlockVolume()
				input.Spec.XrcKmsKeyId = newStringValueOrRef("ocid1.key.oc1..xrcexample")
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with compartment_id via valueFrom ref", func() {
				input := minimalValidBlockVolume()
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

			ginkgo.It("should not return a validation error with full configuration", func() {
				input := minimalValidBlockVolume()
				input.Spec.DisplayName = "production-data-vol"
				input.Spec.SizeInGbs = 1024
				input.Spec.VpusPerGb = int32Ptr(20)
				input.Spec.KmsKeyId = newStringValueOrRef("ocid1.key.oc1..example")
				input.Spec.IsReservationsEnabled = true
				input.Spec.AutotunePolicies = []*OciBlockVolumeSpec_AutotunePolicy{
					{
						AutotuneType: OciBlockVolumeSpec_AutotunePolicy_performance_based,
						MaxVpusPerGb: 60,
					},
				}
				input.Spec.BlockVolumeReplicas = []*OciBlockVolumeSpec_BlockVolumeReplica{
					{
						AvailabilityDomain: "Uocm:US-PHOENIX-AD-1",
						DisplayName:        "dr-replica",
						XrrKmsKeyId:        newStringValueOrRef("ocid1.key.oc1..xrrexample"),
					},
				}
				input.Spec.BackupPolicyId = newStringValueOrRef("ocid1.volumebackuppolicy.oc1..example")
				input.Spec.XrcKmsKeyId = newStringValueOrRef("ocid1.key.oc1..xrcexample")
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("oci_core_volume", func() {

			ginkgo.It("should return a validation error when api_version is wrong", func() {
				input := minimalValidBlockVolume()
				input.ApiVersion = "wrong.planton.dev/v1"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when kind is wrong", func() {
				input := minimalValidBlockVolume()
				input.Kind = "WrongKind"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when metadata is missing", func() {
				input := minimalValidBlockVolume()
				input.Metadata = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when spec is missing", func() {
				input := &OciBlockVolume{
					ApiVersion: "oci.planton.dev/v1",
					Kind:       "OciBlockVolume",
					Metadata:   &shared.CloudResourceMetadata{Name: "test-volume"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when compartment_id is missing", func() {
				input := minimalValidBlockVolume()
				input.Spec.CompartmentId = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when availability_domain is empty", func() {
				input := minimalValidBlockVolume()
				input.Spec.AvailabilityDomain = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when size_in_gbs is 0 (unset)", func() {
				input := minimalValidBlockVolume()
				input.Spec.SizeInGbs = 0
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when size_in_gbs is below minimum (49)", func() {
				input := minimalValidBlockVolume()
				input.Spec.SizeInGbs = 49
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for performance_based autotune without max_vpus_per_gb", func() {
				input := minimalValidBlockVolume()
				input.Spec.AutotunePolicies = []*OciBlockVolumeSpec_AutotunePolicy{
					{AutotuneType: OciBlockVolumeSpec_AutotunePolicy_performance_based},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for autotune with unspecified type", func() {
				input := minimalValidBlockVolume()
				input.Spec.AutotunePolicies = []*OciBlockVolumeSpec_AutotunePolicy{
					{AutotuneType: OciBlockVolumeSpec_AutotunePolicy_unspecified},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when block_volume_replica availability_domain is empty", func() {
				input := minimalValidBlockVolume()
				input.Spec.BlockVolumeReplicas = []*OciBlockVolumeSpec_BlockVolumeReplica{
					{AvailabilityDomain: ""},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})
	})
})
