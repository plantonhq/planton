package kubernetesissuerv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	kubernetesprovider "github.com/plantonhq/planton/catalog/kubernetes"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestKubernetesIssuer(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "KubernetesIssuer Suite")
}

func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

// The deep per-backend contract coverage (solver shapes, credential pairing,
// vault auth) lives in the KubernetesClusterIssuer suite — the config message
// is shared, so those rules hold identically here. This suite locks the
// Issuer-specific surface: the namespace FK and one acceptance per backend.
var _ = ginkgo.Describe("KubernetesIssuer Validation Tests", func() {
	var input *KubernetesIssuer

	ginkgo.BeforeEach(func() {
		input = &KubernetesIssuer{
			ApiVersion: "kubernetes.planton.dev/v1alpha1",
			Kind:       "KubernetesIssuer",
			Metadata:   &shared.CatalogObjectMetadata{Name: "test-issuer"},
			Spec: &KubernetesIssuerSpec{
				Namespace: literal("team-a"),
				Config: &kubernetesprovider.CertManagerIssuerConfig{
					Backend: &kubernetesprovider.CertManagerIssuerConfig_SelfSigned{
						SelfSigned: &kubernetesprovider.CertManagerSelfSignedConfig{},
					},
				},
			},
		}
	})

	ginkgo.Describe("valid configurations", func() {
		ginkgo.It("accepts a self-signed backend", func() {
			gomega.Expect(protovalidate.Validate(input)).To(gomega.Succeed())
		})

		ginkgo.It("accepts a CA backend (the internal-PKI pattern)", func() {
			input.Spec.Config = &kubernetesprovider.CertManagerIssuerConfig{
				Backend: &kubernetesprovider.CertManagerIssuerConfig_Ca{
					Ca: &kubernetesprovider.CertManagerCaConfig{CaSecretName: literal("team-a-root-ca")},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.Succeed())
		})

		ginkgo.It("accepts an ACME backend with a solver", func() {
			input.Spec.Config = &kubernetesprovider.CertManagerIssuerConfig{
				Backend: &kubernetesprovider.CertManagerIssuerConfig_Acme{
					Acme: &kubernetesprovider.CertManagerAcmeConfig{
						Email: "team-a@example.com",
						Solvers: []*kubernetesprovider.CertManagerAcmeSolver{{
							Challenge: &kubernetesprovider.CertManagerAcmeSolver_Dns01{
								Dns01: &kubernetesprovider.CertManagerAcmeDns01Solver{
									Provider: &kubernetesprovider.CertManagerAcmeDns01Solver_Digitalocean{
										Digitalocean: &kubernetesprovider.CertManagerDns01DigitalOcean{Token: "do-token"},
									},
								},
							},
						}},
					},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.Succeed())
		})

		ginkgo.It("accepts a Vault backend with AppRole auth", func() {
			input.Spec.Config = &kubernetesprovider.CertManagerIssuerConfig{
				Backend: &kubernetesprovider.CertManagerIssuerConfig_Vault{
					Vault: &kubernetesprovider.CertManagerVaultConfig{
						Server: "https://vault.example.com:8200",
						Path:   "pki_int/sign/team-a",
						Auth: &kubernetesprovider.CertManagerVaultConfig_AppRoleAuth{
							AppRoleAuth: &kubernetesprovider.CertManagerVaultAppRoleAuth{
								Path:     "approle",
								RoleId:   "role-id",
								SecretId: "secret-id",
							},
						},
					},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.Succeed())
		})
	})

	ginkgo.Describe("required fields", func() {
		ginkgo.It("rejects a missing namespace", func() {
			input.Spec.Namespace = nil
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.Succeed())
		})

		ginkgo.It("rejects a missing config", func() {
			input.Spec.Config = nil
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.Succeed())
		})

		ginkgo.It("rejects a config with no backend selected", func() {
			input.Spec.Config = &kubernetesprovider.CertManagerIssuerConfig{}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.Succeed())
		})

		ginkgo.It("rejects a Vault AppRole auth missing the secret id", func() {
			input.Spec.Config = &kubernetesprovider.CertManagerIssuerConfig{
				Backend: &kubernetesprovider.CertManagerIssuerConfig_Vault{
					Vault: &kubernetesprovider.CertManagerVaultConfig{
						Server: "https://vault.example.com:8200",
						Path:   "pki_int/sign/team-a",
						Auth: &kubernetesprovider.CertManagerVaultConfig_AppRoleAuth{
							AppRoleAuth: &kubernetesprovider.CertManagerVaultAppRoleAuth{
								Path:   "approle",
								RoleId: "role-id",
							},
						},
					},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.Succeed())
		})
	})
})
