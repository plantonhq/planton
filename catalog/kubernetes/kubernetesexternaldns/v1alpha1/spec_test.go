package kubernetesexternaldnsv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	kubernetesprovider "github.com/plantonhq/planton/catalog/kubernetes"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestKubernetesExternalDns(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "KubernetesExternalDns Suite")
}

func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

func strPtr(v string) *string { return &v }

var _ = ginkgo.Describe("KubernetesExternalDns Validation Tests", func() {
	var input *KubernetesExternalDns

	ginkgo.BeforeEach(func() {
		input = &KubernetesExternalDns{
			ApiVersion: "kubernetes.planton.dev/v1alpha1",
			Kind:       "KubernetesExternalDns",
			Metadata:   &shared.CatalogObjectMetadata{Name: "external-dns"},
			Spec: &KubernetesExternalDnsSpec{
				Namespace:       literal("external-dns"),
				CreateNamespace: true,
				DnsProvider: &KubernetesExternalDnsSpec_AwsRoute53{
					AwsRoute53: &KubernetesExternalDnsAwsRoute53{Region: "us-east-1"},
				},
			},
		}
	})

	ginkgo.Describe("valid configurations", func() {
		ginkgo.It("accepts a keyless AWS Route 53 install", func() {
			input.Spec.WorkloadIdentity = &kubernetesprovider.KubernetesWorkloadIdentity{
				Provider: &kubernetesprovider.KubernetesWorkloadIdentity_Eks{
					Eks: &kubernetesprovider.KubernetesWorkloadIdentityEksIrsa{
						RoleArn: literal("arn:aws:iam::123456789012:role/external-dns"),
					},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.Succeed())
		})

		ginkgo.It("accepts the cross-cloud Cloudflare arm with sync tuning", func() {
			perPage := uint32(5000)
			input.Spec.DnsProvider = &KubernetesExternalDnsSpec_Cloudflare{
				Cloudflare: &KubernetesExternalDnsCloudflare{
					ApiToken:          "cf-token",
					ZoneIdFilters:     []*foreignkeyv1.StringValueOrRef{literal("023e105f4ecef8ad9ca31a8372d0c353")},
					Proxied:           true,
					DnsRecordsPerPage: &perPage,
				},
			}
			input.Spec.Policy = strPtr("sync")
			input.Spec.TxtOwnerId = "prod-cluster"
			input.Spec.Sources = []string{"service", "ingress", "gateway-httproute"}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.Succeed())
		})

		ginkgo.It("accepts listener sets beside a Gateway API route source", func() {
			input.Spec.Sources = []string{"gateway-httproute", "gateway-tlsroute"}
			input.Spec.GatewayListenerSets = true
			gomega.Expect(protovalidate.Validate(input)).To(gomega.Succeed())
		})

		ginkgo.It("accepts Google Cloud DNS with a static key", func() {
			input.Spec.DnsProvider = &KubernetesExternalDnsSpec_GoogleCloudDns{
				GoogleCloudDns: &KubernetesExternalDnsGoogleCloudDns{
					Project:               literal("my-project"),
					ZoneVisibility:        strPtr("public"),
					ServiceAccountKeyJson: `{"type":"service_account"}`,
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.Succeed())
		})

		ginkgo.It("accepts Azure private zones with workload identity", func() {
			input.Spec.DnsProvider = &KubernetesExternalDnsSpec_AzureDns{
				AzureDns: &KubernetesExternalDnsAzureDns{
					ResourceGroup:  "dns-rg",
					SubscriptionId: "00000000-0000-0000-0000-000000000000",
					PrivateZones:   true,
				},
			}
			input.Spec.WorkloadIdentity = &kubernetesprovider.KubernetesWorkloadIdentity{
				Provider: &kubernetesprovider.KubernetesWorkloadIdentity_Aks{
					Aks: &kubernetesprovider.KubernetesWorkloadIdentityAks{
						ClientId: literal("11111111-1111-1111-1111-111111111111"),
					},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.Succeed())
		})

		ginkgo.It("accepts the webhook extension arm", func() {
			input.Spec.DnsProvider = &KubernetesExternalDnsSpec_Webhook{
				Webhook: &KubernetesExternalDnsWebhook{
					ImageRepository: "ghcr.io/example/edns-webhook",
					ImageTag:        "v1.0.0",
					Args:            []string{"--zone=example.org"},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.Succeed())
		})

		ginkgo.It("accepts the in-memory sandbox arm with TXT registry knobs", func() {
			input.Spec.DnsProvider = &KubernetesExternalDnsSpec_InMemory{
				InMemory: &KubernetesExternalDnsInMemory{Zones: []string{"example.org"}},
			}
			input.Spec.TxtOwnerId = "sandbox"
			input.Spec.TxtPrefix = "edns-"
			gomega.Expect(protovalidate.Validate(input)).To(gomega.Succeed())
		})

		ginkgo.It("accepts the dynamodb registry with its table settings", func() {
			input.Spec.Registry = strPtr("dynamodb")
			input.Spec.DynamodbTable = "external-dns"
			input.Spec.DynamodbRegion = "us-east-1"
			gomega.Expect(protovalidate.Validate(input)).To(gomega.Succeed())
		})
	})

	ginkgo.Describe("required fields and contracts", func() {
		ginkgo.It("rejects a missing namespace", func() {
			input.Spec.Namespace = nil
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.Succeed())
		})

		ginkgo.It("rejects a spec with no DNS provider", func() {
			input.Spec.DnsProvider = nil
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.Succeed())
		})

		ginkgo.It("rejects txt_prefix and txt_suffix together", func() {
			input.Spec.TxtPrefix = "a-"
			input.Spec.TxtSuffix = "-b"
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.Succeed())
		})

		ginkgo.It("rejects listener sets without a Gateway API route source", func() {
			input.Spec.Sources = []string{"service", "ingress"}
			input.Spec.GatewayListenerSets = true
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.HaveOccurred())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("needs a Gateway API route source"))
		})

		ginkgo.It("rejects dynamodb table settings without the dynamodb registry", func() {
			input.Spec.DynamodbTable = "external-dns"
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.Succeed())
		})

		ginkgo.It("rejects an unpaired AWS static credential", func() {
			input.Spec.DnsProvider = &KubernetesExternalDnsSpec_AwsRoute53{
				AwsRoute53: &KubernetesExternalDnsAwsRoute53{
					Region:      "us-east-1",
					AccessKeyId: "AKIAEXAMPLE",
				},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.Succeed())
		})

		ginkgo.It("rejects an unpaired Azure service-principal credential", func() {
			input.Spec.DnsProvider = &KubernetesExternalDnsSpec_AzureDns{
				AzureDns: &KubernetesExternalDnsAzureDns{
					ResourceGroup:  "dns-rg",
					SubscriptionId: "00000000-0000-0000-0000-000000000000",
					ClientId:       "22222222-2222-2222-2222-222222222222",
				},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.Succeed())
		})

		ginkgo.It("rejects an Azure service principal without a tenant", func() {
			input.Spec.DnsProvider = &KubernetesExternalDnsSpec_AzureDns{
				AzureDns: &KubernetesExternalDnsAzureDns{
					ResourceGroup:  "dns-rg",
					SubscriptionId: "00000000-0000-0000-0000-000000000000",
					ClientId:       "22222222-2222-2222-2222-222222222222",
					ClientSecret:   "sp-secret",
				},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.Succeed())
		})

		ginkgo.It("rejects an empty Cloudflare token", func() {
			input.Spec.DnsProvider = &KubernetesExternalDnsSpec_Cloudflare{
				Cloudflare: &KubernetesExternalDnsCloudflare{},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.Succeed())
		})

		ginkgo.It("rejects a Cloudflare page size beyond the API maximum", func() {
			perPage := uint32(5001)
			input.Spec.DnsProvider = &KubernetesExternalDnsSpec_Cloudflare{
				Cloudflare: &KubernetesExternalDnsCloudflare{
					ApiToken:          "cf-token",
					DnsRecordsPerPage: &perPage,
				},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.Succeed())
		})

		ginkgo.It("rejects a webhook arm without an image", func() {
			input.Spec.DnsProvider = &KubernetesExternalDnsSpec_Webhook{
				Webhook: &KubernetesExternalDnsWebhook{},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.Succeed())
		})

		ginkgo.It("rejects an unknown source", func() {
			input.Spec.Sources = []string{"service", "deployment"}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.Succeed())
		})

		ginkgo.It("rejects an unknown sync policy", func() {
			input.Spec.Policy = strPtr("delete-all")
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.Succeed())
		})

		ginkgo.It("rejects an unknown registry", func() {
			input.Spec.Registry = strPtr("etcd")
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.Succeed())
		})

		ginkgo.It("rejects a malformed reconciliation interval", func() {
			input.Spec.Interval = strPtr("every minute")
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.Succeed())
		})

		ginkgo.It("rejects an unknown zone type on AWS", func() {
			input.Spec.DnsProvider = &KubernetesExternalDnsSpec_AwsRoute53{
				AwsRoute53: &KubernetesExternalDnsAwsRoute53{
					Region:   "us-east-1",
					ZoneType: strPtr("internal"),
				},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.Succeed())
		})
	})
})
