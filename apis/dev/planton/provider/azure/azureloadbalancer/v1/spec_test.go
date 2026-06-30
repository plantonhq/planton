package azureloadbalancerv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/apis/dev/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/apis/dev/planton/shared/foreignkey/v1"
)

func TestAzureLoadBalancerSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureLoadBalancerSpec Validation Tests")
}

var _ = ginkgo.Describe("AzureLoadBalancerSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("azure_load_balancer", func() {

			ginkgo.It("should not return a validation error for a minimal public LB", func() {
				input := &AzureLoadBalancer{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "AzureLoadBalancer",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-lb",
					},
					Spec: &AzureLoadBalancerSpec{
						Region: "eastus",
						ResourceGroup: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "my-rg",
							},
						},
						Name: "test-lb",
						PublicIpId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "/subscriptions/sub/resourceGroups/rg/providers/Microsoft.Network/publicIPAddresses/pip",
							},
						},
						BackendPools: []*AzureBackendPool{
							{Name: "default"},
						},
						HealthProbes: []*AzureHealthProbe{
							{
								Name:     "tcp-probe",
								Protocol: "Tcp",
								Port:     80,
							},
						},
						Rules: []*AzureLoadBalancingRule{
							{
								Name:            "http-rule",
								Protocol:        "Tcp",
								FrontendPort:    80,
								BackendPort:     80,
								BackendPoolName: "default",
								ProbeName:       "tcp-probe",
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for a minimal internal LB", func() {
				input := &AzureLoadBalancer{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "AzureLoadBalancer",
					Metadata: &shared.CloudResourceMetadata{
						Name: "internal-lb",
					},
					Spec: &AzureLoadBalancerSpec{
						Region: "westeurope",
						ResourceGroup: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "prod-rg",
							},
						},
						Name: "internal-lb",
						SubnetId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "/subscriptions/sub/resourceGroups/rg/providers/Microsoft.Network/virtualNetworks/vnet/subnets/default",
							},
						},
						BackendPools: []*AzureBackendPool{
							{Name: "default"},
						},
						HealthProbes: []*AzureHealthProbe{
							{
								Name:     "tcp-probe",
								Protocol: "Tcp",
								Port:     8080,
							},
						},
						Rules: []*AzureLoadBalancingRule{
							{
								Name:            "app-rule",
								Protocol:        "Tcp",
								FrontendPort:    8080,
								BackendPort:     8080,
								BackendPoolName: "default",
								ProbeName:       "tcp-probe",
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for internal LB with static private IP", func() {
				input := &AzureLoadBalancer{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "AzureLoadBalancer",
					Metadata: &shared.CloudResourceMetadata{
						Name: "static-lb",
					},
					Spec: &AzureLoadBalancerSpec{
						Region: "eastus",
						ResourceGroup: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "my-rg",
							},
						},
						Name: "static-lb",
						SubnetId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "/subscriptions/sub/resourceGroups/rg/providers/Microsoft.Network/virtualNetworks/vnet/subnets/default",
							},
						},
						PrivateIpAddress: "10.0.1.100",
						BackendPools: []*AzureBackendPool{
							{Name: "default"},
						},
						HealthProbes: []*AzureHealthProbe{
							{
								Name:        "http-probe",
								Protocol:    "Http",
								Port:        80,
								RequestPath: "/health",
							},
						},
						Rules: []*AzureLoadBalancingRule{
							{
								Name:            "http-rule",
								Protocol:        "Tcp",
								FrontendPort:    80,
								BackendPort:     80,
								BackendPoolName: "default",
								ProbeName:       "http-probe",
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with multiple backend pools and rules", func() {
				input := &AzureLoadBalancer{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "AzureLoadBalancer",
					Metadata: &shared.CloudResourceMetadata{
						Name: "multi-pool-lb",
						Org:  "mycompany",
						Env:  "production",
					},
					Spec: &AzureLoadBalancerSpec{
						Region: "eastus",
						ResourceGroup: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "prod-rg",
							},
						},
						Name: "multi-pool-lb",
						PublicIpId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "/subscriptions/sub/resourceGroups/rg/providers/Microsoft.Network/publicIPAddresses/pip",
							},
						},
						BackendPools: []*AzureBackendPool{
							{Name: "web-pool"},
							{Name: "api-pool"},
						},
						HealthProbes: []*AzureHealthProbe{
							{
								Name:        "http-probe",
								Protocol:    "Http",
								Port:        80,
								RequestPath: "/health",
							},
							{
								Name:     "tcp-8080-probe",
								Protocol: "Tcp",
								Port:     8080,
							},
						},
						Rules: []*AzureLoadBalancingRule{
							{
								Name:            "web-http",
								Protocol:        "Tcp",
								FrontendPort:    80,
								BackendPort:     80,
								BackendPoolName: "web-pool",
								ProbeName:       "http-probe",
							},
							{
								Name:            "web-https",
								Protocol:        "Tcp",
								FrontendPort:    443,
								BackendPort:     443,
								BackendPoolName: "web-pool",
								ProbeName:       "http-probe",
							},
							{
								Name:            "api",
								Protocol:        "Tcp",
								FrontendPort:    8080,
								BackendPort:     8080,
								BackendPoolName: "api-pool",
								ProbeName:       "tcp-8080-probe",
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with Https probe", func() {
				input := &AzureLoadBalancer{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "AzureLoadBalancer",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-lb",
					},
					Spec: &AzureLoadBalancerSpec{
						Region: "eastus",
						ResourceGroup: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "my-rg",
							},
						},
						Name: "test-lb",
						PublicIpId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "/subscriptions/sub/resourceGroups/rg/providers/Microsoft.Network/publicIPAddresses/pip",
							},
						},
						BackendPools: []*AzureBackendPool{
							{Name: "default"},
						},
						HealthProbes: []*AzureHealthProbe{
							{
								Name:        "https-probe",
								Protocol:    "Https",
								Port:        443,
								RequestPath: "/healthz",
							},
						},
						Rules: []*AzureLoadBalancingRule{
							{
								Name:            "https-rule",
								Protocol:        "Tcp",
								FrontendPort:    443,
								BackendPort:     443,
								BackendPoolName: "default",
								ProbeName:       "https-probe",
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with HA ports rule (protocol All)", func() {
				input := &AzureLoadBalancer{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "AzureLoadBalancer",
					Metadata: &shared.CloudResourceMetadata{
						Name: "ha-lb",
					},
					Spec: &AzureLoadBalancerSpec{
						Region: "eastus",
						ResourceGroup: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "my-rg",
							},
						},
						Name: "ha-lb",
						SubnetId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "/subscriptions/sub/resourceGroups/rg/providers/Microsoft.Network/virtualNetworks/vnet/subnets/default",
							},
						},
						BackendPools: []*AzureBackendPool{
							{Name: "default"},
						},
						HealthProbes: []*AzureHealthProbe{
							{
								Name:     "tcp-probe",
								Protocol: "Tcp",
								Port:     443,
							},
						},
						Rules: []*AzureLoadBalancingRule{
							{
								Name:            "ha-ports",
								Protocol:        "All",
								FrontendPort:    0,
								BackendPort:     0,
								BackendPoolName: "default",
								ProbeName:       "tcp-probe",
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with Udp rule protocol", func() {
				input := &AzureLoadBalancer{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "AzureLoadBalancer",
					Metadata: &shared.CloudResourceMetadata{
						Name: "udp-lb",
					},
					Spec: &AzureLoadBalancerSpec{
						Region: "eastus",
						ResourceGroup: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "my-rg",
							},
						},
						Name: "udp-lb",
						PublicIpId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "/subscriptions/sub/resourceGroups/rg/providers/Microsoft.Network/publicIPAddresses/pip",
							},
						},
						BackendPools: []*AzureBackendPool{
							{Name: "default"},
						},
						HealthProbes: []*AzureHealthProbe{
							{
								Name:     "tcp-probe",
								Protocol: "Tcp",
								Port:     53,
							},
						},
						Rules: []*AzureLoadBalancingRule{
							{
								Name:            "dns-rule",
								Protocol:        "Udp",
								FrontendPort:    53,
								BackendPort:     53,
								BackendPoolName: "default",
								ProbeName:       "tcp-probe",
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with floating IP and SNAT disabled", func() {
				floatingIp := true
				disableSnat := true
				input := &AzureLoadBalancer{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "AzureLoadBalancer",
					Metadata: &shared.CloudResourceMetadata{
						Name: "advanced-lb",
					},
					Spec: &AzureLoadBalancerSpec{
						Region: "eastus",
						ResourceGroup: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "my-rg",
							},
						},
						Name: "advanced-lb",
						PublicIpId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "/subscriptions/sub/resourceGroups/rg/providers/Microsoft.Network/publicIPAddresses/pip",
							},
						},
						BackendPools: []*AzureBackendPool{
							{Name: "default"},
						},
						HealthProbes: []*AzureHealthProbe{
							{
								Name:     "tcp-probe",
								Protocol: "Tcp",
								Port:     1433,
							},
						},
						Rules: []*AzureLoadBalancingRule{
							{
								Name:                "sql-always-on",
								Protocol:            "Tcp",
								FrontendPort:        1433,
								BackendPort:         1433,
								BackendPoolName:     "default",
								ProbeName:           "tcp-probe",
								EnableFloatingIp:    &floatingIp,
								DisableOutboundSnat: &disableSnat,
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("azure_load_balancer", func() {

			ginkgo.It("should return a validation error when region is missing", func() {
				input := &AzureLoadBalancer{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "AzureLoadBalancer",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-lb",
					},
					Spec: &AzureLoadBalancerSpec{
						ResourceGroup: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "my-rg",
							},
						},
						Name: "test-lb",
						PublicIpId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "pip-id",
							},
						},
						BackendPools: []*AzureBackendPool{{Name: "default"}},
						HealthProbes: []*AzureHealthProbe{{Name: "probe", Protocol: "Tcp", Port: 80}},
						Rules:        []*AzureLoadBalancingRule{{Name: "rule", Protocol: "Tcp", FrontendPort: 80, BackendPort: 80, BackendPoolName: "default", ProbeName: "probe"}},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when resource_group is missing", func() {
				input := &AzureLoadBalancer{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "AzureLoadBalancer",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-lb",
					},
					Spec: &AzureLoadBalancerSpec{
						Region: "eastus",
						Name:   "test-lb",
						PublicIpId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "pip-id",
							},
						},
						BackendPools: []*AzureBackendPool{{Name: "default"}},
						HealthProbes: []*AzureHealthProbe{{Name: "probe", Protocol: "Tcp", Port: 80}},
						Rules:        []*AzureLoadBalancingRule{{Name: "rule", Protocol: "Tcp", FrontendPort: 80, BackendPort: 80, BackendPoolName: "default", ProbeName: "probe"}},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when name is missing", func() {
				input := &AzureLoadBalancer{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "AzureLoadBalancer",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-lb",
					},
					Spec: &AzureLoadBalancerSpec{
						Region: "eastus",
						ResourceGroup: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "my-rg",
							},
						},
						PublicIpId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "pip-id",
							},
						},
						BackendPools: []*AzureBackendPool{{Name: "default"}},
						HealthProbes: []*AzureHealthProbe{{Name: "probe", Protocol: "Tcp", Port: 80}},
						Rules:        []*AzureLoadBalancingRule{{Name: "rule", Protocol: "Tcp", FrontendPort: 80, BackendPort: 80, BackendPoolName: "default", ProbeName: "probe"}},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when name exceeds 80 characters", func() {
				tooLongName := ""
				for len(tooLongName) < 81 {
					tooLongName += "a"
				}
				input := &AzureLoadBalancer{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "AzureLoadBalancer",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-lb",
					},
					Spec: &AzureLoadBalancerSpec{
						Region: "eastus",
						ResourceGroup: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "my-rg",
							},
						},
						Name: tooLongName,
						PublicIpId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "pip-id",
							},
						},
						BackendPools: []*AzureBackendPool{{Name: "default"}},
						HealthProbes: []*AzureHealthProbe{{Name: "probe", Protocol: "Tcp", Port: 80}},
						Rules:        []*AzureLoadBalancingRule{{Name: "rule", Protocol: "Tcp", FrontendPort: 80, BackendPort: 80, BackendPoolName: "default", ProbeName: "probe"}},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when metadata is missing", func() {
				input := &AzureLoadBalancer{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "AzureLoadBalancer",
					Spec: &AzureLoadBalancerSpec{
						Region: "eastus",
						ResourceGroup: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "my-rg",
							},
						},
						Name: "test-lb",
						PublicIpId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "pip-id",
							},
						},
						BackendPools: []*AzureBackendPool{{Name: "default"}},
						HealthProbes: []*AzureHealthProbe{{Name: "probe", Protocol: "Tcp", Port: 80}},
						Rules:        []*AzureLoadBalancingRule{{Name: "rule", Protocol: "Tcp", FrontendPort: 80, BackendPort: 80, BackendPoolName: "default", ProbeName: "probe"}},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when spec is missing", func() {
				input := &AzureLoadBalancer{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "AzureLoadBalancer",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-lb",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when api_version is incorrect", func() {
				input := &AzureLoadBalancer{
					ApiVersion: "wrong.version/v1",
					Kind:       "AzureLoadBalancer",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-lb",
					},
					Spec: &AzureLoadBalancerSpec{
						Region: "eastus",
						ResourceGroup: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "my-rg",
							},
						},
						Name: "test-lb",
						PublicIpId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "pip-id",
							},
						},
						BackendPools: []*AzureBackendPool{{Name: "default"}},
						HealthProbes: []*AzureHealthProbe{{Name: "probe", Protocol: "Tcp", Port: 80}},
						Rules:        []*AzureLoadBalancingRule{{Name: "rule", Protocol: "Tcp", FrontendPort: 80, BackendPort: 80, BackendPoolName: "default", ProbeName: "probe"}},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when kind is incorrect", func() {
				input := &AzureLoadBalancer{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "WrongKind",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-lb",
					},
					Spec: &AzureLoadBalancerSpec{
						Region: "eastus",
						ResourceGroup: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "my-rg",
							},
						},
						Name: "test-lb",
						PublicIpId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "pip-id",
							},
						},
						BackendPools: []*AzureBackendPool{{Name: "default"}},
						HealthProbes: []*AzureHealthProbe{{Name: "probe", Protocol: "Tcp", Port: 80}},
						Rules:        []*AzureLoadBalancingRule{{Name: "rule", Protocol: "Tcp", FrontendPort: 80, BackendPort: 80, BackendPoolName: "default", ProbeName: "probe"}},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when backend_pools is empty", func() {
				input := &AzureLoadBalancer{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "AzureLoadBalancer",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-lb",
					},
					Spec: &AzureLoadBalancerSpec{
						Region: "eastus",
						ResourceGroup: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "my-rg",
							},
						},
						Name: "test-lb",
						PublicIpId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "pip-id",
							},
						},
						BackendPools: []*AzureBackendPool{},
						HealthProbes: []*AzureHealthProbe{{Name: "probe", Protocol: "Tcp", Port: 80}},
						Rules:        []*AzureLoadBalancingRule{{Name: "rule", Protocol: "Tcp", FrontendPort: 80, BackendPort: 80, BackendPoolName: "default", ProbeName: "probe"}},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when health_probes is empty", func() {
				input := &AzureLoadBalancer{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "AzureLoadBalancer",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-lb",
					},
					Spec: &AzureLoadBalancerSpec{
						Region: "eastus",
						ResourceGroup: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "my-rg",
							},
						},
						Name: "test-lb",
						PublicIpId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "pip-id",
							},
						},
						BackendPools: []*AzureBackendPool{{Name: "default"}},
						HealthProbes: []*AzureHealthProbe{},
						Rules:        []*AzureLoadBalancingRule{{Name: "rule", Protocol: "Tcp", FrontendPort: 80, BackendPort: 80, BackendPoolName: "default", ProbeName: "probe"}},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when rules is empty", func() {
				input := &AzureLoadBalancer{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "AzureLoadBalancer",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-lb",
					},
					Spec: &AzureLoadBalancerSpec{
						Region: "eastus",
						ResourceGroup: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "my-rg",
							},
						},
						Name: "test-lb",
						PublicIpId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "pip-id",
							},
						},
						BackendPools: []*AzureBackendPool{{Name: "default"}},
						HealthProbes: []*AzureHealthProbe{{Name: "probe", Protocol: "Tcp", Port: 80}},
						Rules:        []*AzureLoadBalancingRule{},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when probe protocol is invalid", func() {
				input := &AzureLoadBalancer{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "AzureLoadBalancer",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-lb",
					},
					Spec: &AzureLoadBalancerSpec{
						Region: "eastus",
						ResourceGroup: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "my-rg",
							},
						},
						Name: "test-lb",
						PublicIpId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "pip-id",
							},
						},
						BackendPools: []*AzureBackendPool{{Name: "default"}},
						HealthProbes: []*AzureHealthProbe{
							{Name: "bad-probe", Protocol: "UDP", Port: 80},
						},
						Rules: []*AzureLoadBalancingRule{{Name: "rule", Protocol: "Tcp", FrontendPort: 80, BackendPort: 80, BackendPoolName: "default", ProbeName: "bad-probe"}},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when rule protocol is invalid", func() {
				input := &AzureLoadBalancer{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "AzureLoadBalancer",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-lb",
					},
					Spec: &AzureLoadBalancerSpec{
						Region: "eastus",
						ResourceGroup: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "my-rg",
							},
						},
						Name: "test-lb",
						PublicIpId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "pip-id",
							},
						},
						BackendPools: []*AzureBackendPool{{Name: "default"}},
						HealthProbes: []*AzureHealthProbe{{Name: "probe", Protocol: "Tcp", Port: 80}},
						Rules: []*AzureLoadBalancingRule{
							{Name: "bad-rule", Protocol: "HTTP", FrontendPort: 80, BackendPort: 80, BackendPoolName: "default", ProbeName: "probe"},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when probe port is out of range", func() {
				input := &AzureLoadBalancer{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "AzureLoadBalancer",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-lb",
					},
					Spec: &AzureLoadBalancerSpec{
						Region: "eastus",
						ResourceGroup: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "my-rg",
							},
						},
						Name: "test-lb",
						PublicIpId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "pip-id",
							},
						},
						BackendPools: []*AzureBackendPool{{Name: "default"}},
						HealthProbes: []*AzureHealthProbe{
							{Name: "bad-port", Protocol: "Tcp", Port: 70000},
						},
						Rules: []*AzureLoadBalancingRule{{Name: "rule", Protocol: "Tcp", FrontendPort: 80, BackendPort: 80, BackendPoolName: "default", ProbeName: "bad-port"}},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when probe interval is below minimum", func() {
				interval := int32(3)
				input := &AzureLoadBalancer{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "AzureLoadBalancer",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-lb",
					},
					Spec: &AzureLoadBalancerSpec{
						Region: "eastus",
						ResourceGroup: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "my-rg",
							},
						},
						Name: "test-lb",
						PublicIpId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "pip-id",
							},
						},
						BackendPools: []*AzureBackendPool{{Name: "default"}},
						HealthProbes: []*AzureHealthProbe{
							{Name: "fast-probe", Protocol: "Tcp", Port: 80, IntervalInSeconds: &interval},
						},
						Rules: []*AzureLoadBalancingRule{{Name: "rule", Protocol: "Tcp", FrontendPort: 80, BackendPort: 80, BackendPoolName: "default", ProbeName: "fast-probe"}},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when rule idle_timeout_in_minutes is below minimum", func() {
				timeout := int32(2)
				input := &AzureLoadBalancer{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "AzureLoadBalancer",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-lb",
					},
					Spec: &AzureLoadBalancerSpec{
						Region: "eastus",
						ResourceGroup: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "my-rg",
							},
						},
						Name: "test-lb",
						PublicIpId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "pip-id",
							},
						},
						BackendPools: []*AzureBackendPool{{Name: "default"}},
						HealthProbes: []*AzureHealthProbe{{Name: "probe", Protocol: "Tcp", Port: 80}},
						Rules: []*AzureLoadBalancingRule{
							{Name: "rule", Protocol: "Tcp", FrontendPort: 80, BackendPort: 80, BackendPoolName: "default", ProbeName: "probe", IdleTimeoutInMinutes: &timeout},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when rule idle_timeout_in_minutes exceeds maximum", func() {
				timeout := int32(101)
				input := &AzureLoadBalancer{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "AzureLoadBalancer",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-lb",
					},
					Spec: &AzureLoadBalancerSpec{
						Region: "eastus",
						ResourceGroup: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "my-rg",
							},
						},
						Name: "test-lb",
						PublicIpId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "pip-id",
							},
						},
						BackendPools: []*AzureBackendPool{{Name: "default"}},
						HealthProbes: []*AzureHealthProbe{{Name: "probe", Protocol: "Tcp", Port: 80}},
						Rules: []*AzureLoadBalancingRule{
							{Name: "rule", Protocol: "Tcp", FrontendPort: 80, BackendPort: 80, BackendPoolName: "default", ProbeName: "probe", IdleTimeoutInMinutes: &timeout},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when backend pool name is missing", func() {
				input := &AzureLoadBalancer{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "AzureLoadBalancer",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-lb",
					},
					Spec: &AzureLoadBalancerSpec{
						Region: "eastus",
						ResourceGroup: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "my-rg",
							},
						},
						Name: "test-lb",
						PublicIpId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "pip-id",
							},
						},
						BackendPools: []*AzureBackendPool{{Name: ""}},
						HealthProbes: []*AzureHealthProbe{{Name: "probe", Protocol: "Tcp", Port: 80}},
						Rules:        []*AzureLoadBalancingRule{{Name: "rule", Protocol: "Tcp", FrontendPort: 80, BackendPort: 80, BackendPoolName: "default", ProbeName: "probe"}},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when rule backend_pool_name is missing", func() {
				input := &AzureLoadBalancer{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "AzureLoadBalancer",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-lb",
					},
					Spec: &AzureLoadBalancerSpec{
						Region: "eastus",
						ResourceGroup: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "my-rg",
							},
						},
						Name: "test-lb",
						PublicIpId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "pip-id",
							},
						},
						BackendPools: []*AzureBackendPool{{Name: "default"}},
						HealthProbes: []*AzureHealthProbe{{Name: "probe", Protocol: "Tcp", Port: 80}},
						Rules: []*AzureLoadBalancingRule{
							{Name: "rule", Protocol: "Tcp", FrontendPort: 80, BackendPort: 80, ProbeName: "probe"},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when rule probe_name is missing", func() {
				input := &AzureLoadBalancer{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "AzureLoadBalancer",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-lb",
					},
					Spec: &AzureLoadBalancerSpec{
						Region: "eastus",
						ResourceGroup: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "my-rg",
							},
						},
						Name: "test-lb",
						PublicIpId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "pip-id",
							},
						},
						BackendPools: []*AzureBackendPool{{Name: "default"}},
						HealthProbes: []*AzureHealthProbe{{Name: "probe", Protocol: "Tcp", Port: 80}},
						Rules: []*AzureLoadBalancingRule{
							{Name: "rule", Protocol: "Tcp", FrontendPort: 80, BackendPort: 80, BackendPoolName: "default"},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})
	})
})
