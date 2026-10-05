package azurecontainerappenvironmentv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestAzureContainerAppEnvironmentSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureContainerAppEnvironmentSpec Validation Tests")
}

func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

func boolPtr(b bool) *bool { return &b }

func int32Ptr(i int32) *int32 { return &i }

// minimalSpec returns the smallest valid environment.
func minimalSpec() *AzureContainerAppEnvironment {
	return &AzureContainerAppEnvironment{
		ApiVersion: "azure.planton.dev/v1alpha1",
		Kind:       "AzureContainerAppEnvironment",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-env",
		},
		Spec: &AzureContainerAppEnvironmentSpec{
			Region:          "eastus",
			ResourceGroup:   literal("my-rg"),
			EnvironmentName: "my-container-env",
		},
	}
}

var _ = ginkgo.Describe("AzureContainerAppEnvironmentSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {

		ginkgo.It("accepts a minimal environment", func() {
			gomega.Expect(protovalidate.Validate(minimalSpec())).To(gomega.BeNil())
		})

		ginkgo.It("accepts VNet injection", func() {
			input := minimalSpec()
			input.Spec.InfrastructureSubnetId = literal("/subscriptions/sub/resourceGroups/rg/providers/Microsoft.Network/virtualNetworks/vnet/subnets/apps")
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("accepts a workspace with logs_destination unspecified (modules deploy log-analytics)", func() {
			input := minimalSpec()
			input.Spec.LogAnalyticsWorkspaceId = literal("/subscriptions/sub/resourceGroups/rg/providers/Microsoft.OperationalInsights/workspaces/law")
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("accepts LOG_ANALYTICS with a workspace", func() {
			input := minimalSpec()
			input.Spec.LogsDestination = AzureContainerAppEnvironmentLogsDestination_LOG_ANALYTICS
			input.Spec.LogAnalyticsWorkspaceId = literal("/subscriptions/sub/resourceGroups/rg/providers/Microsoft.OperationalInsights/workspaces/law")
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("accepts AZURE_MONITOR without a workspace", func() {
			input := minimalSpec()
			input.Spec.LogsDestination = AzureContainerAppEnvironmentLogsDestination_AZURE_MONITOR
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("accepts internal load balancing with a subnet", func() {
			input := minimalSpec()
			input.Spec.InternalLoadBalancerEnabled = boolPtr(true)
			input.Spec.InfrastructureSubnetId = literal("/subscriptions/sub/resourceGroups/rg/providers/Microsoft.Network/virtualNetworks/vnet/subnets/apps")
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("accepts zone redundancy with a subnet", func() {
			input := minimalSpec()
			input.Spec.ZoneRedundancyEnabled = boolPtr(true)
			input.Spec.InfrastructureSubnetId = literal("/subscriptions/sub/resourceGroups/rg/providers/Microsoft.Network/virtualNetworks/vnet/subnets/apps")
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("accepts a dedicated workload profile with instance counts", func() {
			input := minimalSpec()
			input.Spec.WorkloadProfiles = []*AzureContainerAppEnvironmentWorkloadProfile{{
				Name:                "dedicated-d4",
				WorkloadProfileType: AzureContainerAppEnvironmentWorkloadProfileType_D4,
				MinimumCount:        int32Ptr(0),
				MaximumCount:        int32Ptr(5),
			}}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("accepts a serverless GPU profile without instance counts", func() {
			input := minimalSpec()
			input.Spec.WorkloadProfiles = []*AzureContainerAppEnvironmentWorkloadProfile{{
				Name:                "gpu-serverless",
				WorkloadProfileType: AzureContainerAppEnvironmentWorkloadProfileType_CONSUMPTION_GPU_NC8AS_T4,
			}}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("accepts infrastructure_resource_group_name with workload profiles", func() {
			input := minimalSpec()
			input.Spec.WorkloadProfiles = []*AzureContainerAppEnvironmentWorkloadProfile{{
				Name:                "dedicated-d4",
				WorkloadProfileType: AzureContainerAppEnvironmentWorkloadProfileType_D4,
			}}
			input.Spec.InfrastructureResourceGroupName = "me-infra-rg"
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("accepts a custom DNS suffix with certificate", func() {
			input := minimalSpec()
			input.Spec.CustomDomain = &AzureContainerAppEnvironmentCustomDomain{
				DnsSuffix:             "apps.example.com",
				CertificateBlobBase64: "aGVsbG8=",
				CertificatePassword:   "pfx-password",
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("accepts a system-assigned identity", func() {
			input := minimalSpec()
			input.Spec.Identity = &AzureContainerAppEnvironmentIdentity{
				Type: AzureContainerAppEnvironmentIdentityType_SYSTEM_ASSIGNED,
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("accepts user-assigned identities", func() {
			input := minimalSpec()
			input.Spec.Identity = &AzureContainerAppEnvironmentIdentity{
				Type: AzureContainerAppEnvironmentIdentityType_USER_ASSIGNED,
				UserAssignedIdentityIds: []*foreignkeyv1.StringValueOrRef{
					literal("/subscriptions/sub/resourceGroups/rg/providers/Microsoft.ManagedIdentity/userAssignedIdentities/uai"),
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("accepts mTLS, public network access DISABLED, dapr AI connection string, and tags", func() {
			input := minimalSpec()
			input.Spec.MutualTlsEnabled = boolPtr(true)
			input.Spec.PublicNetworkAccess = AzureContainerAppEnvironmentPublicNetworkAccess_DISABLED
			input.Spec.DaprApplicationInsightsConnectionString = "InstrumentationKey=00000000-0000-0000-0000-000000000000"
			input.Spec.Tags = map[string]string{"team": "platform"}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {

		ginkgo.It("rejects a missing region", func() {
			input := minimalSpec()
			input.Spec.Region = ""
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects a missing resource group", func() {
			input := minimalSpec()
			input.Spec.ResourceGroup = nil
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects a missing environment name", func() {
			input := minimalSpec()
			input.Spec.EnvironmentName = ""
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects an environment name with a trailing hyphen", func() {
			input := minimalSpec()
			input.Spec.EnvironmentName = "my-env-"
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects an environment name with invalid characters", func() {
			input := minimalSpec()
			input.Spec.EnvironmentName = "my_env!"
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects an environment name longer than 60 characters", func() {
			input := minimalSpec()
			input.Spec.EnvironmentName = "a123456789a123456789a123456789a123456789a123456789a123456789a"
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects internal load balancing without a subnet", func() {
			input := minimalSpec()
			input.Spec.InternalLoadBalancerEnabled = boolPtr(true)
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects zone redundancy without a subnet", func() {
			input := minimalSpec()
			input.Spec.ZoneRedundancyEnabled = boolPtr(true)
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects LOG_ANALYTICS without a workspace", func() {
			input := minimalSpec()
			input.Spec.LogsDestination = AzureContainerAppEnvironmentLogsDestination_LOG_ANALYTICS
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects AZURE_MONITOR combined with a workspace", func() {
			input := minimalSpec()
			input.Spec.LogsDestination = AzureContainerAppEnvironmentLogsDestination_AZURE_MONITOR
			input.Spec.LogAnalyticsWorkspaceId = literal("/subscriptions/sub/resourceGroups/rg/providers/Microsoft.OperationalInsights/workspaces/law")
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects infrastructure_resource_group_name without workload profiles", func() {
			input := minimalSpec()
			input.Spec.InfrastructureResourceGroupName = "me-infra-rg"
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects public network access ENABLED with internal load balancing", func() {
			input := minimalSpec()
			input.Spec.PublicNetworkAccess = AzureContainerAppEnvironmentPublicNetworkAccess_ENABLED
			input.Spec.InternalLoadBalancerEnabled = boolPtr(true)
			input.Spec.InfrastructureSubnetId = literal("/subscriptions/sub/resourceGroups/rg/providers/Microsoft.Network/virtualNetworks/vnet/subnets/apps")
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects a workload profile without a name", func() {
			input := minimalSpec()
			input.Spec.WorkloadProfiles = []*AzureContainerAppEnvironmentWorkloadProfile{{
				WorkloadProfileType: AzureContainerAppEnvironmentWorkloadProfileType_D4,
			}}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects a workload profile with an unspecified type", func() {
			input := minimalSpec()
			input.Spec.WorkloadProfiles = []*AzureContainerAppEnvironmentWorkloadProfile{{
				Name: "pool",
			}}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects instance counts on a Consumption-family profile", func() {
			input := minimalSpec()
			input.Spec.WorkloadProfiles = []*AzureContainerAppEnvironmentWorkloadProfile{{
				Name:                "consumption-pinned",
				WorkloadProfileType: AzureContainerAppEnvironmentWorkloadProfileType_CONSUMPTION,
				MaximumCount:        int32Ptr(3),
			}}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects a custom domain without a certificate blob", func() {
			input := minimalSpec()
			input.Spec.CustomDomain = &AzureContainerAppEnvironmentCustomDomain{
				DnsSuffix:           "apps.example.com",
				CertificatePassword: "pfx-password",
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects a custom domain without a dns suffix", func() {
			input := minimalSpec()
			input.Spec.CustomDomain = &AzureContainerAppEnvironmentCustomDomain{
				CertificateBlobBase64: "aGVsbG8=",
				CertificatePassword:   "pfx-password",
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects USER_ASSIGNED identity without identity ids", func() {
			input := minimalSpec()
			input.Spec.Identity = &AzureContainerAppEnvironmentIdentity{
				Type: AzureContainerAppEnvironmentIdentityType_USER_ASSIGNED,
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects SYSTEM_ASSIGNED identity with identity ids", func() {
			input := minimalSpec()
			input.Spec.Identity = &AzureContainerAppEnvironmentIdentity{
				Type: AzureContainerAppEnvironmentIdentityType_SYSTEM_ASSIGNED,
				UserAssignedIdentityIds: []*foreignkeyv1.StringValueOrRef{
					literal("/subscriptions/sub/resourceGroups/rg/providers/Microsoft.ManagedIdentity/userAssignedIdentities/uai"),
				},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects an identity with an unspecified type", func() {
			input := minimalSpec()
			input.Spec.Identity = &AzureContainerAppEnvironmentIdentity{}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})
	})
})
