package ocivaultsecretv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared"
	foreignkeyv1 "github.com/plantonhq/openmcf/apis/org/openmcf/shared/foreignkey/v1"
)

func TestOciVaultSecretSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "OciVaultSecretSpec Validation Tests")
}

func newStringValueOrRef(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
			Value: value,
		},
	}
}

func minimalValidSecret() *OciVaultSecret {
	return &OciVaultSecret{
		ApiVersion: "oci.openmcf.org/v1",
		Kind:       "OciVaultSecret",
		Metadata: &shared.CloudResourceMetadata{
			Name: "test-secret",
		},
		Spec: &OciVaultSecretSpec{
			CompartmentId: newStringValueOrRef("ocid1.compartment.oc1..example"),
			SecretName:    "my-api-key",
			VaultId:       newStringValueOrRef("ocid1.vault.oc1..example"),
			KeyId:         newStringValueOrRef("ocid1.key.oc1..example"),
		},
	}
}

func secretWithExplicitContent() *OciVaultSecret {
	s := minimalValidSecret()
	s.Spec.SecretContent = &OciVaultSecretSpec_SecretContent{
		Content: "dGVzdC1zZWNyZXQtdmFsdWU=",
	}
	return s
}

func secretWithAutoGeneration() *OciVaultSecret {
	s := minimalValidSecret()
	s.Spec.EnableAutoGeneration = true
	s.Spec.SecretGenerationContext = &OciVaultSecretSpec_SecretGenerationContext{
		GenerationType:     OciVaultSecretSpec_SecretGenerationContext_bytes,
		GenerationTemplate: "secret-bytes-template-v1",
	}
	return s
}

var _ = ginkgo.Describe("OciVaultSecretSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.It("should accept a minimal secret (no content, no auto-generation)", func() {
			err := protovalidate.Validate(minimalValidSecret())
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept a secret with explicit base64 content", func() {
			err := protovalidate.Validate(secretWithExplicitContent())
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept a secret with explicit content and stage CURRENT", func() {
			s := secretWithExplicitContent()
			s.Spec.SecretContent.Stage = "CURRENT"
			err := protovalidate.Validate(s)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept a secret with explicit content and stage PENDING", func() {
			s := secretWithExplicitContent()
			s.Spec.SecretContent.Stage = "PENDING"
			err := protovalidate.Validate(s)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept a secret with explicit content and version name", func() {
			s := secretWithExplicitContent()
			s.Spec.SecretContent.Name = "v1-initial"
			err := protovalidate.Validate(s)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept a secret with auto-generation (bytes)", func() {
			err := protovalidate.Validate(secretWithAutoGeneration())
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept a secret with auto-generation (passphrase)", func() {
			s := minimalValidSecret()
			s.Spec.EnableAutoGeneration = true
			s.Spec.SecretGenerationContext = &OciVaultSecretSpec_SecretGenerationContext{
				GenerationType:     OciVaultSecretSpec_SecretGenerationContext_passphrase,
				GenerationTemplate: "dbaas-default-password",
				PassphraseLength:   24,
			}
			err := protovalidate.Validate(s)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept a secret with auto-generation (ssh_key)", func() {
			s := minimalValidSecret()
			s.Spec.EnableAutoGeneration = true
			s.Spec.SecretGenerationContext = &OciVaultSecretSpec_SecretGenerationContext{
				GenerationType:     OciVaultSecretSpec_SecretGenerationContext_ssh_key,
				GenerationTemplate: "rsa-2048",
			}
			err := protovalidate.Validate(s)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept a secret with auto-generation and secret_template", func() {
			s := secretWithAutoGeneration()
			s.Spec.SecretGenerationContext.SecretTemplate = `{"password": "${secret-content}"}`
			err := protovalidate.Validate(s)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept a secret with description", func() {
			s := minimalValidSecret()
			s.Spec.Description = "Production API key for external service"
			err := protovalidate.Validate(s)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept a secret with secret_metadata", func() {
			s := minimalValidSecret()
			s.Spec.SecretMetadata = map[string]string{
				"target-service": "external-api",
				"rotation-notes": "rotate monthly",
			}
			err := protovalidate.Validate(s)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept a secret with an expiry rule", func() {
			s := secretWithExplicitContent()
			s.Spec.SecretRules = []*OciVaultSecretSpec_SecretRule{
				{
					RuleType:                                        OciVaultSecretSpec_SecretRule_secret_expiry_rule,
					SecretVersionExpiryInterval:                     "P30D",
					IsSecretContentRetrievalBlockedOnExpiry: true,
				},
			}
			err := protovalidate.Validate(s)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept a secret with a reuse rule", func() {
			s := secretWithExplicitContent()
			s.Spec.SecretRules = []*OciVaultSecretSpec_SecretRule{
				{
					RuleType:                          OciVaultSecretSpec_SecretRule_secret_reuse_rule,
					IsEnforcedOnDeletedSecretVersions: true,
				},
			}
			err := protovalidate.Validate(s)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept a secret with both expiry and reuse rules", func() {
			s := secretWithExplicitContent()
			s.Spec.SecretRules = []*OciVaultSecretSpec_SecretRule{
				{
					RuleType:                    OciVaultSecretSpec_SecretRule_secret_expiry_rule,
					SecretVersionExpiryInterval: "P7D",
				},
				{
					RuleType:                          OciVaultSecretSpec_SecretRule_secret_reuse_rule,
					IsEnforcedOnDeletedSecretVersions: true,
				},
			}
			err := protovalidate.Validate(s)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept a secret with rotation config (ADB target)", func() {
			s := secretWithExplicitContent()
			s.Spec.RotationConfig = &OciVaultSecretSpec_RotationConfig{
				IsScheduledRotationEnabled: true,
				RotationInterval:           "P30D",
				TargetSystemDetails: &OciVaultSecretSpec_RotationConfig_TargetSystemDetails{
					TargetSystemType: OciVaultSecretSpec_RotationConfig_TargetSystemDetails_adb,
					AdbId:            newStringValueOrRef("ocid1.autonomousdatabase.oc1..example"),
				},
			}
			err := protovalidate.Validate(s)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept a secret with rotation config (Function target)", func() {
			s := secretWithExplicitContent()
			s.Spec.RotationConfig = &OciVaultSecretSpec_RotationConfig{
				IsScheduledRotationEnabled: true,
				RotationInterval:           "P90D",
				TargetSystemDetails: &OciVaultSecretSpec_RotationConfig_TargetSystemDetails{
					TargetSystemType: OciVaultSecretSpec_RotationConfig_TargetSystemDetails_function,
					FunctionId:       newStringValueOrRef("ocid1.fnfunc.oc1..example"),
				},
			}
			err := protovalidate.Validate(s)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept a secret with rotation config (scheduling disabled)", func() {
			s := secretWithExplicitContent()
			s.Spec.RotationConfig = &OciVaultSecretSpec_RotationConfig{
				TargetSystemDetails: &OciVaultSecretSpec_RotationConfig_TargetSystemDetails{
					TargetSystemType: OciVaultSecretSpec_RotationConfig_TargetSystemDetails_function,
					FunctionId:       newStringValueOrRef("ocid1.fnfunc.oc1..example"),
				},
			}
			err := protovalidate.Validate(s)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept a secret with StringValueOrRef using valueFrom", func() {
			s := minimalValidSecret()
			s.Spec.CompartmentId = &foreignkeyv1.StringValueOrRef{
				LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
					ValueFrom: &foreignkeyv1.ValueFromRef{
						Name: "my-compartment",
					},
				},
			}
			err := protovalidate.Validate(s)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept a fully populated secret", func() {
			s := secretWithExplicitContent()
			s.Spec.Description = "Full-featured test secret"
			s.Spec.SecretMetadata = map[string]string{"team": "platform"}
			s.Spec.SecretRules = []*OciVaultSecretSpec_SecretRule{
				{
					RuleType:                    OciVaultSecretSpec_SecretRule_secret_expiry_rule,
					SecretVersionExpiryInterval: "P30D",
				},
			}
			s.Spec.RotationConfig = &OciVaultSecretSpec_RotationConfig{
				IsScheduledRotationEnabled: true,
				RotationInterval:           "P30D",
				TargetSystemDetails: &OciVaultSecretSpec_RotationConfig_TargetSystemDetails{
					TargetSystemType: OciVaultSecretSpec_RotationConfig_TargetSystemDetails_adb,
					AdbId:            newStringValueOrRef("ocid1.autonomousdatabase.oc1..example"),
				},
			}
			err := protovalidate.Validate(s)
			gomega.Expect(err).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.It("should reject when api_version is wrong", func() {
			s := minimalValidSecret()
			s.ApiVersion = "wrong/v1"
			err := protovalidate.Validate(s)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject when kind is wrong", func() {
			s := minimalValidSecret()
			s.Kind = "WrongKind"
			err := protovalidate.Validate(s)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject when metadata is missing", func() {
			s := minimalValidSecret()
			s.Metadata = nil
			err := protovalidate.Validate(s)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject when spec is missing", func() {
			s := minimalValidSecret()
			s.Spec = nil
			err := protovalidate.Validate(s)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject when compartment_id is missing", func() {
			s := minimalValidSecret()
			s.Spec.CompartmentId = nil
			err := protovalidate.Validate(s)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject when secret_name is empty", func() {
			s := minimalValidSecret()
			s.Spec.SecretName = ""
			err := protovalidate.Validate(s)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject when vault_id is missing", func() {
			s := minimalValidSecret()
			s.Spec.VaultId = nil
			err := protovalidate.Validate(s)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject when key_id is missing", func() {
			s := minimalValidSecret()
			s.Spec.KeyId = nil
			err := protovalidate.Validate(s)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject when secret_content is set with enable_auto_generation", func() {
			s := secretWithExplicitContent()
			s.Spec.EnableAutoGeneration = true
			s.Spec.SecretGenerationContext = &OciVaultSecretSpec_SecretGenerationContext{
				GenerationType:     OciVaultSecretSpec_SecretGenerationContext_bytes,
				GenerationTemplate: "template",
			}
			err := protovalidate.Validate(s)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject when enable_auto_generation is true but context is missing", func() {
			s := minimalValidSecret()
			s.Spec.EnableAutoGeneration = true
			err := protovalidate.Validate(s)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject when secret_generation_context is set without enable_auto_generation", func() {
			s := minimalValidSecret()
			s.Spec.SecretGenerationContext = &OciVaultSecretSpec_SecretGenerationContext{
				GenerationType:     OciVaultSecretSpec_SecretGenerationContext_bytes,
				GenerationTemplate: "template",
			}
			err := protovalidate.Validate(s)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject when generation_type is unspecified", func() {
			s := minimalValidSecret()
			s.Spec.EnableAutoGeneration = true
			s.Spec.SecretGenerationContext = &OciVaultSecretSpec_SecretGenerationContext{
				GenerationType:     OciVaultSecretSpec_SecretGenerationContext_unspecified,
				GenerationTemplate: "template",
			}
			err := protovalidate.Validate(s)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject when generation_template is empty", func() {
			s := minimalValidSecret()
			s.Spec.EnableAutoGeneration = true
			s.Spec.SecretGenerationContext = &OciVaultSecretSpec_SecretGenerationContext{
				GenerationType:     OciVaultSecretSpec_SecretGenerationContext_bytes,
				GenerationTemplate: "",
			}
			err := protovalidate.Validate(s)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject when passphrase generation has zero length", func() {
			s := minimalValidSecret()
			s.Spec.EnableAutoGeneration = true
			s.Spec.SecretGenerationContext = &OciVaultSecretSpec_SecretGenerationContext{
				GenerationType:     OciVaultSecretSpec_SecretGenerationContext_passphrase,
				GenerationTemplate: "dbaas-default-password",
				PassphraseLength:   0,
			}
			err := protovalidate.Validate(s)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject when secret_rule has unspecified rule_type", func() {
			s := secretWithExplicitContent()
			s.Spec.SecretRules = []*OciVaultSecretSpec_SecretRule{
				{RuleType: OciVaultSecretSpec_SecretRule_unspecified},
			}
			err := protovalidate.Validate(s)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject when rotation_config has no target_system_details", func() {
			s := secretWithExplicitContent()
			s.Spec.RotationConfig = &OciVaultSecretSpec_RotationConfig{
				IsScheduledRotationEnabled: true,
				RotationInterval:           "P30D",
			}
			err := protovalidate.Validate(s)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject when target_system_type is unspecified", func() {
			s := secretWithExplicitContent()
			s.Spec.RotationConfig = &OciVaultSecretSpec_RotationConfig{
				TargetSystemDetails: &OciVaultSecretSpec_RotationConfig_TargetSystemDetails{
					TargetSystemType: OciVaultSecretSpec_RotationConfig_TargetSystemDetails_unspecified,
				},
			}
			err := protovalidate.Validate(s)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject when stage has an invalid value", func() {
			s := secretWithExplicitContent()
			s.Spec.SecretContent.Stage = "INVALID"
			err := protovalidate.Validate(s)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})
	})
})
