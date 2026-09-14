package azuredatafactoryintegrationruntimev1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestAzureDataFactoryIntegrationRuntimeSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureDataFactoryIntegrationRuntimeSpec Validation Tests")
}

// literal builds a StringValueOrRef carrying a literal value.
func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

const (
	testFactoryId = "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/app-rg/providers/Microsoft.DataFactory/factories/app-df"
	testRuntimeId = testFactoryId + "/integrationRuntimes/onprem-bridge"
	testSubnetId  = "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/net-rg/providers/Microsoft.Network/virtualNetworks/app-vnet/subnets/ssis-subnet"
	testVnetId    = "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/net-rg/providers/Microsoft.Network/virtualNetworks/app-vnet"
	testPublicIp1 = "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/net-rg/providers/Microsoft.Network/publicIPAddresses/ssis-ip-1"
	testPublicIp2 = "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/net-rg/providers/Microsoft.Network/publicIPAddresses/ssis-ip-2"
)

// validResource returns a valid managed data-flow compute runtime
// (the flagship variant) that individual cases mutate into the shape
// under test.
func validResource() *AzureDataFactoryIntegrationRuntime {
	return &AzureDataFactoryIntegrationRuntime{
		ApiVersion: "azure.planton.dev/v1alpha1",
		Kind:       "AzureDataFactoryIntegrationRuntime",
		Metadata: &shared.CloudResourceMetadata{
			Name: "test-adf-ir",
		},
		Spec: &AzureDataFactoryIntegrationRuntimeSpec{
			DataFactoryId: literal(testFactoryId),
			Name:          "dataflow-compute",
			Variant: &AzureDataFactoryIntegrationRuntimeSpec_Azure{
				Azure: &AzureDataFactoryIntegrationRuntimeAzure{
					Region: "eastus",
				},
			},
		},
	}
}

// withoutVariant clears the fixture's azure variant so a case can
// install a different one.
func withoutVariant() *AzureDataFactoryIntegrationRuntime {
	input := validResource()
	input.Spec.Variant = nil
	return input
}

// minimalSsis builds the smallest valid SSIS variant.
func minimalSsis() *AzureDataFactoryIntegrationRuntimeAzureSsis {
	return &AzureDataFactoryIntegrationRuntimeAzureSsis{
		Region:   "eastus",
		NodeSize: "Standard_D2_v3",
	}
}

var _ = ginkgo.Describe("AzureDataFactoryIntegrationRuntimeSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {

		ginkgo.Context("variant selection -- one minimal document per variant", func() {

			ginkgo.It("should accept a minimal azure (data-flow compute) runtime", func() {
				gomega.Expect(protovalidate.Validate(validResource())).To(gomega.BeNil())
			})

			ginkgo.It("should accept a minimal azure_ssis runtime", func() {
				input := withoutVariant()
				input.Spec.Variant = &AzureDataFactoryIntegrationRuntimeSpec_AzureSsis{AzureSsis: minimalSsis()}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept a self_hosted registration as an empty block", func() {
				input := withoutVariant()
				input.Spec.Variant = &AzureDataFactoryIntegrationRuntimeSpec_SelfHosted{SelfHosted: &AzureDataFactoryIntegrationRuntimeSelfHosted{}}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})
		})

		ginkgo.Context("the azure variant's own contracts", func() {

			ginkgo.It("should accept the AutoResolve region", func() {
				input := validResource()
				input.Spec.GetAzure().Region = "AutoResolve"
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept a fully configured data-flow compute", func() {
				input := validResource()
				cleanup := false
				input.Spec.Variant = &AzureDataFactoryIntegrationRuntimeSpec_Azure{Azure: &AzureDataFactoryIntegrationRuntimeAzure{
					Region:                                  "eastus",
					CleanupEnabled:                          &cleanup,
					ComputeType:                             "MemoryOptimized",
					CoreCount:                               16,
					TimeToLiveMin:                           15,
					VirtualNetworkEnabled:                   true,
					InteractiveAuthoringTimeToLiveInMinutes: 30,
				}}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})
		})

		ginkgo.Context("the azure_ssis variant's own contracts", func() {

			ginkgo.It("should accept the scale and licensing knobs", func() {
				input := withoutVariant()
				ssis := minimalSsis()
				ssis.NumberOfNodes = 4
				ssis.MaxParallelExecutionsPerNode = 8
				ssis.Edition = "Enterprise"
				ssis.LicenseType = "BasePrice"
				ssis.CredentialName = "deploy-identity"
				input.Spec.Variant = &AzureDataFactoryIntegrationRuntimeSpec_AzureSsis{AzureSsis: ssis}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept a catalog on a pricing tier with SQL authentication", func() {
				input := withoutVariant()
				ssis := minimalSsis()
				ssis.CatalogInfo = &AzureDataFactoryIntegrationRuntimeSsisCatalogInfo{
					ServerEndpoint:        "catalog-sql.database.windows.net",
					AdministratorLogin:    "ssisadmin",
					AdministratorPassword: "correct-horse-battery-staple",
					PricingTier:           "S1",
				}
				input.Spec.Variant = &AzureDataFactoryIntegrationRuntimeSpec_AzureSsis{AzureSsis: ssis}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept a catalog in an elastic pool with managed-identity authentication", func() {
				input := withoutVariant()
				ssis := minimalSsis()
				ssis.CatalogInfo = &AzureDataFactoryIntegrationRuntimeSsisCatalogInfo{
					ServerEndpoint:  "catalog-sql.database.windows.net",
					ElasticPoolName: "ssis-pool",
				}
				input.Spec.Variant = &AzureDataFactoryIntegrationRuntimeSpec_AzureSsis{AzureSsis: ssis}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept a custom setup script container", func() {
				input := withoutVariant()
				ssis := minimalSsis()
				ssis.CustomSetupScript = &AzureDataFactoryIntegrationRuntimeSsisCustomSetupScript{
					BlobContainerUri: "https://setupsa.blob.core.windows.net/ssis-setup",
					SasToken:         "sv=2024-01-01&sig=redacted",
				}
				input.Spec.Variant = &AzureDataFactoryIntegrationRuntimeSpec_AzureSsis{AzureSsis: ssis}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept an express custom setup with only a PowerShell version", func() {
				input := withoutVariant()
				ssis := minimalSsis()
				ssis.ExpressCustomSetup = &AzureDataFactoryIntegrationRuntimeSsisExpressCustomSetup{
					PowershellVersion: "7.2.1",
				}
				input.Spec.Variant = &AzureDataFactoryIntegrationRuntimeSpec_AzureSsis{AzureSsis: ssis}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept a command key with an inline password", func() {
				input := withoutVariant()
				ssis := minimalSsis()
				ssis.ExpressCustomSetup = &AzureDataFactoryIntegrationRuntimeSsisExpressCustomSetup{
					CommandKey: []*AzureDataFactoryIntegrationRuntimeSsisCommandKey{{
						TargetName: "fileshare.corp.local",
						UserName:   "svc-ssis",
						Password:   "correct-horse-battery-staple",
					}},
				}
				input.Spec.Variant = &AzureDataFactoryIntegrationRuntimeSpec_AzureSsis{AzureSsis: ssis}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept a command key with a Key Vault password", func() {
				input := withoutVariant()
				ssis := minimalSsis()
				ssis.ExpressCustomSetup = &AzureDataFactoryIntegrationRuntimeSsisExpressCustomSetup{
					CommandKey: []*AzureDataFactoryIntegrationRuntimeSsisCommandKey{{
						TargetName: "fileshare.corp.local",
						UserName:   "svc-ssis",
						KeyVaultPassword: &AzureDataFactoryIntegrationRuntimeSsisKeyVaultSecretReference{
							LinkedServiceName: literal("kv-conn"),
							SecretName:        "fileshare-password",
						},
					}},
				}
				input.Spec.Variant = &AzureDataFactoryIntegrationRuntimeSpec_AzureSsis{AzureSsis: ssis}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept a licensed component with a Key Vault license and environment variables", func() {
				input := withoutVariant()
				ssis := minimalSsis()
				ssis.ExpressCustomSetup = &AzureDataFactoryIntegrationRuntimeSsisExpressCustomSetup{
					Environment: map[string]string{"SSIS_ENV": "prod"},
					Component: []*AzureDataFactoryIntegrationRuntimeSsisComponent{{
						Name: "KingswaySoft",
						KeyVaultLicense: &AzureDataFactoryIntegrationRuntimeSsisKeyVaultSecretReference{
							LinkedServiceName: literal("kv-conn"),
							SecretName:        "kingswaysoft-license",
							SecretVersion:     "3f2a",
						},
					}},
				}
				input.Spec.Variant = &AzureDataFactoryIntegrationRuntimeSpec_AzureSsis{AzureSsis: ssis}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept express virtual network injection", func() {
				input := withoutVariant()
				ssis := minimalSsis()
				ssis.ExpressVnetIntegration = &AzureDataFactoryIntegrationRuntimeSsisExpressVnetIntegration{
					SubnetId: literal(testSubnetId),
				}
				input.Spec.Variant = &AzureDataFactoryIntegrationRuntimeSpec_AzureSsis{AzureSsis: ssis}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept standard vnet integration addressed by subnet ID", func() {
				input := withoutVariant()
				ssis := minimalSsis()
				ssis.VnetIntegration = &AzureDataFactoryIntegrationRuntimeSsisVnetIntegration{
					SubnetId: literal(testSubnetId),
				}
				input.Spec.Variant = &AzureDataFactoryIntegrationRuntimeSpec_AzureSsis{AzureSsis: ssis}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept standard vnet integration addressed by network + subnet name with two public IPs", func() {
				input := withoutVariant()
				ssis := minimalSsis()
				ssis.VnetIntegration = &AzureDataFactoryIntegrationRuntimeSsisVnetIntegration{
					VnetId:     literal(testVnetId),
					SubnetName: "ssis-subnet",
					PublicIps: []*foreignkeyv1.StringValueOrRef{
						literal(testPublicIp1),
						literal(testPublicIp2),
					},
				}
				input.Spec.Variant = &AzureDataFactoryIntegrationRuntimeSpec_AzureSsis{AzureSsis: ssis}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept package stores", func() {
				input := withoutVariant()
				ssis := minimalSsis()
				ssis.PackageStore = []*AzureDataFactoryIntegrationRuntimeSsisPackageStore{{
					Name:              "shared-packages",
					LinkedServiceName: literal("fileshare-conn"),
				}}
				input.Spec.Variant = &AzureDataFactoryIntegrationRuntimeSpec_AzureSsis{AzureSsis: ssis}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept both compute-scale blocks", func() {
				input := withoutVariant()
				ssis := minimalSsis()
				ssis.CopyComputeScale = &AzureDataFactoryIntegrationRuntimeSsisCopyComputeScale{
					DataIntegrationUnit: 8,
					TimeToLive:          10,
				}
				ssis.PipelineExternalComputeScale = &AzureDataFactoryIntegrationRuntimeSsisPipelineExternalComputeScale{
					NumberOfExternalNodes: 2,
					NumberOfPipelineNodes: 3,
					TimeToLive:            5,
				}
				input.Spec.Variant = &AzureDataFactoryIntegrationRuntimeSpec_AzureSsis{AzureSsis: ssis}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept an on-premises proxy through a self-hosted runtime", func() {
				input := withoutVariant()
				ssis := minimalSsis()
				ssis.Proxy = &AzureDataFactoryIntegrationRuntimeSsisProxy{
					SelfHostedIntegrationRuntimeName: literal("onprem-bridge"),
					StagingStorageLinkedServiceName:  literal("staging-blob"),
					Path:                             "ssis-staging",
				}
				input.Spec.Variant = &AzureDataFactoryIntegrationRuntimeSpec_AzureSsis{AzureSsis: ssis}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})
		})

		ginkgo.Context("the self_hosted variant's own contracts", func() {

			ginkgo.It("should accept a linked registration through RBAC authorization", func() {
				input := withoutVariant()
				input.Spec.Name = "shared-bridge"
				input.Spec.Variant = &AzureDataFactoryIntegrationRuntimeSpec_SelfHosted{SelfHosted: &AzureDataFactoryIntegrationRuntimeSelfHosted{
					RbacAuthorization: &AzureDataFactoryIntegrationRuntimeSelfHostedRbacAuthorization{
						ResourceId: literal(testRuntimeId),
					},
				}}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept self-contained interactive authoring", func() {
				input := withoutVariant()
				input.Spec.Variant = &AzureDataFactoryIntegrationRuntimeSpec_SelfHosted{SelfHosted: &AzureDataFactoryIntegrationRuntimeSelfHosted{
					SelfContainedInteractiveAuthoringEnabled: true,
				}}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept a single-character self-hosted name (looser than the managed rule)", func() {
				input := withoutVariant()
				input.Spec.Name = "a"
				input.Spec.Variant = &AzureDataFactoryIntegrationRuntimeSpec_SelfHosted{SelfHosted: &AzureDataFactoryIntegrationRuntimeSelfHosted{}}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {

		ginkgo.Context("root contracts", func() {

			ginkgo.It("should reject a spec with no variant block", func() {
				input := withoutVariant()
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("carries exactly one variant by construction (setting a second arm replaces the first)", func() {
				input := validResource()
				input.Spec.Variant = &AzureDataFactoryIntegrationRuntimeSpec_SelfHosted{SelfHosted: &AzureDataFactoryIntegrationRuntimeSelfHosted{}}
				gomega.Expect(input.Spec.GetAzure()).To(gomega.BeNil())
				gomega.Expect(input.Spec.GetSelfHosted()).ToNot(gomega.BeNil())
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should reject a missing data_factory_id", func() {
				input := validResource()
				input.Spec.DataFactoryId = nil
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject a missing name", func() {
				input := validResource()
				input.Spec.Name = ""
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject a two-character managed runtime name (Azure's 3-character minimum)", func() {
				input := validResource()
				input.Spec.Name = "ab"
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject a managed runtime name with an underscore", func() {
				input := validResource()
				input.Spec.Name = "dataflow_compute"
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject a self-hosted name with consecutive dashes", func() {
				input := withoutVariant()
				input.Spec.Name = "onprem--bridge"
				input.Spec.Variant = &AzureDataFactoryIntegrationRuntimeSpec_SelfHosted{SelfHosted: &AzureDataFactoryIntegrationRuntimeSelfHosted{}}
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject a self-hosted name with a trailing dash", func() {
				input := withoutVariant()
				input.Spec.Name = "onprem-bridge-"
				input.Spec.Variant = &AzureDataFactoryIntegrationRuntimeSpec_SelfHosted{SelfHosted: &AzureDataFactoryIntegrationRuntimeSelfHosted{}}
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("the azure variant's contracts", func() {

			ginkgo.It("should reject a missing region", func() {
				input := validResource()
				input.Spec.GetAzure().Region = ""
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject a core count outside Azure's menu", func() {
				input := validResource()
				input.Spec.GetAzure().CoreCount = 12
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject an unknown compute type", func() {
				input := validResource()
				input.Spec.GetAzure().ComputeType = "GpuOptimized"
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject an interactive authoring TTL outside Azure's menu", func() {
				input := validResource()
				input.Spec.GetAzure().VirtualNetworkEnabled = true
				input.Spec.GetAzure().InteractiveAuthoringTimeToLiveInMinutes = 45
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject interactive authoring without the managed virtual network", func() {
				input := validResource()
				input.Spec.GetAzure().InteractiveAuthoringTimeToLiveInMinutes = 30
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("the azure_ssis variant's contracts", func() {

			ginkgo.It("should reject a missing node_size", func() {
				input := withoutVariant()
				ssis := minimalSsis()
				ssis.NodeSize = ""
				input.Spec.Variant = &AzureDataFactoryIntegrationRuntimeSpec_AzureSsis{AzureSsis: ssis}
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject a node size outside Azure's SSIS menu", func() {
				input := withoutVariant()
				ssis := minimalSsis()
				ssis.NodeSize = "Standard_B2s"
				input.Spec.Variant = &AzureDataFactoryIntegrationRuntimeSpec_AzureSsis{AzureSsis: ssis}
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject more than 10 nodes", func() {
				input := withoutVariant()
				ssis := minimalSsis()
				ssis.NumberOfNodes = 11
				input.Spec.Variant = &AzureDataFactoryIntegrationRuntimeSpec_AzureSsis{AzureSsis: ssis}
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject more than 16 parallel executions per node", func() {
				input := withoutVariant()
				ssis := minimalSsis()
				ssis.MaxParallelExecutionsPerNode = 17
				input.Spec.Variant = &AzureDataFactoryIntegrationRuntimeSpec_AzureSsis{AzureSsis: ssis}
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject an unknown edition", func() {
				input := withoutVariant()
				ssis := minimalSsis()
				ssis.Edition = "Developer"
				input.Spec.Variant = &AzureDataFactoryIntegrationRuntimeSpec_AzureSsis{AzureSsis: ssis}
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject a catalog without its server endpoint", func() {
				input := withoutVariant()
				ssis := minimalSsis()
				ssis.CatalogInfo = &AzureDataFactoryIntegrationRuntimeSsisCatalogInfo{
					PricingTier: "Basic",
				}
				input.Spec.Variant = &AzureDataFactoryIntegrationRuntimeSpec_AzureSsis{AzureSsis: ssis}
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject a catalog naming both a pricing tier and an elastic pool", func() {
				input := withoutVariant()
				ssis := minimalSsis()
				ssis.CatalogInfo = &AzureDataFactoryIntegrationRuntimeSsisCatalogInfo{
					ServerEndpoint:  "catalog-sql.database.windows.net",
					PricingTier:     "S1",
					ElasticPoolName: "ssis-pool",
				}
				input.Spec.Variant = &AzureDataFactoryIntegrationRuntimeSpec_AzureSsis{AzureSsis: ssis}
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject an unknown catalog pricing tier", func() {
				input := withoutVariant()
				ssis := minimalSsis()
				ssis.CatalogInfo = &AzureDataFactoryIntegrationRuntimeSsisCatalogInfo{
					ServerEndpoint: "catalog-sql.database.windows.net",
					PricingTier:    "S99",
				}
				input.Spec.Variant = &AzureDataFactoryIntegrationRuntimeSpec_AzureSsis{AzureSsis: ssis}
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject a setup script container without its SAS token", func() {
				input := withoutVariant()
				ssis := minimalSsis()
				ssis.CustomSetupScript = &AzureDataFactoryIntegrationRuntimeSsisCustomSetupScript{
					BlobContainerUri: "https://setupsa.blob.core.windows.net/ssis-setup",
				}
				input.Spec.Variant = &AzureDataFactoryIntegrationRuntimeSpec_AzureSsis{AzureSsis: ssis}
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject an express custom setup declaring nothing", func() {
				input := withoutVariant()
				ssis := minimalSsis()
				ssis.ExpressCustomSetup = &AzureDataFactoryIntegrationRuntimeSsisExpressCustomSetup{}
				input.Spec.Variant = &AzureDataFactoryIntegrationRuntimeSpec_AzureSsis{AzureSsis: ssis}
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject a command key without its target", func() {
				input := withoutVariant()
				ssis := minimalSsis()
				ssis.ExpressCustomSetup = &AzureDataFactoryIntegrationRuntimeSsisExpressCustomSetup{
					CommandKey: []*AzureDataFactoryIntegrationRuntimeSsisCommandKey{{
						UserName: "svc-ssis",
						Password: "correct-horse-battery-staple",
					}},
				}
				input.Spec.Variant = &AzureDataFactoryIntegrationRuntimeSpec_AzureSsis{AzureSsis: ssis}
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject a Key Vault reference without its secret name", func() {
				input := withoutVariant()
				ssis := minimalSsis()
				ssis.ExpressCustomSetup = &AzureDataFactoryIntegrationRuntimeSsisExpressCustomSetup{
					Component: []*AzureDataFactoryIntegrationRuntimeSsisComponent{{
						Name: "KingswaySoft",
						KeyVaultLicense: &AzureDataFactoryIntegrationRuntimeSsisKeyVaultSecretReference{
							LinkedServiceName: literal("kv-conn"),
						},
					}},
				}
				input.Spec.Variant = &AzureDataFactoryIntegrationRuntimeSpec_AzureSsis{AzureSsis: ssis}
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject vnet integration addressing both the network and the subnet", func() {
				input := withoutVariant()
				ssis := minimalSsis()
				ssis.VnetIntegration = &AzureDataFactoryIntegrationRuntimeSsisVnetIntegration{
					VnetId:     literal(testVnetId),
					SubnetId:   literal(testSubnetId),
					SubnetName: "ssis-subnet",
				}
				input.Spec.Variant = &AzureDataFactoryIntegrationRuntimeSpec_AzureSsis{AzureSsis: ssis}
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject vnet integration addressing neither form", func() {
				input := withoutVariant()
				ssis := minimalSsis()
				ssis.VnetIntegration = &AzureDataFactoryIntegrationRuntimeSsisVnetIntegration{}
				input.Spec.Variant = &AzureDataFactoryIntegrationRuntimeSpec_AzureSsis{AzureSsis: ssis}
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject vnet_id without its subnet_name", func() {
				input := withoutVariant()
				ssis := minimalSsis()
				ssis.VnetIntegration = &AzureDataFactoryIntegrationRuntimeSsisVnetIntegration{
					VnetId: literal(testVnetId),
				}
				input.Spec.Variant = &AzureDataFactoryIntegrationRuntimeSpec_AzureSsis{AzureSsis: ssis}
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject subnet_name alongside subnet_id (it pairs with vnet_id only)", func() {
				input := withoutVariant()
				ssis := minimalSsis()
				ssis.VnetIntegration = &AzureDataFactoryIntegrationRuntimeSsisVnetIntegration{
					SubnetId:   literal(testSubnetId),
					SubnetName: "ssis-subnet",
				}
				input.Spec.Variant = &AzureDataFactoryIntegrationRuntimeSpec_AzureSsis{AzureSsis: ssis}
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject a single public IP (Azure takes exactly two)", func() {
				input := withoutVariant()
				ssis := minimalSsis()
				ssis.VnetIntegration = &AzureDataFactoryIntegrationRuntimeSsisVnetIntegration{
					VnetId:     literal(testVnetId),
					SubnetName: "ssis-subnet",
					PublicIps:  []*foreignkeyv1.StringValueOrRef{literal(testPublicIp1)},
				}
				input.Spec.Variant = &AzureDataFactoryIntegrationRuntimeSpec_AzureSsis{AzureSsis: ssis}
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject a data integration unit that is not a multiple of 4", func() {
				input := withoutVariant()
				ssis := minimalSsis()
				ssis.CopyComputeScale = &AzureDataFactoryIntegrationRuntimeSsisCopyComputeScale{
					DataIntegrationUnit: 10,
				}
				input.Spec.Variant = &AzureDataFactoryIntegrationRuntimeSpec_AzureSsis{AzureSsis: ssis}
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject a copy compute time-to-live under 5 minutes", func() {
				input := withoutVariant()
				ssis := minimalSsis()
				ssis.CopyComputeScale = &AzureDataFactoryIntegrationRuntimeSsisCopyComputeScale{
					TimeToLive: 3,
				}
				input.Spec.Variant = &AzureDataFactoryIntegrationRuntimeSpec_AzureSsis{AzureSsis: ssis}
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject more than 10 external nodes", func() {
				input := withoutVariant()
				ssis := minimalSsis()
				ssis.PipelineExternalComputeScale = &AzureDataFactoryIntegrationRuntimeSsisPipelineExternalComputeScale{
					NumberOfExternalNodes: 11,
				}
				input.Spec.Variant = &AzureDataFactoryIntegrationRuntimeSpec_AzureSsis{AzureSsis: ssis}
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject a package store without its linked service", func() {
				input := withoutVariant()
				ssis := minimalSsis()
				ssis.PackageStore = []*AzureDataFactoryIntegrationRuntimeSsisPackageStore{{
					Name: "shared-packages",
				}}
				input.Spec.Variant = &AzureDataFactoryIntegrationRuntimeSpec_AzureSsis{AzureSsis: ssis}
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject a proxy without its staging linked service", func() {
				input := withoutVariant()
				ssis := minimalSsis()
				ssis.Proxy = &AzureDataFactoryIntegrationRuntimeSsisProxy{
					SelfHostedIntegrationRuntimeName: literal("onprem-bridge"),
				}
				input.Spec.Variant = &AzureDataFactoryIntegrationRuntimeSpec_AzureSsis{AzureSsis: ssis}
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("the self_hosted variant's contracts", func() {

			ginkgo.It("should reject an RBAC authorization without its resource ID", func() {
				input := withoutVariant()
				input.Spec.Variant = &AzureDataFactoryIntegrationRuntimeSpec_SelfHosted{SelfHosted: &AzureDataFactoryIntegrationRuntimeSelfHosted{
					RbacAuthorization: &AzureDataFactoryIntegrationRuntimeSelfHostedRbacAuthorization{},
				}}
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})
		})
	})
})
