package kubernetesclusterissuerv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	kubernetesprovider "github.com/plantonhq/planton/catalog/kubernetes"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestKubernetesClusterIssuer(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "KubernetesClusterIssuer Suite")
}

func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

func cloudflareTokenSolver() *kubernetesprovider.CertManagerAcmeSolver {
	return &kubernetesprovider.CertManagerAcmeSolver{
		Challenge: &kubernetesprovider.CertManagerAcmeSolver_Dns01{
			Dns01: &kubernetesprovider.CertManagerAcmeDns01Solver{
				Provider: &kubernetesprovider.CertManagerAcmeDns01Solver_Cloudflare{
					Cloudflare: &kubernetesprovider.CertManagerDns01Cloudflare{
						Credential: &kubernetesprovider.CertManagerDns01Cloudflare_ApiToken{
							ApiToken: &kubernetesprovider.CertManagerCloudflareApiToken{Token: "test-token"},
						},
					},
				},
			},
		},
	}
}

var _ = ginkgo.Describe("KubernetesClusterIssuer Validation Tests", func() {
	var input *KubernetesClusterIssuer

	ginkgo.BeforeEach(func() {
		input = &KubernetesClusterIssuer{
			ApiVersion: "kubernetes.planton.dev/v1alpha1",
			Kind:       "KubernetesClusterIssuer",
			Metadata:   &shared.CatalogObjectMetadata{Name: "test-cluster-issuer"},
			Spec: &KubernetesClusterIssuerSpec{
				CertManagerNamespace: literal("cert-manager"),
				Config: &kubernetesprovider.CertManagerIssuerConfig{
					Backend: &kubernetesprovider.CertManagerIssuerConfig_Acme{
						Acme: &kubernetesprovider.CertManagerAcmeConfig{
							Email:   "platform@example.com",
							Solvers: []*kubernetesprovider.CertManagerAcmeSolver{cloudflareTokenSolver()},
						},
					},
				},
			},
		}
	})

	ginkgo.Describe("valid configurations", func() {
		ginkgo.It("accepts an ACME issuer with a Cloudflare token solver", func() {
			gomega.Expect(protovalidate.Validate(input)).To(gomega.Succeed())
		})

		ginkgo.It("accepts a self-signed backend", func() {
			input.Spec.Config = &kubernetesprovider.CertManagerIssuerConfig{
				Backend: &kubernetesprovider.CertManagerIssuerConfig_SelfSigned{
					SelfSigned: &kubernetesprovider.CertManagerSelfSignedConfig{},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.Succeed())
		})

		ginkgo.It("accepts a CA backend referencing a certificate output", func() {
			input.Spec.Config = &kubernetesprovider.CertManagerIssuerConfig{
				Backend: &kubernetesprovider.CertManagerIssuerConfig_Ca{
					Ca: &kubernetesprovider.CertManagerCaConfig{CaSecretName: literal("root-ca-tls")},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.Succeed())
		})

		ginkgo.It("accepts a Vault backend with kubernetes auth", func() {
			input.Spec.Config = &kubernetesprovider.CertManagerIssuerConfig{
				Backend: &kubernetesprovider.CertManagerIssuerConfig_Vault{
					Vault: &kubernetesprovider.CertManagerVaultConfig{
						Server: "https://vault.example.com:8200",
						Path:   "pki_int/sign/example",
						Auth: &kubernetesprovider.CertManagerVaultConfig_KubernetesAuth{
							KubernetesAuth: &kubernetesprovider.CertManagerVaultKubernetesAuth{
								Role:               "cert-manager",
								ServiceAccountName: literal("cert-manager-vault"),
							},
						},
					},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.Succeed())
		})

		ginkgo.It("accepts an HTTP-01 ingress solver", func() {
			input.Spec.Config.GetAcme().Solvers = []*kubernetesprovider.CertManagerAcmeSolver{{
				Challenge: &kubernetesprovider.CertManagerAcmeSolver_Http01{
					Http01: &kubernetesprovider.CertManagerAcmeHttp01Solver{
						Exposure: &kubernetesprovider.CertManagerAcmeHttp01Solver_Ingress{
							Ingress: &kubernetesprovider.CertManagerAcmeHttp01IngressSolver{IngressClassName: "nginx"},
						},
					},
				},
			}}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.Succeed())
		})

		ginkgo.It("accepts a Route53 solver with ambient (keyless) auth", func() {
			input.Spec.Config.GetAcme().Solvers = []*kubernetesprovider.CertManagerAcmeSolver{{
				Challenge: &kubernetesprovider.CertManagerAcmeSolver_Dns01{
					Dns01: &kubernetesprovider.CertManagerAcmeDns01Solver{
						Provider: &kubernetesprovider.CertManagerAcmeDns01Solver_Route53{
							Route53: &kubernetesprovider.CertManagerDns01Route53{Region: "us-east-1"},
						},
					},
				},
			}}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.Succeed())
		})
	})

	ginkgo.Describe("required fields", func() {
		ginkgo.It("rejects a missing cert_manager_namespace", func() {
			input.Spec.CertManagerNamespace = nil
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

		ginkgo.It("rejects an ACME backend without an email", func() {
			input.Spec.Config.GetAcme().Email = ""
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.Succeed())
		})

		ginkgo.It("rejects an ACME backend with zero solvers", func() {
			input.Spec.Config.GetAcme().Solvers = nil
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.Succeed())
		})
	})

	ginkgo.Describe("solver contracts", func() {
		ginkgo.It("rejects a solver with neither http01 nor dns01", func() {
			input.Spec.Config.GetAcme().Solvers = []*kubernetesprovider.CertManagerAcmeSolver{{}}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.Succeed())
		})

		ginkgo.It("rejects a dns01 solver with no provider", func() {
			input.Spec.Config.GetAcme().Solvers = []*kubernetesprovider.CertManagerAcmeSolver{{
				Challenge: &kubernetesprovider.CertManagerAcmeSolver_Dns01{
					Dns01: &kubernetesprovider.CertManagerAcmeDns01Solver{},
				},
			}}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.Succeed())
		})

		ginkgo.It("rejects an http01 solver with no exposure", func() {
			input.Spec.Config.GetAcme().Solvers = []*kubernetesprovider.CertManagerAcmeSolver{{
				Challenge: &kubernetesprovider.CertManagerAcmeSolver_Http01{
					Http01: &kubernetesprovider.CertManagerAcmeHttp01Solver{},
				},
			}}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.Succeed())
		})

		ginkgo.It("rejects an http01 ingress solver setting both class and name", func() {
			input.Spec.Config.GetAcme().Solvers = []*kubernetesprovider.CertManagerAcmeSolver{{
				Challenge: &kubernetesprovider.CertManagerAcmeSolver_Http01{
					Http01: &kubernetesprovider.CertManagerAcmeHttp01Solver{
						Exposure: &kubernetesprovider.CertManagerAcmeHttp01Solver_Ingress{
							Ingress: &kubernetesprovider.CertManagerAcmeHttp01IngressSolver{
								IngressClassName: "nginx",
								Name:             "existing-ingress",
							},
						},
					},
				},
			}}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.Succeed())
		})

		ginkgo.It("rejects an invalid cname_strategy", func() {
			solver := cloudflareTokenSolver()
			bad := "chase"
			solver.GetDns01().CnameStrategy = &bad
			input.Spec.Config.GetAcme().Solvers = []*kubernetesprovider.CertManagerAcmeSolver{solver}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.Succeed())
		})
	})

	ginkgo.Describe("credential contracts", func() {
		ginkgo.It("rejects an external account binding missing the hmac key", func() {
			input.Spec.Config.GetAcme().ExternalAccountBinding = &kubernetesprovider.CertManagerAcmeExternalAccountBinding{
				KeyId: "key-id",
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.Succeed())
		})

		ginkgo.It("rejects a route53 solver with both static credentials and a service account", func() {
			input.Spec.Config.GetAcme().Solvers = []*kubernetesprovider.CertManagerAcmeSolver{{
				Challenge: &kubernetesprovider.CertManagerAcmeSolver_Dns01{
					Dns01: &kubernetesprovider.CertManagerAcmeDns01Solver{
						Provider: &kubernetesprovider.CertManagerAcmeDns01Solver_Route53{
							Route53: &kubernetesprovider.CertManagerDns01Route53{
								Region: "us-east-1",
								StaticCredentials: &kubernetesprovider.CertManagerRoute53StaticCredentials{
									AccessKeyId:     "AKIAEXAMPLE",
									SecretAccessKey: "secret",
								},
								ServiceAccount: &kubernetesprovider.CertManagerRoute53ServiceAccountAuth{
									ServiceAccountName: literal("dns01-sa"),
								},
							},
						},
					},
				},
			}}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.Succeed())
		})

		ginkgo.It("rejects an azure_dns service principal missing the tenant id", func() {
			input.Spec.Config.GetAcme().Solvers = []*kubernetesprovider.CertManagerAcmeSolver{{
				Challenge: &kubernetesprovider.CertManagerAcmeSolver_Dns01{
					Dns01: &kubernetesprovider.CertManagerAcmeDns01Solver{
						Provider: &kubernetesprovider.CertManagerAcmeDns01Solver_AzureDns{
							AzureDns: &kubernetesprovider.CertManagerDns01AzureDns{
								SubscriptionId:    "00000000-0000-0000-0000-000000000000",
								ResourceGroupName: "dns-rg",
								ClientId:          "11111111-1111-1111-1111-111111111111",
								ClientSecret:      "secret",
							},
						},
					},
				},
			}}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.Succeed())
		})

		ginkgo.It("rejects an azure_dns managed identity selecting both client and resource ids", func() {
			input.Spec.Config.GetAcme().Solvers = []*kubernetesprovider.CertManagerAcmeSolver{{
				Challenge: &kubernetesprovider.CertManagerAcmeSolver_Dns01{
					Dns01: &kubernetesprovider.CertManagerAcmeDns01Solver{
						Provider: &kubernetesprovider.CertManagerAcmeDns01Solver_AzureDns{
							AzureDns: &kubernetesprovider.CertManagerDns01AzureDns{
								SubscriptionId:    "00000000-0000-0000-0000-000000000000",
								ResourceGroupName: "dns-rg",
								ManagedIdentity: &kubernetesprovider.CertManagerAzureManagedIdentity{
									ClientId:   "11111111-1111-1111-1111-111111111111",
									ResourceId: "/subscriptions/x/resourceGroups/y/providers/Microsoft.ManagedIdentity/userAssignedIdentities/z",
								},
							},
						},
					},
				},
			}}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.Succeed())
		})

		ginkgo.It("rejects an rfc2136 solver with a TSIG key name but no secret", func() {
			input.Spec.Config.GetAcme().Solvers = []*kubernetesprovider.CertManagerAcmeSolver{{
				Challenge: &kubernetesprovider.CertManagerAcmeSolver_Dns01{
					Dns01: &kubernetesprovider.CertManagerAcmeDns01Solver{
						Provider: &kubernetesprovider.CertManagerAcmeDns01Solver_Rfc2136{
							Rfc2136: &kubernetesprovider.CertManagerDns01Rfc2136{
								Nameserver:  "10.0.0.53:53",
								TsigKeyName: "acme-key",
							},
						},
					},
				},
			}}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.Succeed())
		})

		ginkgo.It("rejects a vault backend with no auth method", func() {
			input.Spec.Config = &kubernetesprovider.CertManagerIssuerConfig{
				Backend: &kubernetesprovider.CertManagerIssuerConfig_Vault{
					Vault: &kubernetesprovider.CertManagerVaultConfig{
						Server: "https://vault.example.com:8200",
						Path:   "pki_int/sign/example",
					},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.Succeed())
		})
	})
})
