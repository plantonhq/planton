package kubernetesclusterissuerv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/apis/dev/planton/provider/kubernetes"
	"github.com/plantonhq/planton/apis/dev/planton/shared"
	"github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"
	foreignkeyv1 "github.com/plantonhq/planton/apis/dev/planton/shared/foreignkey/v1"
)

func TestKubernetesClusterIssuer(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "KubernetesClusterIssuer Suite")
}

var _ = ginkgo.Describe("KubernetesClusterIssuer Validation Tests", func() {
	var input *KubernetesClusterIssuer

	ginkgo.BeforeEach(func() {
		input = &KubernetesClusterIssuer{
			ApiVersion: "kubernetes.planton.dev/v1",
			Kind:       "KubernetesClusterIssuer",
			Metadata: &shared.CloudResourceMetadata{
				Name: "test-cluster-issuer",
			},
			Spec: &KubernetesClusterIssuerSpec{
				TargetCluster: &kubernetes.KubernetesClusterSelector{
					ClusterKind: cloudresourcekind.CloudResourceKind_GcpGkeCluster,
					ClusterName: "test-cluster",
				},
				CertManagerNamespace: &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
						Value: "cert-manager",
					},
				},
				DnsDomain: "example.com",
				Acme: &ClusterIssuerAcmeConfig{
					Email: "admin@example.com",
				},
				Provider: &KubernetesClusterIssuerSpec_Cloudflare{
					Cloudflare: &CloudflareDnsSolver{
						ApiToken: "test-cloudflare-token",
					},
				},
			},
		}
	})

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("with Cloudflare provider", func() {
			ginkgo.It("should not return a validation error", func() {
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with GCP Cloud DNS provider", func() {
			ginkgo.It("should not return a validation error", func() {
				input.Spec.Provider = &KubernetesClusterIssuerSpec_GcpCloudDns{
					GcpCloudDns: &GcpCloudDnsSolver{
						ProjectId: "my-gcp-project",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with AWS Route53 provider", func() {
			ginkgo.It("should not return a validation error", func() {
				input.Spec.Provider = &KubernetesClusterIssuerSpec_AwsRoute53{
					AwsRoute53: &AwsRoute53Solver{
						Region: "us-east-1",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with Azure DNS provider", func() {
			ginkgo.It("should not return a validation error", func() {
				input.Spec.Provider = &KubernetesClusterIssuerSpec_AzureDns{
					AzureDns: &AzureDnsSolver{
						SubscriptionId: "sub-12345",
						ResourceGroup:  "my-rg",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with custom ACME server", func() {
			ginkgo.It("should not return a validation error", func() {
				server := "https://acme-staging-v02.api.letsencrypt.org/directory"
				input.Spec.Acme.Server = &server
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with ACME server omitted (uses default)", func() {
			ginkgo.It("should not return a validation error", func() {
				input.Spec.Acme.Server = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("missing dns_domain", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.DnsDomain = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("missing acme config", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.Acme = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("missing acme email", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.Acme.Email = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("missing cert_manager_namespace", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.CertManagerNamespace = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("empty cert_manager_namespace value", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.CertManagerNamespace = &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
						Value: "",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("empty cloudflare api_token", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.Provider = &KubernetesClusterIssuerSpec_Cloudflare{
					Cloudflare: &CloudflareDnsSolver{
						ApiToken: "",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("GCP Cloud DNS missing project_id", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.Provider = &KubernetesClusterIssuerSpec_GcpCloudDns{
					GcpCloudDns: &GcpCloudDnsSolver{
						ProjectId: "",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("AWS Route53 missing region", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.Provider = &KubernetesClusterIssuerSpec_AwsRoute53{
					AwsRoute53: &AwsRoute53Solver{
						Region: "",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("Azure DNS missing subscription_id", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.Provider = &KubernetesClusterIssuerSpec_AzureDns{
					AzureDns: &AzureDnsSolver{
						SubscriptionId: "",
						ResourceGroup:  "my-rg",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("Azure DNS missing resource_group", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.Provider = &KubernetesClusterIssuerSpec_AzureDns{
					AzureDns: &AzureDnsSolver{
						SubscriptionId: "sub-12345",
						ResourceGroup:  "",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("missing metadata", func() {
			ginkgo.It("should return a validation error", func() {
				input.Metadata = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})
	})
})
