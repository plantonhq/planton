package ocikmskeyv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared"
	foreignkeyv1 "github.com/plantonhq/openmcf/apis/org/openmcf/shared/foreignkey/v1"
)

func TestOciKmsKeySpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "OciKmsKeySpec Validation Tests")
}

func newStringValueOrRef(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
			Value: value,
		},
	}
}

func minimalValidAesKey() *OciKmsKey {
	return &OciKmsKey{
		ApiVersion: "oci.openmcf.org/v1",
		Kind:       "OciKmsKey",
		Metadata: &shared.CloudResourceMetadata{
			Name: "test-key",
		},
		Spec: &OciKmsKeySpec{
			CompartmentId:      newStringValueOrRef("ocid1.compartment.oc1..example"),
			ManagementEndpoint: newStringValueOrRef("https://mgmt-vault.kms.us-ashburn-1.oraclecloud.com"),
			KeyShape: &OciKmsKeySpec_KeyShape{
				Algorithm: OciKmsKeySpec_KeyShape_aes,
				Length:    32,
			},
		},
	}
}

func minimalValidRsaKey() *OciKmsKey {
	return &OciKmsKey{
		ApiVersion: "oci.openmcf.org/v1",
		Kind:       "OciKmsKey",
		Metadata: &shared.CloudResourceMetadata{
			Name: "test-rsa-key",
		},
		Spec: &OciKmsKeySpec{
			CompartmentId:      newStringValueOrRef("ocid1.compartment.oc1..example"),
			ManagementEndpoint: newStringValueOrRef("https://mgmt-vault.kms.us-ashburn-1.oraclecloud.com"),
			KeyShape: &OciKmsKeySpec_KeyShape{
				Algorithm: OciKmsKeySpec_KeyShape_rsa,
				Length:    256,
			},
		},
	}
}

func minimalValidEcdsaKey() *OciKmsKey {
	return &OciKmsKey{
		ApiVersion: "oci.openmcf.org/v1",
		Kind:       "OciKmsKey",
		Metadata: &shared.CloudResourceMetadata{
			Name: "test-ecdsa-key",
		},
		Spec: &OciKmsKeySpec{
			CompartmentId:      newStringValueOrRef("ocid1.compartment.oc1..example"),
			ManagementEndpoint: newStringValueOrRef("https://mgmt-vault.kms.us-ashburn-1.oraclecloud.com"),
			KeyShape: &OciKmsKeySpec_KeyShape{
				Algorithm: OciKmsKeySpec_KeyShape_ecdsa,
				Length:    32,
				CurveId:   OciKmsKeySpec_KeyShape_nist_p256,
			},
		},
	}
}

func minimalValidExternalKey() *OciKmsKey {
	return &OciKmsKey{
		ApiVersion: "oci.openmcf.org/v1",
		Kind:       "OciKmsKey",
		Metadata: &shared.CloudResourceMetadata{
			Name: "test-external-key",
		},
		Spec: &OciKmsKeySpec{
			CompartmentId:      newStringValueOrRef("ocid1.compartment.oc1..example"),
			ManagementEndpoint: newStringValueOrRef("https://mgmt-vault.kms.us-ashburn-1.oraclecloud.com"),
			KeyShape: &OciKmsKeySpec_KeyShape{
				Algorithm: OciKmsKeySpec_KeyShape_aes,
				Length:    32,
			},
			ProtectionMode: OciKmsKeySpec_external,
			ExternalKeyReference: &OciKmsKeySpec_ExternalKeyReference{
				ExternalKeyId: "ext-key-abc-123",
			},
		},
	}
}

var _ = ginkgo.Describe("OciKmsKeySpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("oci_kms_key", func() {

			ginkgo.It("should accept a minimal AES-256 key", func() {
				input := minimalValidAesKey()
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept AES-128 (length 16)", func() {
				input := minimalValidAesKey()
				input.Spec.KeyShape.Length = 16
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept AES-192 (length 24)", func() {
				input := minimalValidAesKey()
				input.Spec.KeyShape.Length = 24
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept RSA-2048 (length 256)", func() {
				input := minimalValidRsaKey()
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept RSA-3072 (length 384)", func() {
				input := minimalValidRsaKey()
				input.Spec.KeyShape.Length = 384
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept RSA-4096 (length 512)", func() {
				input := minimalValidRsaKey()
				input.Spec.KeyShape.Length = 512
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept ECDSA P-256 (curve nist_p256, length 32)", func() {
				input := minimalValidEcdsaKey()
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept ECDSA P-384 (curve nist_p384, length 48)", func() {
				input := minimalValidEcdsaKey()
				input.Spec.KeyShape.CurveId = OciKmsKeySpec_KeyShape_nist_p384
				input.Spec.KeyShape.Length = 48
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept ECDSA P-521 (curve nist_p521, length 66)", func() {
				input := minimalValidEcdsaKey()
				input.Spec.KeyShape.CurveId = OciKmsKeySpec_KeyShape_nist_p521
				input.Spec.KeyShape.Length = 66
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept explicit HSM protection mode", func() {
				input := minimalValidAesKey()
				input.Spec.ProtectionMode = OciKmsKeySpec_hsm
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept SOFTWARE protection mode", func() {
				input := minimalValidAesKey()
				input.Spec.ProtectionMode = OciKmsKeySpec_software
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept EXTERNAL protection mode with external_key_reference", func() {
				input := minimalValidExternalKey()
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept auto-rotation enabled with rotation details", func() {
				input := minimalValidAesKey()
				input.Spec.IsAutoRotationEnabled = true
				input.Spec.AutoKeyRotationDetails = &OciKmsKeySpec_AutoKeyRotationDetails{
					RotationIntervalInDays: 90,
					TimeOfScheduleStart:    "2026-04-01T00:00:00Z",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept auto-rotation enabled without explicit details", func() {
				input := minimalValidAesKey()
				input.Spec.IsAutoRotationEnabled = true
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept display_name set", func() {
				input := minimalValidAesKey()
				input.Spec.DisplayName = "my-encryption-key"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept compartment_id via valueFrom ref", func() {
				input := minimalValidAesKey()
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

			ginkgo.It("should accept management_endpoint via valueFrom ref", func() {
				input := minimalValidAesKey()
				input.Spec.ManagementEndpoint = &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
						ValueFrom: &foreignkeyv1.ValueFromRef{
							Name: "my-vault",
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("oci_kms_key", func() {

			ginkgo.It("should reject wrong api_version", func() {
				input := minimalValidAesKey()
				input.ApiVersion = "wrong.openmcf.org/v1"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject wrong kind", func() {
				input := minimalValidAesKey()
				input.Kind = "WrongKind"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject missing metadata", func() {
				input := minimalValidAesKey()
				input.Metadata = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject missing spec", func() {
				input := &OciKmsKey{
					ApiVersion: "oci.openmcf.org/v1",
					Kind:       "OciKmsKey",
					Metadata:   &shared.CloudResourceMetadata{Name: "test"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject missing compartment_id", func() {
				input := minimalValidAesKey()
				input.Spec.CompartmentId = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject missing management_endpoint", func() {
				input := minimalValidAesKey()
				input.Spec.ManagementEndpoint = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject missing key_shape", func() {
				input := minimalValidAesKey()
				input.Spec.KeyShape = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject key_shape with unspecified algorithm", func() {
				input := minimalValidAesKey()
				input.Spec.KeyShape.Algorithm = OciKmsKeySpec_KeyShape_unspecified
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject key_shape with zero length", func() {
				input := minimalValidAesKey()
				input.Spec.KeyShape.Length = 0
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject invalid AES length (64)", func() {
				input := minimalValidAesKey()
				input.Spec.KeyShape.Length = 64
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject invalid RSA length (128)", func() {
				input := minimalValidRsaKey()
				input.Spec.KeyShape.Length = 128
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject ECDSA without curve_id", func() {
				input := minimalValidEcdsaKey()
				input.Spec.KeyShape.CurveId = OciKmsKeySpec_KeyShape_curve_unspecified
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject ECDSA with mismatched curve_id and length (P-256 with length 48)", func() {
				input := minimalValidEcdsaKey()
				input.Spec.KeyShape.CurveId = OciKmsKeySpec_KeyShape_nist_p256
				input.Spec.KeyShape.Length = 48
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject AES with curve_id set", func() {
				input := minimalValidAesKey()
				input.Spec.KeyShape.CurveId = OciKmsKeySpec_KeyShape_nist_p256
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject RSA with curve_id set", func() {
				input := minimalValidRsaKey()
				input.Spec.KeyShape.CurveId = OciKmsKeySpec_KeyShape_nist_p384
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject EXTERNAL protection mode without external_key_reference", func() {
				input := minimalValidAesKey()
				input.Spec.ProtectionMode = OciKmsKeySpec_external
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject HSM with external_key_reference set", func() {
				input := minimalValidAesKey()
				input.Spec.ProtectionMode = OciKmsKeySpec_hsm
				input.Spec.ExternalKeyReference = &OciKmsKeySpec_ExternalKeyReference{
					ExternalKeyId: "ext-key-abc",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject unspecified protection mode with external_key_reference set", func() {
				input := minimalValidAesKey()
				input.Spec.ExternalKeyReference = &OciKmsKeySpec_ExternalKeyReference{
					ExternalKeyId: "ext-key-abc",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject external_key_reference with empty external_key_id", func() {
				input := minimalValidExternalKey()
				input.Spec.ExternalKeyReference.ExternalKeyId = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject auto_key_rotation_details when auto-rotation is not enabled", func() {
				input := minimalValidAesKey()
				input.Spec.AutoKeyRotationDetails = &OciKmsKeySpec_AutoKeyRotationDetails{
					RotationIntervalInDays: 90,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})
	})
})
