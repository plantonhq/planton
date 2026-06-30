package azureserviceplanv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/apis/dev/planton/shared"
	"github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"
	foreignkeyv1 "github.com/plantonhq/planton/apis/dev/planton/shared/foreignkey/v1"
)

func TestAzureServicePlanSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureServicePlanSpec Validation Tests")
}

// helper to create a minimal valid spec (Linux, P1v3)
func minimalSpec() *AzureServicePlan {
	return &AzureServicePlan{
		ApiVersion: "azure.planton.dev/v1",
		Kind:       "AzureServicePlan",
		Metadata: &shared.CloudResourceMetadata{
			Name: "test-plan",
		},
		Spec: &AzureServicePlanSpec{
			Region: "eastus",
			ResourceGroup: &foreignkeyv1.StringValueOrRef{
				LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
					Value: "my-rg",
				},
			},
			Name:    "myapp-plan",
			SkuName: "P1v3",
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
				osType := "Windows"
				input := minimalSpec()
				input.Spec.OsType = &osType
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for a Linux plan with explicit os_type", func() {
				osType := "Linux"
				input := minimalSpec()
				input.Spec.OsType = &osType
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for a Consumption plan (Y1)", func() {
				input := minimalSpec()
				input.Spec.SkuName = "Y1"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for an Elastic Premium plan (EP1)", func() {
				maxElastic := int32(50)
				input := minimalSpec()
				input.Spec.SkuName = "EP1"
				input.Spec.MaximumElasticWorkerCount = &maxElastic
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for a Basic plan (B1)", func() {
				input := minimalSpec()
				input.Spec.SkuName = "B1"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for a Standard plan (S1)", func() {
				input := minimalSpec()
				input.Spec.SkuName = "S1"
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

			ginkgo.It("should not return a validation error with zone_balancing_enabled set", func() {
				zoneBalancing := true
				workers := int32(3)
				input := minimalSpec()
				input.Spec.ZoneBalancingEnabled = &zoneBalancing
				input.Spec.WorkerCount = &workers
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
				input.Spec.SkuName = "EP2"
				input.Spec.MaximumElasticWorkerCount = &maxElastic
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with all optional fields set", func() {
				osType := "Linux"
				workers := int32(6)
				zoneBalancing := true
				perSite := true
				maxElastic := int32(30)
				input := minimalSpec()
				input.Metadata.Org = "mycompany"
				input.Metadata.Env = "production"
				input.Spec.OsType = &osType
				input.Spec.SkuName = "EP1"
				input.Spec.WorkerCount = &workers
				input.Spec.ZoneBalancingEnabled = &zoneBalancing
				input.Spec.PerSiteScalingEnabled = &perSite
				input.Spec.MaximumElasticWorkerCount = &maxElastic
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with valueFrom reference for resource_group", func() {
				input := minimalSpec()
				input.Spec.ResourceGroup = &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
						ValueFrom: &foreignkeyv1.ValueFromRef{
							Kind:      cloudresourcekind.CloudResourceKind_AzureResourceGroup,
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
				input.Spec.Name = "my-app_plan-01"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for plan name starting with a number", func() {
				input := minimalSpec()
				input.Spec.Name = "01-plan"
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

			ginkgo.It("should return a validation error when name is missing", func() {
				input := minimalSpec()
				input.Spec.Name = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when name exceeds 60 characters", func() {
				tooLong := ""
				for len(tooLong) < 61 {
					tooLong += "a"
				}
				input := minimalSpec()
				input.Spec.Name = tooLong
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when name contains spaces", func() {
				input := minimalSpec()
				input.Spec.Name = "my plan"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when name contains special characters", func() {
				input := minimalSpec()
				input.Spec.Name = "my.plan@test"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when sku_name is missing", func() {
				input := minimalSpec()
				input.Spec.SkuName = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when os_type is invalid", func() {
				invalidOs := "MacOS"
				input := minimalSpec()
				input.Spec.OsType = &invalidOs
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when os_type uses wrong casing", func() {
				invalidOs := "linux"
				input := minimalSpec()
				input.Spec.OsType = &invalidOs
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

			ginkgo.It("should return a validation error when metadata is missing", func() {
				input := minimalSpec()
				input.Metadata = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when spec is missing", func() {
				input := &AzureServicePlan{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "AzureServicePlan",
					Metadata: &shared.CloudResourceMetadata{
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
