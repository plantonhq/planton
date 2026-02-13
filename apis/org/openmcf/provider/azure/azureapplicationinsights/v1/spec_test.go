package azureapplicationinsightsv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared"
	foreignkeyv1 "github.com/plantonhq/openmcf/apis/org/openmcf/shared/foreignkey/v1"
)

func TestAzureApplicationInsightsSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureApplicationInsightsSpec Validation Tests")
}

var _ = ginkgo.Describe("AzureApplicationInsightsSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("azure_application_insights", func() {

			ginkgo.It("should not return a validation error for minimal valid fields", func() {
				input := &AzureApplicationInsights{
					ApiVersion: "azure.openmcf.org/v1",
					Kind:       "AzureApplicationInsights",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-ai",
					},
					Spec: &AzureApplicationInsightsSpec{
						Region: "eastus",
						ResourceGroup: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "test-resource-group",
							},
						},
						Name: "test-app-insights",
						WorkspaceId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.OperationalInsights/workspaces/test-law",
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for production configuration", func() {
				input := &AzureApplicationInsights{
					ApiVersion: "azure.openmcf.org/v1",
					Kind:       "AzureApplicationInsights",
					Metadata: &shared.CloudResourceMetadata{
						Name: "prod-ai",
						Org:  "mycompany",
						Env:  "production",
					},
					Spec: &AzureApplicationInsightsSpec{
						Region: "westeurope",
						ResourceGroup: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "prod-monitoring-rg",
							},
						},
						Name:            "prod-platform-ai",
						ApplicationType: strPtr("web"),
						WorkspaceId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/prod-rg/providers/Microsoft.OperationalInsights/workspaces/prod-law",
							},
						},
						RetentionInDays:    int32Ptr(90),
						DailyDataCapInGb:   float64Ptr(100.0),
						SamplingPercentage: float64Ptr(50.0),
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with application_type java", func() {
				input := &AzureApplicationInsights{
					ApiVersion: "azure.openmcf.org/v1",
					Kind:       "AzureApplicationInsights",
					Metadata: &shared.CloudResourceMetadata{
						Name: "java-ai",
					},
					Spec: &AzureApplicationInsightsSpec{
						Region: "eastus",
						ResourceGroup: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "test-rg",
							},
						},
						Name:            "java-app-insights",
						ApplicationType: strPtr("java"),
						WorkspaceId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "/subscriptions/00000000/resourceGroups/rg/providers/Microsoft.OperationalInsights/workspaces/law",
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with application_type Node.JS", func() {
				input := &AzureApplicationInsights{
					ApiVersion: "azure.openmcf.org/v1",
					Kind:       "AzureApplicationInsights",
					Metadata: &shared.CloudResourceMetadata{
						Name: "node-ai",
					},
					Spec: &AzureApplicationInsightsSpec{
						Region: "eastus",
						ResourceGroup: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "test-rg",
							},
						},
						Name:            "node-app-insights",
						ApplicationType: strPtr("Node.JS"),
						WorkspaceId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "/subscriptions/00000000/resourceGroups/rg/providers/Microsoft.OperationalInsights/workspaces/law",
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with application_type other", func() {
				input := &AzureApplicationInsights{
					ApiVersion: "azure.openmcf.org/v1",
					Kind:       "AzureApplicationInsights",
					Metadata: &shared.CloudResourceMetadata{
						Name: "other-ai",
					},
					Spec: &AzureApplicationInsightsSpec{
						Region: "eastus",
						ResourceGroup: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "test-rg",
							},
						},
						Name:            "other-app-insights",
						ApplicationType: strPtr("other"),
						WorkspaceId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "/subscriptions/00000000/resourceGroups/rg/providers/Microsoft.OperationalInsights/workspaces/law",
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with all retention_in_days allowed values", func() {
				allowedValues := []int32{30, 60, 90, 120, 180, 270, 365, 550, 730}
				for _, days := range allowedValues {
					d := days
					input := &AzureApplicationInsights{
						ApiVersion: "azure.openmcf.org/v1",
						Kind:       "AzureApplicationInsights",
						Metadata: &shared.CloudResourceMetadata{
							Name: "retention-test-ai",
						},
						Spec: &AzureApplicationInsightsSpec{
							Region: "eastus",
							ResourceGroup: &foreignkeyv1.StringValueOrRef{
								LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
									Value: "test-rg",
								},
							},
							Name:            "retention-ai",
							RetentionInDays: &d,
							WorkspaceId: &foreignkeyv1.StringValueOrRef{
								LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
									Value: "/subscriptions/00000000/resourceGroups/rg/providers/Microsoft.OperationalInsights/workspaces/law",
								},
							},
						},
					}
					err := protovalidate.Validate(input)
					gomega.Expect(err).To(gomega.BeNil())
				}
			})

			ginkgo.It("should not return a validation error with zero daily_data_cap_in_gb", func() {
				input := &AzureApplicationInsights{
					ApiVersion: "azure.openmcf.org/v1",
					Kind:       "AzureApplicationInsights",
					Metadata: &shared.CloudResourceMetadata{
						Name: "zero-cap-ai",
					},
					Spec: &AzureApplicationInsightsSpec{
						Region: "eastus",
						ResourceGroup: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "test-rg",
							},
						},
						Name:             "zero-cap-ai",
						DailyDataCapInGb: float64Ptr(0),
						WorkspaceId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "/subscriptions/00000000/resourceGroups/rg/providers/Microsoft.OperationalInsights/workspaces/law",
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with fractional daily_data_cap_in_gb", func() {
				input := &AzureApplicationInsights{
					ApiVersion: "azure.openmcf.org/v1",
					Kind:       "AzureApplicationInsights",
					Metadata: &shared.CloudResourceMetadata{
						Name: "fractional-cap-ai",
					},
					Spec: &AzureApplicationInsightsSpec{
						Region: "eastus",
						ResourceGroup: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "test-rg",
							},
						},
						Name:             "fractional-cap-ai",
						DailyDataCapInGb: float64Ptr(0.5),
						WorkspaceId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "/subscriptions/00000000/resourceGroups/rg/providers/Microsoft.OperationalInsights/workspaces/law",
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with sampling_percentage at boundaries", func() {
				for _, pct := range []float64{0, 50, 100} {
					p := pct
					input := &AzureApplicationInsights{
						ApiVersion: "azure.openmcf.org/v1",
						Kind:       "AzureApplicationInsights",
						Metadata: &shared.CloudResourceMetadata{
							Name: "sampling-ai",
						},
						Spec: &AzureApplicationInsightsSpec{
							Region: "eastus",
							ResourceGroup: &foreignkeyv1.StringValueOrRef{
								LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
									Value: "test-rg",
								},
							},
							Name:               "sampling-ai",
							SamplingPercentage: &p,
							WorkspaceId: &foreignkeyv1.StringValueOrRef{
								LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
									Value: "/subscriptions/00000000/resourceGroups/rg/providers/Microsoft.OperationalInsights/workspaces/law",
								},
							},
						},
					}
					err := protovalidate.Validate(input)
					gomega.Expect(err).To(gomega.BeNil())
				}
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("azure_application_insights", func() {

			ginkgo.It("should return a validation error when region is missing", func() {
				input := &AzureApplicationInsights{
					ApiVersion: "azure.openmcf.org/v1",
					Kind:       "AzureApplicationInsights",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-ai",
					},
					Spec: &AzureApplicationInsightsSpec{
						ResourceGroup: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "test-rg",
							},
						},
						Name: "test-ai",
						WorkspaceId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "/subscriptions/00000000/resourceGroups/rg/providers/Microsoft.OperationalInsights/workspaces/law",
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when resource_group is missing", func() {
				input := &AzureApplicationInsights{
					ApiVersion: "azure.openmcf.org/v1",
					Kind:       "AzureApplicationInsights",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-ai",
					},
					Spec: &AzureApplicationInsightsSpec{
						Region: "eastus",
						Name:   "test-ai",
						WorkspaceId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "/subscriptions/00000000/resourceGroups/rg/providers/Microsoft.OperationalInsights/workspaces/law",
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when name is missing", func() {
				input := &AzureApplicationInsights{
					ApiVersion: "azure.openmcf.org/v1",
					Kind:       "AzureApplicationInsights",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-ai",
					},
					Spec: &AzureApplicationInsightsSpec{
						Region: "eastus",
						ResourceGroup: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "test-rg",
							},
						},
						WorkspaceId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "/subscriptions/00000000/resourceGroups/rg/providers/Microsoft.OperationalInsights/workspaces/law",
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when workspace_id is missing", func() {
				input := &AzureApplicationInsights{
					ApiVersion: "azure.openmcf.org/v1",
					Kind:       "AzureApplicationInsights",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-ai",
					},
					Spec: &AzureApplicationInsightsSpec{
						Region: "eastus",
						ResourceGroup: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "test-rg",
							},
						},
						Name: "test-ai",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for invalid application_type", func() {
				input := &AzureApplicationInsights{
					ApiVersion: "azure.openmcf.org/v1",
					Kind:       "AzureApplicationInsights",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-ai",
					},
					Spec: &AzureApplicationInsightsSpec{
						Region: "eastus",
						ResourceGroup: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "test-rg",
							},
						},
						Name:            "test-ai",
						ApplicationType: strPtr("invalid-type"),
						WorkspaceId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "/subscriptions/00000000/resourceGroups/rg/providers/Microsoft.OperationalInsights/workspaces/law",
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for disallowed retention_in_days value", func() {
				input := &AzureApplicationInsights{
					ApiVersion: "azure.openmcf.org/v1",
					Kind:       "AzureApplicationInsights",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-ai",
					},
					Spec: &AzureApplicationInsightsSpec{
						Region: "eastus",
						ResourceGroup: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "test-rg",
							},
						},
						Name:            "test-ai",
						RetentionInDays: int32Ptr(45), // not in allowed list
						WorkspaceId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "/subscriptions/00000000/resourceGroups/rg/providers/Microsoft.OperationalInsights/workspaces/law",
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when daily_data_cap_in_gb is negative", func() {
				input := &AzureApplicationInsights{
					ApiVersion: "azure.openmcf.org/v1",
					Kind:       "AzureApplicationInsights",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-ai",
					},
					Spec: &AzureApplicationInsightsSpec{
						Region: "eastus",
						ResourceGroup: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "test-rg",
							},
						},
						Name:             "test-ai",
						DailyDataCapInGb: float64Ptr(-1),
						WorkspaceId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "/subscriptions/00000000/resourceGroups/rg/providers/Microsoft.OperationalInsights/workspaces/law",
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when sampling_percentage exceeds 100", func() {
				input := &AzureApplicationInsights{
					ApiVersion: "azure.openmcf.org/v1",
					Kind:       "AzureApplicationInsights",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-ai",
					},
					Spec: &AzureApplicationInsightsSpec{
						Region: "eastus",
						ResourceGroup: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "test-rg",
							},
						},
						Name:               "test-ai",
						SamplingPercentage: float64Ptr(101),
						WorkspaceId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "/subscriptions/00000000/resourceGroups/rg/providers/Microsoft.OperationalInsights/workspaces/law",
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when sampling_percentage is negative", func() {
				input := &AzureApplicationInsights{
					ApiVersion: "azure.openmcf.org/v1",
					Kind:       "AzureApplicationInsights",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-ai",
					},
					Spec: &AzureApplicationInsightsSpec{
						Region: "eastus",
						ResourceGroup: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "test-rg",
							},
						},
						Name:               "test-ai",
						SamplingPercentage: float64Ptr(-1),
						WorkspaceId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "/subscriptions/00000000/resourceGroups/rg/providers/Microsoft.OperationalInsights/workspaces/law",
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when api_version is incorrect", func() {
				input := &AzureApplicationInsights{
					ApiVersion: "wrong.version/v1",
					Kind:       "AzureApplicationInsights",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-ai",
					},
					Spec: &AzureApplicationInsightsSpec{
						Region: "eastus",
						ResourceGroup: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "test-rg",
							},
						},
						Name: "test-ai",
						WorkspaceId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "/subscriptions/00000000/resourceGroups/rg/providers/Microsoft.OperationalInsights/workspaces/law",
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when kind is incorrect", func() {
				input := &AzureApplicationInsights{
					ApiVersion: "azure.openmcf.org/v1",
					Kind:       "WrongKind",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-ai",
					},
					Spec: &AzureApplicationInsightsSpec{
						Region: "eastus",
						ResourceGroup: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "test-rg",
							},
						},
						Name: "test-ai",
						WorkspaceId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "/subscriptions/00000000/resourceGroups/rg/providers/Microsoft.OperationalInsights/workspaces/law",
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when metadata is missing", func() {
				input := &AzureApplicationInsights{
					ApiVersion: "azure.openmcf.org/v1",
					Kind:       "AzureApplicationInsights",
					Spec: &AzureApplicationInsightsSpec{
						Region: "eastus",
						ResourceGroup: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "test-rg",
							},
						},
						Name: "test-ai",
						WorkspaceId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "/subscriptions/00000000/resourceGroups/rg/providers/Microsoft.OperationalInsights/workspaces/law",
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when spec is missing", func() {
				input := &AzureApplicationInsights{
					ApiVersion: "azure.openmcf.org/v1",
					Kind:       "AzureApplicationInsights",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-ai",
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
