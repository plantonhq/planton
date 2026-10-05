package azureloganalyticsworkspacev1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
	"google.golang.org/protobuf/proto"
)

func TestAzureLogAnalyticsWorkspaceSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureLogAnalyticsWorkspaceSpec Validation Tests")
}

// buildValidWorkspace returns a minimal valid resource; tests mutate copies of
// it to probe individual rules.
func buildValidWorkspace() *AzureLogAnalyticsWorkspace {
	return &AzureLogAnalyticsWorkspace{
		ApiVersion: "azure.planton.dev/v1alpha1",
		Kind:       "AzureLogAnalyticsWorkspace",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-law",
		},
		Spec: &AzureLogAnalyticsWorkspaceSpec{
			Region: "eastus",
			ResourceGroup: &foreignkeyv1.StringValueOrRef{
				LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
					Value: "test-resource-group",
				},
			},
			WorkspaceName: "test-workspace",
		},
	}
}

var _ = ginkgo.Describe("AzureLogAnalyticsWorkspaceSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {

		ginkgo.It("should not return a validation error for minimal valid fields", func() {
			err := protovalidate.Validate(buildValidWorkspace())
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept a full production configuration", func() {
			input := buildValidWorkspace()
			input.Metadata.Org = "mycompany"
			input.Metadata.Env = "production"
			input.Spec.Sku = AzureLogAnalyticsWorkspaceSku_PER_GB_2018
			input.Spec.RetentionInDays = proto.Int32(365)
			input.Spec.DailyQuotaGb = proto.Float64(50)
			input.Spec.LocalAuthenticationEnabled = proto.Bool(false)
			input.Spec.InternetIngestionEnabled = proto.Bool(false)
			input.Spec.InternetQueryEnabled = proto.Bool(false)
			input.Spec.AllowResourceOnlyPermissions = proto.Bool(false)
			input.Spec.CmkForQueryForced = true
			input.Spec.ImmediateDataPurgeOn_30DaysEnabled = true
			input.Spec.Tags = map[string]string{"cost-center": "platform"}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept the CAPACITY_RESERVATION sku with a commitment tier", func() {
			input := buildValidWorkspace()
			input.Spec.Sku = AzureLogAnalyticsWorkspaceSku_CAPACITY_RESERVATION
			input.Spec.ReservationCapacityInGbPerDay = proto.Int32(100)
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept the highest commitment tier", func() {
			input := buildValidWorkspace()
			input.Spec.Sku = AzureLogAnalyticsWorkspaceSku_CAPACITY_RESERVATION
			input.Spec.ReservationCapacityInGbPerDay = proto.Int32(50000)
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept legacy skus without a commitment tier", func() {
			input := buildValidWorkspace()
			input.Spec.Sku = AzureLogAnalyticsWorkspaceSku_PER_NODE
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
			input.Spec.Sku = AzureLogAnalyticsWorkspaceSku_STANDALONE
			err = protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept retention boundaries", func() {
			input := buildValidWorkspace()
			input.Spec.RetentionInDays = proto.Int32(30)
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			input.Spec.RetentionInDays = proto.Int32(730)
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept -1 (unlimited) and positive daily quotas", func() {
			input := buildValidWorkspace()
			input.Spec.DailyQuotaGb = proto.Float64(-1)
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			input.Spec.DailyQuotaGb = proto.Float64(0.5)
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept a system-assigned identity", func() {
			input := buildValidWorkspace()
			input.Spec.Identity = &AzureLogAnalyticsWorkspaceIdentity{
				Type: AzureLogAnalyticsWorkspaceIdentityType_SYSTEM_ASSIGNED,
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept a user-assigned identity with identity ids", func() {
			input := buildValidWorkspace()
			input.Spec.Identity = &AzureLogAnalyticsWorkspaceIdentity{
				Type: AzureLogAnalyticsWorkspaceIdentityType_USER_ASSIGNED,
				UserAssignedIdentityIds: []*foreignkeyv1.StringValueOrRef{
					{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
						Value: "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/rg/providers/Microsoft.ManagedIdentity/userAssignedIdentities/uai",
					}},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept a data collection rule id and workspace names at the length boundaries", func() {
			input := buildValidWorkspace()
			input.Spec.DataCollectionRuleId = &foreignkeyv1.StringValueOrRef{
				LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
					Value: "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/rg/providers/Microsoft.Insights/dataCollectionRules/default-dcr",
				},
			}
			input.Spec.WorkspaceName = "ab1c"
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			// 63 characters -- the upper boundary.
			input.Spec.WorkspaceName = "a12345678901234567890123456789012345678901234567890123456789zz"
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {

		ginkgo.It("should reject a missing region", func() {
			input := buildValidWorkspace()
			input.Spec.Region = ""
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a missing resource group", func() {
			input := buildValidWorkspace()
			input.Spec.ResourceGroup = nil
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a missing workspace name", func() {
			input := buildValidWorkspace()
			input.Spec.WorkspaceName = ""
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a workspace name shorter than 4 characters", func() {
			input := buildValidWorkspace()
			input.Spec.WorkspaceName = "ab1"
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a workspace name starting with a hyphen", func() {
			input := buildValidWorkspace()
			input.Spec.WorkspaceName = "-workspace"
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("start and end with a letter or digit"))
		})

		ginkgo.It("should reject a workspace name ending with a hyphen", func() {
			input := buildValidWorkspace()
			input.Spec.WorkspaceName = "workspace-"
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a workspace name with invalid characters", func() {
			input := buildValidWorkspace()
			input.Spec.WorkspaceName = "work_space"
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject an undefined sku enum number", func() {
			input := buildValidWorkspace()
			input.Spec.Sku = AzureLogAnalyticsWorkspaceSku(99)
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a commitment tier on the pay-as-you-go sku", func() {
			input := buildValidWorkspace()
			input.Spec.Sku = AzureLogAnalyticsWorkspaceSku_PER_GB_2018
			input.Spec.ReservationCapacityInGbPerDay = proto.Int32(100)
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("CAPACITY_RESERVATION"))
		})

		ginkgo.It("should reject a commitment tier when sku is unspecified", func() {
			input := buildValidWorkspace()
			input.Spec.ReservationCapacityInGbPerDay = proto.Int32(200)
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject the CAPACITY_RESERVATION sku without a commitment tier", func() {
			input := buildValidWorkspace()
			input.Spec.Sku = AzureLogAnalyticsWorkspaceSku_CAPACITY_RESERVATION
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("reservation_capacity_in_gb_per_day"))
		})

		ginkgo.It("should reject a commitment tier outside Azure's fixed set", func() {
			input := buildValidWorkspace()
			input.Spec.Sku = AzureLogAnalyticsWorkspaceSku_CAPACITY_RESERVATION
			input.Spec.ReservationCapacityInGbPerDay = proto.Int32(150)
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject retention below 30 days", func() {
			input := buildValidWorkspace()
			input.Spec.RetentionInDays = proto.Int32(29)
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject retention above 730 days", func() {
			input := buildValidWorkspace()
			input.Spec.RetentionInDays = proto.Int32(731)
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a daily quota below -1", func() {
			input := buildValidWorkspace()
			input.Spec.DailyQuotaGb = proto.Float64(-2)
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject an identity block without a type", func() {
			input := buildValidWorkspace()
			input.Spec.Identity = &AzureLogAnalyticsWorkspaceIdentity{}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject USER_ASSIGNED identity without identity ids", func() {
			input := buildValidWorkspace()
			input.Spec.Identity = &AzureLogAnalyticsWorkspaceIdentity{
				Type: AzureLogAnalyticsWorkspaceIdentityType_USER_ASSIGNED,
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("user_assigned_identity_ids"))
		})

		ginkgo.It("should reject SYSTEM_ASSIGNED identity carrying identity ids", func() {
			input := buildValidWorkspace()
			input.Spec.Identity = &AzureLogAnalyticsWorkspaceIdentity{
				Type: AzureLogAnalyticsWorkspaceIdentityType_SYSTEM_ASSIGNED,
				UserAssignedIdentityIds: []*foreignkeyv1.StringValueOrRef{
					{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "some-identity"}},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})
	})
})
