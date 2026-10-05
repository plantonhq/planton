package kubernetessecretstorev1alpha1

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

func TestKubernetesSecretStore(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "KubernetesSecretStore Suite")
}

func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

func valueFrom(kind catalogkind.CatalogKind, name, fieldPath string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
			ValueFrom: &foreignkeyv1.ValueFromRef{
				Kind:      kind,
				Name:      name,
				FieldPath: fieldPath,
			},
		},
	}
}

var _ = ginkgo.Describe("KubernetesSecretStore Validation Tests", func() {
	var input *KubernetesSecretStore

	ginkgo.BeforeEach(func() {
		input = &KubernetesSecretStore{
			ApiVersion: "kubernetes.planton.dev/v1alpha1",
			Kind:       "KubernetesSecretStore",
			Metadata:   &shared.CatalogObjectMetadata{Name: "team-a-gcp"},
			Spec: &KubernetesSecretStoreSpec{
				Namespace: literal("team-a"),
				Config: &kubernetesprovider.ExternalSecretsStoreConfig{
					Backend: &kubernetesprovider.ExternalSecretsStoreConfig_GcpSecretManager{
						GcpSecretManager: &kubernetesprovider.ExternalSecretsStoreGcp{
							ProjectId: literal("my-project"),
						},
					},
				},
			},
		}
	})

	ginkgo.Describe("valid configurations", func() {
		ginkgo.It("accepts a keyless GCP Secret Manager store", func() {
			gomega.Expect(protovalidate.Validate(input)).To(gomega.Succeed())
		})

		ginkgo.It("accepts an Azure Key Vault store with a service principal", func() {
			authType := "ServicePrincipal"
			input.Spec.Config = &kubernetesprovider.ExternalSecretsStoreConfig{
				Backend: &kubernetesprovider.ExternalSecretsStoreConfig_AzureKeyVault{
					AzureKeyVault: &kubernetesprovider.ExternalSecretsStoreAzure{
						VaultUrl:     literal("https://my-vault.vault.azure.net"),
						TenantId:     "00000000-0000-0000-0000-000000000000",
						AuthType:     &authType,
						ClientId:     "11111111-1111-1111-1111-111111111111",
						ClientSecret: "sp-secret",
					},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.Succeed())
		})

		ginkgo.It("accepts a Vault AppRole store with retry tuning", func() {
			maxRetries := int32(3)
			input.Spec.Config = &kubernetesprovider.ExternalSecretsStoreConfig{
				Backend: &kubernetesprovider.ExternalSecretsStoreConfig_Vault{
					Vault: &kubernetesprovider.ExternalSecretsStoreVault{
						Server: "https://openbao.example.com:8200",
						Auth: &kubernetesprovider.ExternalSecretsStoreVault_AppRole{
							AppRole: &kubernetesprovider.ExternalSecretsStoreVaultAppRoleAuth{
								RoleId:   "role-id",
								SecretId: "secret-id",
							},
						},
					},
				},
				Retry: &kubernetesprovider.ExternalSecretsStoreRetry{
					MaxRetries:    &maxRetries,
					RetryInterval: "10s",
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.Succeed())
		})
	})

	ginkgo.Describe("required fields and contracts", func() {
		ginkgo.It("rejects a missing namespace", func() {
			input.Spec.Namespace = nil
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.Succeed())
		})

		ginkgo.It("rejects a missing config", func() {
			input.Spec.Config = nil
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.Succeed())
		})

		ginkgo.It("rejects a GCP store mixing a key with a ServiceAccount identity", func() {
			gcp := input.Spec.Config.GetGcpSecretManager()
			gcp.ServiceAccountName = literal("secrets-reader")
			gcp.ServiceAccountKeyJson = `{"type":"service_account"}`
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.Succeed())
		})

		ginkgo.It("rejects an Azure service principal missing its secret", func() {
			authType := "ServicePrincipal"
			input.Spec.Config = &kubernetesprovider.ExternalSecretsStoreConfig{
				Backend: &kubernetesprovider.ExternalSecretsStoreConfig_AzureKeyVault{
					AzureKeyVault: &kubernetesprovider.ExternalSecretsStoreAzure{
						VaultUrl: literal("https://my-vault.vault.azure.net"),
						TenantId: "00000000-0000-0000-0000-000000000000",
						AuthType: &authType,
					},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.Succeed())
		})

		ginkgo.It("accepts a vault_url referencing an AzureKeyVault", func() {
			// The composed-environment shape: the vault deploys in the same
			// run and its data-plane URI flows in by reference.
			input.Spec.Config = &kubernetesprovider.ExternalSecretsStoreConfig{
				Backend: &kubernetesprovider.ExternalSecretsStoreConfig_AzureKeyVault{
					AzureKeyVault: &kubernetesprovider.ExternalSecretsStoreAzure{
						VaultUrl: valueFrom(
							catalogkind.CatalogKind_AzureKeyVault, "platform-kv", "status.outputs.vault_uri"),
						TenantId: "00000000-0000-0000-0000-000000000000",
					},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.Succeed())
		})

		ginkgo.It("rejects an Azure store without a vault_url", func() {
			input.Spec.Config = &kubernetesprovider.ExternalSecretsStoreConfig{
				Backend: &kubernetesprovider.ExternalSecretsStoreConfig_AzureKeyVault{
					AzureKeyVault: &kubernetesprovider.ExternalSecretsStoreAzure{},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.Succeed())
		})

		ginkgo.It("rejects a Vault token auth without the token", func() {
			input.Spec.Config = &kubernetesprovider.ExternalSecretsStoreConfig{
				Backend: &kubernetesprovider.ExternalSecretsStoreConfig_Vault{
					Vault: &kubernetesprovider.ExternalSecretsStoreVault{
						Server: "https://vault.example.com:8200",
						Auth: &kubernetesprovider.ExternalSecretsStoreVault_Token{
							Token: &kubernetesprovider.ExternalSecretsStoreVaultTokenAuth{},
						},
					},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.Succeed())
		})

		ginkgo.It("rejects a Vault kubernetes auth without a role", func() {
			input.Spec.Config = &kubernetesprovider.ExternalSecretsStoreConfig{
				Backend: &kubernetesprovider.ExternalSecretsStoreConfig_Vault{
					Vault: &kubernetesprovider.ExternalSecretsStoreVault{
						Server: "https://vault.example.com:8200",
						Auth: &kubernetesprovider.ExternalSecretsStoreVault_Kubernetes{
							Kubernetes: &kubernetesprovider.ExternalSecretsStoreVaultKubernetesAuth{},
						},
					},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.Succeed())
		})

		ginkgo.It("rejects a remote-Kubernetes store with both token and ServiceAccount", func() {
			input.Spec.Config = &kubernetesprovider.ExternalSecretsStoreConfig{
				Backend: &kubernetesprovider.ExternalSecretsStoreConfig_Kubernetes{
					Kubernetes: &kubernetesprovider.ExternalSecretsStoreKubernetes{
						Token:              "bearer-token",
						ServiceAccountName: literal("reader"),
					},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.Succeed())
		})

		ginkgo.It("rejects a malformed retry interval", func() {
			input.Spec.Config.Retry = &kubernetesprovider.ExternalSecretsStoreRetry{RetryInterval: "soon"}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.Succeed())
		})
	})
})
