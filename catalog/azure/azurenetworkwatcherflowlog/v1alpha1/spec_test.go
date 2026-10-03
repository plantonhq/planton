package azurenetworkwatcherflowlogv1alpha1

import (
	"strings"
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestAzureNetworkWatcherFlowLogSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureNetworkWatcherFlowLogSpec Validation Tests")
}

// literal builds a StringValueOrRef carrying a literal value.
func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

func int32Ptr(v int32) *int32 { return &v }

func boolPtr(v bool) *bool { return &v }

const (
	testVnetId    = "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/platform-rg/providers/Microsoft.Network/virtualNetworks/platform-vnet"
	testSubnetId  = testVnetId + "/subnets/app-subnet"
	testNicId     = "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/platform-rg/providers/Microsoft.Network/networkInterfaces/app-nic"
	testNsgId     = "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/platform-rg/providers/Microsoft.Network/networkSecurityGroups/app-nsg"
	testStorageId = "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/platform-rg/providers/Microsoft.Storage/storageAccounts/platformflowlogs"
	testLawArmId  = "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/platform-rg/providers/Microsoft.OperationalInsights/workspaces/platform-law"
	testLawGuid   = "11111111-2222-3333-4444-555555555555"
)

// validTrafficAnalytics returns a fully-populated Traffic Analytics
// block.
func validTrafficAnalytics() *AzureNetworkWatcherFlowLogTrafficAnalytics {
	return &AzureNetworkWatcherFlowLogTrafficAnalytics{
		WorkspaceId:         literal(testLawGuid),
		WorkspaceRegion:     "eastus",
		WorkspaceResourceId: literal(testLawArmId),
	}
}

// validResource returns a minimal valid flow log (a virtual-network
// target on the region's auto-created watcher) that individual cases
// mutate into the shape under test.
func validResource() *AzureNetworkWatcherFlowLog {
	return &AzureNetworkWatcherFlowLog{
		ApiVersion: "azure.planton.dev/v1alpha1",
		Kind:       "AzureNetworkWatcherFlowLog",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-flow-log",
		},
		Spec: &AzureNetworkWatcherFlowLogSpec{
			Region:           "eastus",
			Name:             "vnet-flow-log",
			TargetResourceId: literal(testVnetId),
			StorageAccountId: literal(testStorageId),
			RetentionPolicy: &AzureNetworkWatcherFlowLogRetentionPolicy{
				Enabled: true,
				Days:    7,
			},
		},
	}
}

var _ = ginkgo.Describe("AzureNetworkWatcherFlowLogSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("azure_network_watcher_flow_log", func() {

			ginkgo.It("should not return a validation error for a minimal virtual-network flow log", func() {
				err := protovalidate.Validate(validResource())
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a subnet target", func() {
				input := validResource()
				input.Spec.TargetResourceId = literal(testSubnetId)
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a network-interface target", func() {
				input := validResource()
				input.Spec.TargetResourceId = literal(testNicId)
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept schema version 2 with collection paused", func() {
				input := validResource()
				input.Spec.Version = int32Ptr(2)
				input.Spec.Enabled = boolPtr(false)
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a full Traffic Analytics block at the 10-minute interval", func() {
				input := validResource()
				input.Spec.TrafficAnalytics = validTrafficAnalytics()
				input.Spec.TrafficAnalytics.IntervalInMinutes = int32Ptr(10)
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a self-managed watcher addressed by name and resource group", func() {
				input := validResource()
				input.Spec.NetworkWatcherName = "custom-watcher"
				input.Spec.NetworkWatcherResourceGroup = literal("watcher-rg")
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a single-character name", func() {
				input := validResource()
				input.Spec.Name = "a"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept an 80-character name ending in an underscore", func() {
				input := validResource()
				input.Spec.Name = "f" + strings.Repeat("a", 78) + "_"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept retention disabled with zero days (keep forever)", func() {
				input := validResource()
				input.Spec.RetentionPolicy = &AzureNetworkWatcherFlowLogRetentionPolicy{
					Enabled: false,
					Days:    0,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("azure_network_watcher_flow_log", func() {

			ginkgo.It("should reject a missing region", func() {
				input := validResource()
				input.Spec.Region = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a missing name", func() {
				input := validResource()
				input.Spec.Name = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a name starting with an underscore", func() {
				input := validResource()
				input.Spec.Name = "_flowlog"
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a name over 80 characters", func() {
				input := validResource()
				input.Spec.Name = "f" + strings.Repeat("a", 80)
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a missing target", func() {
				input := validResource()
				input.Spec.TargetResourceId = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject an NSG target (retired by Azure for new flow logs)", func() {
				input := validResource()
				input.Spec.TargetResourceId = literal(testNsgId)
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a missing storage account", func() {
				input := validResource()
				input.Spec.StorageAccountId = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a missing retention policy", func() {
				input := validResource()
				input.Spec.RetentionPolicy = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject negative retention days", func() {
				input := validResource()
				input.Spec.RetentionPolicy.Days = -1
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject schema version 0", func() {
				input := validResource()
				input.Spec.Version = int32Ptr(0)
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject schema version 3", func() {
				input := validResource()
				input.Spec.Version = int32Ptr(3)
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject Traffic Analytics without the workspace GUID", func() {
				input := validResource()
				input.Spec.TrafficAnalytics = validTrafficAnalytics()
				input.Spec.TrafficAnalytics.WorkspaceId = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject Traffic Analytics without the workspace region", func() {
				input := validResource()
				input.Spec.TrafficAnalytics = validTrafficAnalytics()
				input.Spec.TrafficAnalytics.WorkspaceRegion = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject Traffic Analytics without the workspace ARM id", func() {
				input := validResource()
				input.Spec.TrafficAnalytics = validTrafficAnalytics()
				input.Spec.TrafficAnalytics.WorkspaceResourceId = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a 30-minute Traffic Analytics interval", func() {
				input := validResource()
				input.Spec.TrafficAnalytics = validTrafficAnalytics()
				input.Spec.TrafficAnalytics.IntervalInMinutes = int32Ptr(30)
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a watcher name without its resource group", func() {
				input := validResource()
				input.Spec.NetworkWatcherName = "custom-watcher"
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a watcher resource group without its name", func() {
				input := validResource()
				input.Spec.NetworkWatcherResourceGroup = literal("watcher-rg")
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})
	})
})
