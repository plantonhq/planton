package azurekeyvaultcertificatev1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestAzureKeyVaultCertificateSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureKeyVaultCertificateSpec Custom Validation Tests")
}

// validSelfSignedSpec is the fully-hands-off internal-TLS shape: vault
// generates and self-signs, auto-renewing at 80% lifetime.
func validSelfSignedSpec() *AzureKeyVaultCertificateSpec {
	return &AzureKeyVaultCertificateSpec{
		Name:       "internal-tls",
		KeyVaultId: stringRef("/subscriptions/sub/resourceGroups/rg/providers/Microsoft.KeyVault/vaults/my-vault"),
		CertificatePolicy: &AzureKeyVaultCertificatePolicy{
			IssuerName: "Self",
			KeyProperties: &AzureKeyVaultCertificateKeyProperties{
				Exportable: true,
				KeyType:    AzureKeyVaultCertificateKeyType_RSA,
				KeySize:    int32Ptr(2048),
				ReuseKey:   false,
			},
			LifetimeActions: []*AzureKeyVaultCertificateLifetimeAction{
				{
					ActionType: AzureKeyVaultCertificateLifetimeActionType_AUTO_RENEW,
					Trigger: &AzureKeyVaultCertificateLifetimeTrigger{
						LifetimePercentage: int32Ptr(80),
					},
				},
			},
			SecretProperties: &AzureKeyVaultCertificateSecretProperties{
				ContentType: AzureKeyVaultCertificateContentType_PKCS12,
			},
			X509CertificateProperties: &AzureKeyVaultCertificateX509Properties{
				Subject: "CN=internal.example.com",
				SubjectAlternativeNames: &AzureKeyVaultCertificateSubjectAlternativeNames{
					DnsNames: []string{"internal.example.com"},
				},
				KeyUsage: []AzureKeyVaultCertificateKeyUsage{
					AzureKeyVaultCertificateKeyUsage_DIGITAL_SIGNATURE,
					AzureKeyVaultCertificateKeyUsage_KEY_ENCIPHERMENT,
				},
				ValidityInMonths: 12,
			},
		},
	}
}

func cert(spec *AzureKeyVaultCertificateSpec) *AzureKeyVaultCertificate {
	return &AzureKeyVaultCertificate{
		ApiVersion: "azure.planton.dev/v1alpha1",
		Kind:       "AzureKeyVaultCertificate",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-certificate",
		},
		Spec: spec,
	}
}

var _ = ginkgo.Describe("AzureKeyVaultCertificateSpec Custom Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("azure_key_vault_certificate", func() {

			ginkgo.It("should accept a self-signed auto-renewing certificate", func() {
				err := protovalidate.Validate(cert(validSelfSignedSpec()))
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept an imported bundle without a policy", func() {
				spec := &AzureKeyVaultCertificateSpec{
					Name:       "imported-tls",
					KeyVaultId: stringRef("/subscriptions/sub/resourceGroups/rg/providers/Microsoft.KeyVault/vaults/my-vault"),
					Certificate: &AzureKeyVaultCertificateImport{
						Contents: "bWljcm9zb2Z0LXBmeC1ieXRlcw==",
						Password: strPtr("bundle-password"),
					},
				}
				err := protovalidate.Validate(cert(spec))
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept an import accompanied by an explicit policy without x509 properties", func() {
				spec := validSelfSignedSpec()
				spec.Certificate = &AzureKeyVaultCertificateImport{
					Contents: "bWljcm9zb2Z0LXBmeC1ieXRlcw==",
				}
				spec.CertificatePolicy.X509CertificateProperties = nil
				err := protovalidate.Validate(cert(spec))
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept an EC certificate key with a curve and no size", func() {
				spec := validSelfSignedSpec()
				spec.CertificatePolicy.KeyProperties = &AzureKeyVaultCertificateKeyProperties{
					Exportable: false,
					KeyType:    AzureKeyVaultCertificateKeyType_EC,
					Curve:      AzureKeyVaultCertificateKeyCurve_P_384,
					ReuseKey:   true,
				}
				err := protovalidate.Validate(cert(spec))
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a days-before-expiry email action for a CA-issued certificate", func() {
				spec := validSelfSignedSpec()
				spec.CertificatePolicy.IssuerName = "Unknown"
				spec.CertificatePolicy.LifetimeActions = []*AzureKeyVaultCertificateLifetimeAction{
					{
						ActionType: AzureKeyVaultCertificateLifetimeActionType_EMAIL_CONTACTS,
						Trigger: &AzureKeyVaultCertificateLifetimeTrigger{
							DaysBeforeExpiry: int32Ptr(30),
						},
					},
				}
				err := protovalidate.Validate(cert(spec))
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept extended key usage OIDs and PEM content type", func() {
				spec := validSelfSignedSpec()
				spec.CertificatePolicy.SecretProperties.ContentType = AzureKeyVaultCertificateContentType_PEM
				spec.CertificatePolicy.X509CertificateProperties.ExtendedKeyUsage = []string{"1.3.6.1.5.5.7.3.1"}
				err := protovalidate.Validate(cert(spec))
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept user tags", func() {
				spec := validSelfSignedSpec()
				spec.Tags = map[string]string{"team": "platform"}
				err := protovalidate.Validate(cert(spec))
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("azure_key_vault_certificate", func() {

			ginkgo.It("should return a validation error when name is missing", func() {
				spec := validSelfSignedSpec()
				spec.Name = ""
				gomega.Expect(protovalidate.Validate(cert(spec))).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when name carries invalid characters", func() {
				spec := validSelfSignedSpec()
				spec.Name = "my.cert"
				gomega.Expect(protovalidate.Validate(cert(spec))).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when key_vault_id is missing", func() {
				spec := validSelfSignedSpec()
				spec.KeyVaultId = nil
				gomega.Expect(protovalidate.Validate(cert(spec))).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when neither import nor policy is provided", func() {
				spec := &AzureKeyVaultCertificateSpec{
					Name:       "empty-cert",
					KeyVaultId: stringRef("/subscriptions/sub/resourceGroups/rg/providers/Microsoft.KeyVault/vaults/my-vault"),
				}
				gomega.Expect(protovalidate.Validate(cert(spec))).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when a generated certificate omits x509 properties", func() {
				spec := validSelfSignedSpec()
				spec.CertificatePolicy.X509CertificateProperties = nil
				gomega.Expect(protovalidate.Validate(cert(spec))).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when an import omits contents", func() {
				spec := &AzureKeyVaultCertificateSpec{
					Name:        "imported-tls",
					KeyVaultId:  stringRef("/subscriptions/sub/resourceGroups/rg/providers/Microsoft.KeyVault/vaults/my-vault"),
					Certificate: &AzureKeyVaultCertificateImport{},
				}
				gomega.Expect(protovalidate.Validate(cert(spec))).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when the policy omits issuer_name", func() {
				spec := validSelfSignedSpec()
				spec.CertificatePolicy.IssuerName = ""
				gomega.Expect(protovalidate.Validate(cert(spec))).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when the policy omits key_properties", func() {
				spec := validSelfSignedSpec()
				spec.CertificatePolicy.KeyProperties = nil
				gomega.Expect(protovalidate.Validate(cert(spec))).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when the policy omits secret_properties", func() {
				spec := validSelfSignedSpec()
				spec.CertificatePolicy.SecretProperties = nil
				gomega.Expect(protovalidate.Validate(cert(spec))).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when key_type is unspecified", func() {
				spec := validSelfSignedSpec()
				spec.CertificatePolicy.KeyProperties.KeyType = AzureKeyVaultCertificateKeyType_azure_key_vault_certificate_key_type_unspecified
				gomega.Expect(protovalidate.Validate(cert(spec))).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for an unsupported key size", func() {
				spec := validSelfSignedSpec()
				spec.CertificatePolicy.KeyProperties.KeySize = int32Ptr(1024)
				gomega.Expect(protovalidate.Validate(cert(spec))).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when content_type is unspecified", func() {
				spec := validSelfSignedSpec()
				spec.CertificatePolicy.SecretProperties.ContentType = AzureKeyVaultCertificateContentType_azure_key_vault_certificate_content_type_unspecified
				gomega.Expect(protovalidate.Validate(cert(spec))).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when a lifetime action carries both triggers", func() {
				spec := validSelfSignedSpec()
				spec.CertificatePolicy.LifetimeActions[0].Trigger = &AzureKeyVaultCertificateLifetimeTrigger{
					DaysBeforeExpiry:   int32Ptr(30),
					LifetimePercentage: int32Ptr(80),
				}
				gomega.Expect(protovalidate.Validate(cert(spec))).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when a lifetime action carries no trigger fields", func() {
				spec := validSelfSignedSpec()
				spec.CertificatePolicy.LifetimeActions[0].Trigger = &AzureKeyVaultCertificateLifetimeTrigger{}
				gomega.Expect(protovalidate.Validate(cert(spec))).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for an out-of-range lifetime percentage", func() {
				spec := validSelfSignedSpec()
				spec.CertificatePolicy.LifetimeActions[0].Trigger = &AzureKeyVaultCertificateLifetimeTrigger{
					LifetimePercentage: int32Ptr(100),
				}
				gomega.Expect(protovalidate.Validate(cert(spec))).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when x509 properties omit the subject", func() {
				spec := validSelfSignedSpec()
				spec.CertificatePolicy.X509CertificateProperties.Subject = ""
				gomega.Expect(protovalidate.Validate(cert(spec))).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when x509 properties omit key_usage", func() {
				spec := validSelfSignedSpec()
				spec.CertificatePolicy.X509CertificateProperties.KeyUsage = nil
				gomega.Expect(protovalidate.Validate(cert(spec))).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for zero validity months", func() {
				spec := validSelfSignedSpec()
				spec.CertificatePolicy.X509CertificateProperties.ValidityInMonths = 0
				gomega.Expect(protovalidate.Validate(cert(spec))).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for an empty SAN block", func() {
				spec := validSelfSignedSpec()
				spec.CertificatePolicy.X509CertificateProperties.SubjectAlternativeNames = &AzureKeyVaultCertificateSubjectAlternativeNames{}
				gomega.Expect(protovalidate.Validate(cert(spec))).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when spec is missing", func() {
				input := cert(validSelfSignedSpec())
				input.Spec = nil
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})
		})
	})
})

// Helper functions for pointer types
func int32Ptr(i int32) *int32 {
	return &i
}

func strPtr(s string) *string {
	return &s
}

func stringRef(s string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: s}}
}
