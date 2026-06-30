package azureloganalyticsworkspacev1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/apis/dev/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/apis/dev/planton/shared/foreignkey/v1"
)

func TestAzureLogAnalyticsWorkspaceSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureLogAnalyticsWorkspaceSpec Validation Tests")
}

var _ = ginkgo.Describe("AzureLogAnalyticsWorkspaceSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("azure_log_analytics_workspace", func() {

			ginkgo.It("should not return a validation error for minimal valid fields", func() {
				input := &AzureLogAnalyticsWorkspace{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "AzureLogAnalyticsWorkspace",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-law",
					},
					Spec: &AzureLogAnalyticsWorkspaceSpec{
						Region: "eastus",
						ResourceGroup: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "test-resource-group",
							},
						},
						Name: "test-workspace",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for production configuration", func() {
				input := &AzureLogAnalyticsWorkspace{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "AzureLogAnalyticsWorkspace",
					Metadata: &shared.CloudResourceMetadata{
						Name: "prod-law",
						Org:  "mycompany",
						Env:  "production",
					},
					Spec: &AzureLogAnalyticsWorkspaceSpec{
						Region: "westeurope",
						ResourceGroup: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "prod-monitoring-rg",
							},
						},
						Name:            "prod-platform-law",
						Sku:             strPtr("PerGB2018"),
						RetentionInDays: int32Ptr(90),
						DailyQuotaGb:    float64Ptr(10.0),
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with minimum retention days", func() {
				input := &AzureLogAnalyticsWorkspace{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "AzureLogAnalyticsWorkspace",
					Metadata: &shared.CloudResourceMetadata{
						Name: "min-retention-law",
					},
					Spec: &AzureLogAnalyticsWorkspaceSpec{
						Region: "eastus",
						ResourceGroup: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "test-rg",
							},
						},
						Name:            "min-retention-law",
						RetentionInDays: int32Ptr(30),
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with maximum retention days", func() {
				input := &AzureLogAnalyticsWorkspace{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "AzureLogAnalyticsWorkspace",
					Metadata: &shared.CloudResourceMetadata{
						Name: "max-retention-law",
					},
					Spec: &AzureLogAnalyticsWorkspaceSpec{
						Region: "eastus",
						ResourceGroup: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "test-rg",
							},
						},
						Name:            "max-retention-law",
						RetentionInDays: int32Ptr(730),
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with unlimited daily quota", func() {
				input := &AzureLogAnalyticsWorkspace{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "AzureLogAnalyticsWorkspace",
					Metadata: &shared.CloudResourceMetadata{
						Name: "unlimited-law",
					},
					Spec: &AzureLogAnalyticsWorkspaceSpec{
						Region: "eastus",
						ResourceGroup: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "test-rg",
							},
						},
						Name:         "unlimited-quota-law",
						DailyQuotaGb: float64Ptr(-1),
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with fractional daily quota", func() {
				input := &AzureLogAnalyticsWorkspace{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "AzureLogAnalyticsWorkspace",
					Metadata: &shared.CloudResourceMetadata{
						Name: "fractional-law",
					},
					Spec: &AzureLogAnalyticsWorkspaceSpec{
						Region: "eastus",
						ResourceGroup: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "test-rg",
							},
						},
						Name:         "fractional-quota-law",
						DailyQuotaGb: float64Ptr(0.5),
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("azure_log_analytics_workspace", func() {

			ginkgo.It("should return a validation error when region is missing", func() {
				input := &AzureLogAnalyticsWorkspace{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "AzureLogAnalyticsWorkspace",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-law",
					},
					Spec: &AzureLogAnalyticsWorkspaceSpec{
						ResourceGroup: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "test-rg",
							},
						},
						Name: "test-law",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when resource_group is missing", func() {
				input := &AzureLogAnalyticsWorkspace{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "AzureLogAnalyticsWorkspace",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-law",
					},
					Spec: &AzureLogAnalyticsWorkspaceSpec{
						Region: "eastus",
						Name:   "test-law",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when name is missing", func() {
				input := &AzureLogAnalyticsWorkspace{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "AzureLogAnalyticsWorkspace",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-law",
					},
					Spec: &AzureLogAnalyticsWorkspaceSpec{
						Region: "eastus",
						ResourceGroup: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "test-rg",
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when name is too short", func() {
				input := &AzureLogAnalyticsWorkspace{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "AzureLogAnalyticsWorkspace",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-law",
					},
					Spec: &AzureLogAnalyticsWorkspaceSpec{
						Region: "eastus",
						ResourceGroup: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "test-rg",
							},
						},
						Name: "abc", // 3 chars, minimum is 4
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when name exceeds maximum length", func() {
				longName := ""
				for len(longName) < 64 {
					longName += "a"
				}
				input := &AzureLogAnalyticsWorkspace{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "AzureLogAnalyticsWorkspace",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-law",
					},
					Spec: &AzureLogAnalyticsWorkspaceSpec{
						Region: "eastus",
						ResourceGroup: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "test-rg",
							},
						},
						Name: longName, // 64 chars, maximum is 63
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when retention_in_days is below minimum", func() {
				input := &AzureLogAnalyticsWorkspace{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "AzureLogAnalyticsWorkspace",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-law",
					},
					Spec: &AzureLogAnalyticsWorkspaceSpec{
						Region: "eastus",
						ResourceGroup: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "test-rg",
							},
						},
						Name:            "test-law-low-retention",
						RetentionInDays: int32Ptr(29), // minimum is 30
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when retention_in_days exceeds maximum", func() {
				input := &AzureLogAnalyticsWorkspace{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "AzureLogAnalyticsWorkspace",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-law",
					},
					Spec: &AzureLogAnalyticsWorkspaceSpec{
						Region: "eastus",
						ResourceGroup: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "test-rg",
							},
						},
						Name:            "test-law-high-retention",
						RetentionInDays: int32Ptr(731), // maximum is 730
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when daily_quota_gb is below -1", func() {
				input := &AzureLogAnalyticsWorkspace{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "AzureLogAnalyticsWorkspace",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-law",
					},
					Spec: &AzureLogAnalyticsWorkspaceSpec{
						Region: "eastus",
						ResourceGroup: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "test-rg",
							},
						},
						Name:         "test-law-bad-quota",
						DailyQuotaGb: float64Ptr(-2),
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when api_version is incorrect", func() {
				input := &AzureLogAnalyticsWorkspace{
					ApiVersion: "wrong.version/v1",
					Kind:       "AzureLogAnalyticsWorkspace",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-law",
					},
					Spec: &AzureLogAnalyticsWorkspaceSpec{
						Region: "eastus",
						ResourceGroup: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "test-rg",
							},
						},
						Name: "test-law",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when kind is incorrect", func() {
				input := &AzureLogAnalyticsWorkspace{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "WrongKind",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-law",
					},
					Spec: &AzureLogAnalyticsWorkspaceSpec{
						Region: "eastus",
						ResourceGroup: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "test-rg",
							},
						},
						Name: "test-law",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when metadata is missing", func() {
				input := &AzureLogAnalyticsWorkspace{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "AzureLogAnalyticsWorkspace",
					Spec: &AzureLogAnalyticsWorkspaceSpec{
						Region: "eastus",
						ResourceGroup: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "test-rg",
							},
						},
						Name: "test-law",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when spec is missing", func() {
				input := &AzureLogAnalyticsWorkspace{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "AzureLogAnalyticsWorkspace",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-law",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})
	})
})

// Helper functions for pointer types
func strPtr(s string) *string {
	return &s
}

func int32Ptr(i int32) *int32 {
	return &i
}

func float64Ptr(f float64) *float64 {
	return &f
}
