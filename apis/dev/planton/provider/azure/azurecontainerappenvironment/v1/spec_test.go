package azurecontainerappenvironmentv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/apis/dev/planton/shared"
	"github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"
	foreignkeyv1 "github.com/plantonhq/planton/apis/dev/planton/shared/foreignkey/v1"
)

func TestAzureContainerAppEnvironmentSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureContainerAppEnvironmentSpec Validation Tests")
}

// helper to create a minimal valid spec
func minimalSpec() *AzureContainerAppEnvironment {
	return &AzureContainerAppEnvironment{
		ApiVersion: "azure.planton.dev/v1",
		Kind:       "AzureContainerAppEnvironment",
		Metadata: &shared.CloudResourceMetadata{
			Name: "test-env",
		},
		Spec: &AzureContainerAppEnvironmentSpec{
			Region: "eastus",
			ResourceGroup: &foreignkeyv1.StringValueOrRef{
				LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
					Value: "my-rg",
				},
			},
			Name: "my-container-env",
		},
	}
}

var _ = ginkgo.Describe("AzureContainerAppEnvironmentSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("azure_container_app_environment", func() {

			ginkgo.It("should not return a validation error for a minimal environment", func() {
				input := minimalSpec()
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with VNet injection", func() {
				input := minimalSpec()
				input.Spec.InfrastructureSubnetId = &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
						Value: "/subscriptions/sub/resourceGroups/rg/providers/Microsoft.Network/virtualNetworks/vnet/subnets/apps",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with Log Analytics workspace", func() {
				input := minimalSpec()
				input.Spec.LogAnalyticsWorkspaceId = &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
						Value: "/subscriptions/sub/resourceGroups/rg/providers/Microsoft.OperationalInsights/workspaces/law",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with internal load balancer enabled", func() {
				ilb := true
				input := minimalSpec()
				input.Spec.InternalLoadBalancerEnabled = &ilb
				input.Spec.InfrastructureSubnetId = &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
						Value: "/subscriptions/sub/resourceGroups/rg/providers/Microsoft.Network/virtualNetworks/vnet/subnets/apps",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with zone redundancy enabled", func() {
				zr := true
				input := minimalSpec()
				input.Spec.ZoneRedundancyEnabled = &zr
				input.Spec.InfrastructureSubnetId = &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
						Value: "/subscriptions/sub/resourceGroups/rg/providers/Microsoft.Network/virtualNetworks/vnet/subnets/apps",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with workload profiles", func() {
				minCount := int32(1)
				maxCount := int32(10)
				input := minimalSpec()
				input.Spec.WorkloadProfiles = []*AzureContainerAppWorkloadProfile{
					{
						Name:                "gpu-pool",
						WorkloadProfileType: "NC24-A100",
						MinimumCount:        &minCount,
						MaximumCount:        &maxCount,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with multiple workload profiles", func() {
				minGp := int32(2)
				maxGp := int32(8)
				minMem := int32(0)
				maxMem := int32(4)
				input := minimalSpec()
				input.Spec.WorkloadProfiles = []*AzureContainerAppWorkloadProfile{
					{
						Name:                "general",
						WorkloadProfileType: "D4",
						MinimumCount:        &minGp,
						MaximumCount:        &maxGp,
					},
					{
						Name:                "high-memory",
						WorkloadProfileType: "E8",
						MinimumCount:        &minMem,
						MaximumCount:        &maxMem,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with workload profile minimum_count of 0", func() {
				minCount := int32(0)
				maxCount := int32(5)
				input := minimalSpec()
				input.Spec.WorkloadProfiles = []*AzureContainerAppWorkloadProfile{
					{
						Name:                "scale-to-zero",
						WorkloadProfileType: "D4",
						MinimumCount:        &minCount,
						MaximumCount:        &maxCount,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with all optional fields set", func() {
				ilb := true
				zr := true
				minCount := int32(2)
				maxCount := int32(10)
				input := minimalSpec()
				input.Metadata.Org = "mycompany"
				input.Metadata.Env = "production"
				input.Spec.InfrastructureSubnetId = &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
						Value: "/subscriptions/sub/resourceGroups/rg/providers/Microsoft.Network/virtualNetworks/vnet/subnets/apps",
					},
				}
				input.Spec.LogAnalyticsWorkspaceId = &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
						Value: "/subscriptions/sub/resourceGroups/rg/providers/Microsoft.OperationalInsights/workspaces/law",
					},
				}
				input.Spec.InternalLoadBalancerEnabled = &ilb
				input.Spec.ZoneRedundancyEnabled = &zr
				input.Spec.WorkloadProfiles = []*AzureContainerAppWorkloadProfile{
					{
						Name:                "dedicated",
						WorkloadProfileType: "D8",
						MinimumCount:        &minCount,
						MaximumCount:        &maxCount,
					},
				}
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

			ginkgo.It("should not return a validation error with valueFrom reference for subnet", func() {
				input := minimalSpec()
				input.Spec.InfrastructureSubnetId = &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
						ValueFrom: &foreignkeyv1.ValueFromRef{
							Kind:      cloudresourcekind.CloudResourceKind_AzureSubnet,
							Name:      "apps-subnet",
							FieldPath: "status.outputs.subnet_id",
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with valueFrom reference for workspace", func() {
				input := minimalSpec()
				input.Spec.LogAnalyticsWorkspaceId = &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
						ValueFrom: &foreignkeyv1.ValueFromRef{
							Kind:      cloudresourcekind.CloudResourceKind_AzureLogAnalyticsWorkspace,
							Name:      "central-law",
							FieldPath: "status.outputs.workspace_id",
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for name with hyphens", func() {
				input := minimalSpec()
				input.Spec.Name = "prod-apps-env"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for two-character name", func() {
				input := minimalSpec()
				input.Spec.Name = "ab"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("azure_container_app_environment", func() {

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

			ginkgo.It("should return a validation error when name is too short (1 char)", func() {
				input := minimalSpec()
				input.Spec.Name = "a"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when name exceeds 60 characters", func() {
				tooLong := "a"
				for len(tooLong) < 61 {
					tooLong += "b"
				}
				input := minimalSpec()
				input.Spec.Name = tooLong
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when name starts with a number", func() {
				input := minimalSpec()
				input.Spec.Name = "1bad-name"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when name starts with a hyphen", func() {
				input := minimalSpec()
				input.Spec.Name = "-bad-name"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when name ends with a hyphen", func() {
				input := minimalSpec()
				input.Spec.Name = "bad-name-"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when name contains uppercase letters", func() {
				input := minimalSpec()
				input.Spec.Name = "Bad-Name"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when name contains underscores", func() {
				input := minimalSpec()
				input.Spec.Name = "bad_name"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when name contains spaces", func() {
				input := minimalSpec()
				input.Spec.Name = "bad name"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when name contains dots", func() {
				input := minimalSpec()
				input.Spec.Name = "bad.name"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when workload profile name is empty", func() {
				maxCount := int32(5)
				input := minimalSpec()
				input.Spec.WorkloadProfiles = []*AzureContainerAppWorkloadProfile{
					{
						Name:                "",
						WorkloadProfileType: "D4",
						MaximumCount:        &maxCount,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when workload profile type is empty", func() {
				maxCount := int32(5)
				input := minimalSpec()
				input.Spec.WorkloadProfiles = []*AzureContainerAppWorkloadProfile{
					{
						Name:                "my-pool",
						WorkloadProfileType: "",
						MaximumCount:        &maxCount,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when workload profile minimum_count is negative", func() {
				minCount := int32(-1)
				input := minimalSpec()
				input.Spec.WorkloadProfiles = []*AzureContainerAppWorkloadProfile{
					{
						Name:                "my-pool",
						WorkloadProfileType: "D4",
						MinimumCount:        &minCount,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when workload profile maximum_count is negative", func() {
				maxCount := int32(-1)
				input := minimalSpec()
				input.Spec.WorkloadProfiles = []*AzureContainerAppWorkloadProfile{
					{
						Name:                "my-pool",
						WorkloadProfileType: "D4",
						MaximumCount:        &maxCount,
					},
				}
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
				input := &AzureContainerAppEnvironment{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "AzureContainerAppEnvironment",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-env",
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
