package kubernetesclustersecretstorev1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	kubernetesprovider "github.com/plantonhq/planton/catalog/kubernetes"
	"github.com/plantonhq/planton/shared"
	"github.com/plantonhq/planton/shared/catalogkind"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestKubernetesClusterSecretStore(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "KubernetesClusterSecretStore Suite")
}

func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

func awsBackend() *kubernetesprovider.ExternalSecretsStoreConfig {
	return &kubernetesprovider.ExternalSecretsStoreConfig{
		Backend: &kubernetesprovider.ExternalSecretsStoreConfig_Aws{
			Aws: &kubernetesprovider.ExternalSecretsStoreAws{Region: "us-east-1"},
		},
	}
}

var _ = ginkgo.Describe("KubernetesClusterSecretStore Validation Tests", func() {
	var input *KubernetesClusterSecretStore

	ginkgo.BeforeEach(func() {
		input = &KubernetesClusterSecretStore{
			ApiVersion: "kubernetes.planton.dev/v1alpha1",
			Kind:       "KubernetesClusterSecretStore",
			Metadata:   &shared.CatalogObjectMetadata{Name: "aws-prod"},
			Spec: &KubernetesClusterSecretStoreSpec{
				SecretsNamespace: literal("external-secrets"),
				Config:           awsBackend(),
			},
		}
	})

	ginkgo.Describe("valid configurations", func() {
		ginkgo.It("accepts a keyless AWS Secrets Manager store", func() {
			gomega.Expect(protovalidate.Validate(input)).To(gomega.Succeed())
		})

		ginkgo.It("accepts a store with a ServiceAccount identity", func() {
			input.Spec.Config.GetAws().ServiceAccountName = literal("secrets-reader")
			input.Spec.Config.GetAws().ServiceAccountNamespace = "external-secrets"
			gomega.Expect(protovalidate.Validate(input)).To(gomega.Succeed())
		})

		ginkgo.It("accepts namespace conditions in all three modes", func() {
			input.Spec.Conditions = []*KubernetesClusterSecretStoreCondition{
				{Namespaces: []string{"team-a", "team-b"}},
				{NamespaceLabelSelector: map[string]string{"env": "prod"}},
				{NamespaceRegexes: []string{"^app-.*"}},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.Succeed())
		})

		ginkgo.It("accepts a Vault backend with kubernetes auth", func() {
			input.Spec.Config = &kubernetesprovider.ExternalSecretsStoreConfig{
				Backend: &kubernetesprovider.ExternalSecretsStoreConfig_Vault{
					Vault: &kubernetesprovider.ExternalSecretsStoreVault{
						Server: "https://vault.example.com:8200",
						Auth: &kubernetesprovider.ExternalSecretsStoreVault_Kubernetes{
							Kubernetes: &kubernetesprovider.ExternalSecretsStoreVaultKubernetesAuth{
								Role: "external-secrets",
							},
						},
					},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.Succeed())
		})

		ginkgo.It("accepts an Azure Key Vault store with a referenced vault_url", func() {
			// The composed-environment shape (what a one-run cluster chart
			// renders): the vault's data-plane URI flows in by reference and
			// the reference is also the deploy-ordering edge.
			input.Spec.Config = &kubernetesprovider.ExternalSecretsStoreConfig{
				Backend: &kubernetesprovider.ExternalSecretsStoreConfig_AzureKeyVault{
					AzureKeyVault: &kubernetesprovider.ExternalSecretsStoreAzure{
						VaultUrl: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
								ValueFrom: &foreignkeyv1.ValueFromRef{
									Kind:      catalogkind.CatalogKind_AzureKeyVault,
									Name:      "platform-kv",
									FieldPath: "status.outputs.vault_uri",
								},
							},
						},
						TenantId: "00000000-0000-0000-0000-000000000000",
					},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.Succeed())
		})

		ginkgo.It("accepts a fake backend with entries", func() {
			input.Spec.Config = &kubernetesprovider.ExternalSecretsStoreConfig{
				Backend: &kubernetesprovider.ExternalSecretsStoreConfig_Fake{
					Fake: &kubernetesprovider.ExternalSecretsStoreFake{
						Data: []*kubernetesprovider.ExternalSecretsStoreFakeEntry{
							{Key: "db-password", Value: "s3cret"},
						},
					},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.Succeed())
		})
	})

	ginkgo.Describe("required fields and contracts", func() {
		ginkgo.It("rejects a missing secrets namespace", func() {
			input.Spec.SecretsNamespace = nil
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.Succeed())
		})

		ginkgo.It("rejects a missing config", func() {
			input.Spec.Config = nil
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.Succeed())
		})

		ginkgo.It("rejects a config with no backend", func() {
			input.Spec.Config = &kubernetesprovider.ExternalSecretsStoreConfig{}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.Succeed())
		})

		ginkgo.It("rejects an empty condition", func() {
			input.Spec.Conditions = []*KubernetesClusterSecretStoreCondition{{}}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.Succeed())
		})

		ginkgo.It("rejects an invalid namespace name in a condition", func() {
			input.Spec.Conditions = []*KubernetesClusterSecretStoreCondition{
				{Namespaces: []string{"Team_A"}},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.Succeed())
		})

		ginkgo.It("rejects AWS static keys combined with a ServiceAccount identity", func() {
			aws := input.Spec.Config.GetAws()
			aws.ServiceAccountName = literal("secrets-reader")
			aws.AccessKeyId = "AKIAEXAMPLE"
			aws.SecretAccessKey = "secret"
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.Succeed())
		})

		ginkgo.It("rejects a Vault backend without an auth method", func() {
			input.Spec.Config = &kubernetesprovider.ExternalSecretsStoreConfig{
				Backend: &kubernetesprovider.ExternalSecretsStoreConfig_Vault{
					Vault: &kubernetesprovider.ExternalSecretsStoreVault{
						Server: "https://vault.example.com:8200",
					},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.Succeed())
		})

		ginkgo.It("rejects a fake backend with no entries", func() {
			input.Spec.Config = &kubernetesprovider.ExternalSecretsStoreConfig{
				Backend: &kubernetesprovider.ExternalSecretsStoreConfig_Fake{
					Fake: &kubernetesprovider.ExternalSecretsStoreFake{},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.Succeed())
		})

		ginkgo.It("rejects a malformed store refresh interval", func() {
			input.Spec.Config.RefreshInterval = "five minutes"
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.Succeed())
		})
	})
})
