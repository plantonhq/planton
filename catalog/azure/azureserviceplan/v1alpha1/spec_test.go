package azureserviceplanv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	"github.com/plantonhq/planton/shared/catalogkind"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestAzureServicePlanSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureServicePlanSpec Validation Tests")
}

// helper to create a minimal valid spec (Linux, P1v3)
func minimalSpec() *AzureServicePlan {
	return &AzureServicePlan{
		ApiVersion: "azure.planton.dev/v1alpha1",
		Kind:       "AzureServicePlan",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-plan",
		},
		Spec: &AzureServicePlanSpec{
			Region: "eastus",
			ResourceGroup: &foreignkeyv1.StringValueOrRef{
				LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
					Value: "my-rg",
				},
			},
			ServicePlanName: "myapp-plan",
			SkuName:         AzureServicePlanSku_PREMIUM_P1V3,
		},
	}
}

var _ = ginkgo.Describe("AzureServicePlanSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("azure_service_plan", func() {

			ginkgo.It("should not return a validation error for a minimal Linux plan", func() {
				input := minimalSpec()
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for a Windows plan", func() {
				input := minimalSpec()
				input.Spec.OsType = AzureServicePlanOsType_WINDOWS
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for a Windows Container plan", func() {
				input := minimalSpec()
				input.Spec.OsType = AzureServicePlanOsType_WINDOWS_CONTAINER
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for a Linux plan with explicit os_type", func() {
				input := minimalSpec()
				input.Spec.OsType = AzureServicePlanOsType_LINUX
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for a Consumption plan (Y1)", func() {
				input := minimalSpec()
				input.Spec.SkuName = AzureServicePlanSku_CONSUMPTION_Y1
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for an Elastic Premium plan with a scale-out ceiling", func() {
				maxElastic := int32(50)
				input := minimalSpec()
				input.Spec.SkuName = AzureServicePlanSku_ELASTIC_PREMIUM_EP1
				input.Spec.MaximumElasticWorkerCount = &maxElastic
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for a Workflow plan with a scale-out ceiling", func() {
				maxElastic := int32(20)
				input := minimalSpec()
				input.Spec.SkuName = AzureServicePlanSku_WORKFLOW_WS1
				input.Spec.MaximumElasticWorkerCount = &maxElastic
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for a Basic plan (B1)", func() {
				input := minimalSpec()
				input.Spec.SkuName = AzureServicePlanSku_BASIC_B1
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for a Standard plan (S1)", func() {
				input := minimalSpec()
				input.Spec.SkuName = AzureServicePlanSku_STANDARD_S1
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with worker_count set", func() {
				workers := int32(3)
				input := minimalSpec()
				input.Spec.WorkerCount = &workers
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with zone balancing on a Premium plan", func() {
				zoneBalancing := true
				workers := int32(3)
				input := minimalSpec()
				input.Spec.ZoneBalancingEnabled = &zoneBalancing
				input.Spec.WorkerCount = &workers
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with zone balancing on an Elastic Premium plan", func() {
				zoneBalancing := true
				input := minimalSpec()
				input.Spec.SkuName = AzureServicePlanSku_ELASTIC_PREMIUM_EP1
				input.Spec.ZoneBalancingEnabled = &zoneBalancing
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with zone balancing on an Isolated v2 plan", func() {
				zoneBalancing := true
				input := minimalSpec()
				input.Spec.SkuName = AzureServicePlanSku_ISOLATED_I1V2
				input.Spec.ZoneBalancingEnabled = &zoneBalancing
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with per_site_scaling_enabled set", func() {
				perSite := true
				input := minimalSpec()
				input.Spec.PerSiteScalingEnabled = &perSite
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with maximum_elastic_worker_count of 0", func() {
				maxElastic := int32(0)
				input := minimalSpec()
				input.Spec.SkuName = AzureServicePlanSku_ELASTIC_PREMIUM_EP2
				input.Spec.MaximumElasticWorkerCount = &maxElastic
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for premium auto-scale on a Premium v3 plan", func() {
				autoScale := true
				input := minimalSpec()
				input.Spec.PremiumPlanAutoScaleEnabled = &autoScale
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for a scale-out ceiling on a Premium plan with auto-scale on", func() {
				autoScale := true
				maxElastic := int32(10)
				input := minimalSpec()
				input.Spec.PremiumPlanAutoScaleEnabled = &autoScale
				input.Spec.MaximumElasticWorkerCount = &maxElastic
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for an Isolated v2 plan inside an App Service Environment", func() {
				input := minimalSpec()
				input.Spec.SkuName = AzureServicePlanSku_ISOLATED_I1V2
				input.Spec.AppServiceEnvironmentId = "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/my-rg/providers/Microsoft.Web/hostingEnvironments/my-ase"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with user tags", func() {
				input := minimalSpec()
				input.Spec.Tags = map[string]string{"cost-center": "platform", "owner": "web-team"}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with all optional fields set", func() {
				workers := int32(6)
				zoneBalancing := true
				perSite := true
				maxElastic := int32(30)
				input := minimalSpec()
				input.Metadata.Org = "mycompany"
				input.Metadata.Env = "production"
				input.Spec.OsType = AzureServicePlanOsType_LINUX
				input.Spec.SkuName = AzureServicePlanSku_ELASTIC_PREMIUM_EP1
				input.Spec.WorkerCount = &workers
				input.Spec.ZoneBalancingEnabled = &zoneBalancing
				input.Spec.PerSiteScalingEnabled = &perSite
				input.Spec.MaximumElasticWorkerCount = &maxElastic
				input.Spec.Tags = map[string]string{"team": "platform"}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with valueFrom reference for resource_group", func() {
				input := minimalSpec()
				input.Spec.ResourceGroup = &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
						ValueFrom: &foreignkeyv1.ValueFromRef{
							Kind:      catalogkind.CatalogKind_AzureResourceGroup,
							Name:      "shared-rg",
							FieldPath: "status.outputs.resource_group_name",
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for plan name with hyphens and underscores", func() {
				input := minimalSpec()
				input.Spec.ServicePlanName = "my-app_plan-01"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for plan name starting with a number", func() {
				input := minimalSpec()
				input.Spec.ServicePlanName = "01-plan"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for worker_count of 1", func() {
				workers := int32(1)
				input := minimalSpec()
				input.Spec.WorkerCount = &workers
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("azure_service_plan", func() {

			ginkgo.It("should return a validation error when region is missing", func() {
				input := minimalSpec()
				input.Spec.Region = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when resource_group is missing", func() {
				input := minimalSpec()
				input.Spec.ResourceGroup = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when service_plan_name is missing", func() {
				input := minimalSpec()
				input.Spec.ServicePlanName = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when service_plan_name exceeds 60 characters", func() {
				tooLong := ""
				for len(tooLong) < 61 {
					tooLong += "a"
				}
				input := minimalSpec()
				input.Spec.ServicePlanName = tooLong
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when service_plan_name contains spaces", func() {
				input := minimalSpec()
				input.Spec.ServicePlanName = "my plan"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when service_plan_name contains special characters", func() {
				input := minimalSpec()
				input.Spec.ServicePlanName = "my.plan@test"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when sku_name is unspecified", func() {
				input := minimalSpec()
				input.Spec.SkuName = AzureServicePlanSku_azure_service_plan_sku_unspecified
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when worker_count is zero", func() {
				workers := int32(0)
				input := minimalSpec()
				input.Spec.WorkerCount = &workers
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when worker_count is negative", func() {
				workers := int32(-1)
				input := minimalSpec()
				input.Spec.WorkerCount = &workers
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when maximum_elastic_worker_count is negative", func() {
				maxElastic := int32(-1)
				input := minimalSpec()
				input.Spec.MaximumElasticWorkerCount = &maxElastic
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for zone balancing on a Basic plan", func() {
				zoneBalancing := true
				input := minimalSpec()
				input.Spec.SkuName = AzureServicePlanSku_BASIC_B1
				input.Spec.ZoneBalancingEnabled = &zoneBalancing
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for zone balancing on a Standard plan", func() {
				zoneBalancing := true
				input := minimalSpec()
				input.Spec.SkuName = AzureServicePlanSku_STANDARD_S3
				input.Spec.ZoneBalancingEnabled = &zoneBalancing
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for premium auto-scale on a Standard plan", func() {
				autoScale := true
				input := minimalSpec()
				input.Spec.SkuName = AzureServicePlanSku_STANDARD_S1
				input.Spec.PremiumPlanAutoScaleEnabled = &autoScale
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for premium auto-scale on an Elastic Premium plan", func() {
				autoScale := true
				input := minimalSpec()
				input.Spec.SkuName = AzureServicePlanSku_ELASTIC_PREMIUM_EP1
				input.Spec.PremiumPlanAutoScaleEnabled = &autoScale
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a scale-out ceiling above 1 on a Premium plan without auto-scale", func() {
				maxElastic := int32(5)
				input := minimalSpec()
				input.Spec.MaximumElasticWorkerCount = &maxElastic
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for an App Service Environment on a Premium plan", func() {
				input := minimalSpec()
				input.Spec.AppServiceEnvironmentId = "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/my-rg/providers/Microsoft.Web/hostingEnvironments/my-ase"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a malformed App Service Environment ID", func() {
				input := minimalSpec()
				input.Spec.SkuName = AzureServicePlanSku_ISOLATED_I1V2
				input.Spec.AppServiceEnvironmentId = "my-ase"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when metadata is missing", func() {
				input := minimalSpec()
				input.Metadata = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when spec is missing", func() {
				input := &AzureServicePlan{
					ApiVersion: "azure.planton.dev/v1alpha1",
					Kind:       "AzureServicePlan",
					Metadata: &shared.CatalogObjectMetadata{
						Name: "test-plan",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when api_version is incorrect", func() {
				input := minimalSpec()
				input.ApiVersion = "wrong.version/v1"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when kind is incorrect", func() {
				input := minimalSpec()
				input.Kind = "WrongKind"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})
	})
})
