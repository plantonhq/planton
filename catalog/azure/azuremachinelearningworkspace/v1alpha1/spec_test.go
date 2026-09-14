package azuremachinelearningworkspacev1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
	"google.golang.org/protobuf/proto"
)

func TestAzureMachineLearningWorkspaceSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureMachineLearningWorkspaceSpec Validation Tests")
}

// literal builds a StringValueOrRef carrying a literal value.
func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

const (
	testInsightsId = "/subscriptions/s/resourceGroups/rg/providers/Microsoft.Insights/components/ml-insights"
	testVaultId    = "/subscriptions/s/resourceGroups/rg/providers/Microsoft.KeyVault/vaults/ml-vault"
	testStorageId  = "/subscriptions/s/resourceGroups/rg/providers/Microsoft.Storage/storageAccounts/mlstorage"
	testRegistryId = "/subscriptions/s/resourceGroups/rg/providers/Microsoft.ContainerRegistry/registries/mlacr"
	testSubnetId   = "/subscriptions/s/resourceGroups/rg/providers/Microsoft.Network/virtualNetworks/vnet/subnets/snet"
	testUaiId      = "/subscriptions/s/resourceGroups/rg/providers/Microsoft.ManagedIdentity/userAssignedIdentities/ml-uai"
)

// validResource returns a minimal valid workspace that individual
// cases mutate into the shape under test.
func validResource() *AzureMachineLearningWorkspace {
	return &AzureMachineLearningWorkspace{
		ApiVersion: "azure.planton.dev/v1alpha1",
		Kind:       "AzureMachineLearningWorkspace",
		Metadata: &shared.CloudResourceMetadata{
			Name: "test-ml-workspace",
		},
		Spec: &AzureMachineLearningWorkspaceSpec{
			Region:                "eastus",
			ResourceGroup:         literal("ml-rg"),
			Name:                  "ml-workspace",
			ApplicationInsightsId: literal(testInsightsId),
			KeyVaultId:            literal(testVaultId),
			StorageAccountId:      literal(testStorageId),
			Identity: &AzureMachineLearningWorkspaceIdentity{
				Type: AzureMachineLearningWorkspaceIdentityType_SYSTEM_ASSIGNED,
			},
		},
	}
}

var _ = ginkgo.Describe("AzureMachineLearningWorkspaceSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("azure_machine_learning_workspace", func() {

			ginkgo.It("should not return a validation error for a minimal workspace", func() {
				err := protovalidate.Validate(validResource())
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a feature-store workspace with its block", func() {
				input := validResource()
				input.Spec.Kind = AzureMachineLearningWorkspaceKind_FEATURE_STORE
				input.Spec.FeatureStore = &AzureMachineLearningWorkspaceFeatureStore{
					ComputerSparkRuntimeVersion: "3.4",
					OfflineConnectionName:       "offline-conn",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a feature_store block switched off on a DEFAULT workspace and refuse it on FEATURE_STORE", func() {
				input := validResource()
				input.Spec.FeatureStore = &AzureMachineLearningWorkspaceFeatureStore{Enabled: proto.Bool(false), ComputerSparkRuntimeVersion: "3.4"}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())

				input.Spec.Kind = AzureMachineLearningWorkspaceKind_FEATURE_STORE
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
				gomega.Expect(err.Error()).To(gomega.ContainSubstring("switch on"))
			})

			ginkgo.It("should accept CMK encryption with service-side encryption enabled", func() {
				input := validResource()
				input.Spec.Encryption = &AzureMachineLearningWorkspaceEncryption{
					KeyVaultId:             literal(testVaultId),
					KeyId:                  literal("https://ml-vault.vault.azure.net/keys/ml-cmk"),
					UserAssignedIdentityId: literal(testUaiId),
				}
				input.Spec.ServiceSideEncryptionEnabled = true
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a user-assigned identity with a primary identity", func() {
				input := validResource()
				input.Spec.Identity = &AzureMachineLearningWorkspaceIdentity{
					Type:        AzureMachineLearningWorkspaceIdentityType_USER_ASSIGNED,
					IdentityIds: []*foreignkeyv1.StringValueOrRef{literal(testUaiId)},
				}
				input.Spec.PrimaryUserAssignedIdentity = literal(testUaiId)
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a managed network with all three outbound rule types", func() {
				input := validResource()
				input.Spec.ManagedNetwork = &AzureMachineLearningWorkspaceManagedNetwork{
					IsolationMode:              AzureMachineLearningWorkspaceIsolationMode_ALLOW_ONLY_APPROVED_OUTBOUND,
					ProvisionOnCreationEnabled: true,
				}
				input.Spec.FqdnOutboundRules = []*AzureMachineLearningWorkspaceFqdnOutboundRule{
					{Name: "allow-pypi", DestinationFqdn: "pypi.org"},
				}
				input.Spec.PrivateEndpointOutboundRules = []*AzureMachineLearningWorkspacePrivateEndpointOutboundRule{
					{Name: "to-vault", ServiceResourceId: literal(testVaultId), SubResourceTarget: "vault"},
				}
				input.Spec.ServiceTagOutboundRules = []*AzureMachineLearningWorkspaceServiceTagOutboundRule{
					{Name: "allow-storage", ServiceTag: "Storage", Protocol: "TCP", PortRanges: "443"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept serverless compute without public IPs when a subnet is set", func() {
				input := validResource()
				input.Spec.PublicNetworkAccessEnabled = proto.Bool(false)
				input.Spec.ServerlessCompute = &AzureMachineLearningWorkspaceServerlessCompute{
					SubnetId:        literal(testSubnetId),
					PublicIpEnabled: false,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept serverless compute without a subnet while the workspace is public", func() {
				input := validResource()
				input.Spec.ServerlessCompute = &AzureMachineLearningWorkspaceServerlessCompute{
					PublicIpEnabled: false,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a private-endpoint rule whose target is a reference", func() {
				input := validResource()
				input.Spec.PrivateEndpointOutboundRules = []*AzureMachineLearningWorkspacePrivateEndpointOutboundRule{
					{
						Name: "to-workspace-storage",
						ServiceResourceId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
								ValueFrom: &foreignkeyv1.ValueFromRef{Name: "some-storage-account"},
							},
						},
						SubResourceTarget: "blob",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept the explicit Basic sku and container registry attachment", func() {
				input := validResource()
				input.Spec.SkuName = "Basic"
				input.Spec.ContainerRegistryId = literal(testRegistryId)
				input.Spec.HighBusinessImpact = true
				input.Spec.StorageAccountAccessType = AzureMachineLearningWorkspaceStorageAccountAccessType_IDENTITY
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("azure_machine_learning_workspace", func() {

			ginkgo.It("should reject a missing region", func() {
				input := validResource()
				input.Spec.Region = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a missing resource group", func() {
				input := validResource()
				input.Spec.ResourceGroup = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a name shorter than three characters", func() {
				input := validResource()
				input.Spec.Name = "ml"
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a name with a leading hyphen", func() {
				input := validResource()
				input.Spec.Name = "-ml-workspace"
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a missing application insights reference", func() {
				input := validResource()
				input.Spec.ApplicationInsightsId = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a missing key vault reference", func() {
				input := validResource()
				input.Spec.KeyVaultId = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a missing storage account reference", func() {
				input := validResource()
				input.Spec.StorageAccountId = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a missing identity", func() {
				input := validResource()
				input.Spec.Identity = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a user-assigned identity without identity ids", func() {
				input := validResource()
				input.Spec.Identity = &AzureMachineLearningWorkspaceIdentity{
					Type: AzureMachineLearningWorkspaceIdentityType_USER_ASSIGNED,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a system-assigned identity carrying identity ids", func() {
				input := validResource()
				input.Spec.Identity = &AzureMachineLearningWorkspaceIdentity{
					Type:        AzureMachineLearningWorkspaceIdentityType_SYSTEM_ASSIGNED,
					IdentityIds: []*foreignkeyv1.StringValueOrRef{literal(testUaiId)},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a feature_store block on a default-kind workspace", func() {
				input := validResource()
				input.Spec.FeatureStore = &AzureMachineLearningWorkspaceFeatureStore{
					OfflineConnectionName: "offline-conn",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject kind FEATURE_STORE without the feature_store block", func() {
				input := validResource()
				input.Spec.Kind = AzureMachineLearningWorkspaceKind_FEATURE_STORE
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject service-side encryption without the encryption block", func() {
				input := validResource()
				input.Spec.ServiceSideEncryptionEnabled = true
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject an encryption block without a key", func() {
				input := validResource()
				input.Spec.Encryption = &AzureMachineLearningWorkspaceEncryption{
					KeyVaultId: literal(testVaultId),
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject no-public-ip serverless compute without a subnet on a private workspace", func() {
				input := validResource()
				input.Spec.PublicNetworkAccessEnabled = proto.Bool(false)
				input.Spec.ServerlessCompute = &AzureMachineLearningWorkspaceServerlessCompute{
					PublicIpEnabled: false,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a sku outside the vocabulary", func() {
				input := validResource()
				input.Spec.SkuName = "Premium"
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject duplicate rule names across different outbound rule types", func() {
				input := validResource()
				input.Spec.FqdnOutboundRules = []*AzureMachineLearningWorkspaceFqdnOutboundRule{
					{Name: "shared-name", DestinationFqdn: "pypi.org"},
				}
				input.Spec.ServiceTagOutboundRules = []*AzureMachineLearningWorkspaceServiceTagOutboundRule{
					{Name: "shared-name", ServiceTag: "Storage", Protocol: "TCP", PortRanges: "443"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a key vault private-endpoint target with a storage sub-resource", func() {
				input := validResource()
				input.Spec.PrivateEndpointOutboundRules = []*AzureMachineLearningWorkspacePrivateEndpointOutboundRule{
					{Name: "to-vault", ServiceResourceId: literal(testVaultId), SubResourceTarget: "blob"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a storage private-endpoint target with the vault sub-resource", func() {
				input := validResource()
				input.Spec.PrivateEndpointOutboundRules = []*AzureMachineLearningWorkspacePrivateEndpointOutboundRule{
					{Name: "to-storage", ServiceResourceId: literal(testStorageId), SubResourceTarget: "vault"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a service tag outside the provider's allowlist", func() {
				input := validResource()
				input.Spec.ServiceTagOutboundRules = []*AzureMachineLearningWorkspaceServiceTagOutboundRule{
					{Name: "bad-tag", ServiceTag: "MyCustomTag", Protocol: "TCP", PortRanges: "443"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a protocol outside the vocabulary", func() {
				input := validResource()
				input.Spec.ServiceTagOutboundRules = []*AzureMachineLearningWorkspaceServiceTagOutboundRule{
					{Name: "bad-proto", ServiceTag: "Storage", Protocol: "SCTP", PortRanges: "443"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a service-tag rule without port ranges", func() {
				input := validResource()
				input.Spec.ServiceTagOutboundRules = []*AzureMachineLearningWorkspaceServiceTagOutboundRule{
					{Name: "no-ports", ServiceTag: "Storage", Protocol: "TCP"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})
	})
})
