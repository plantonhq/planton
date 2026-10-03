package azuremachinelearningcomputeclusterv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestAzureMachineLearningComputeClusterSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureMachineLearningComputeClusterSpec Validation Tests")
}

// literal builds a StringValueOrRef carrying a literal value.
func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

const (
	testWorkspaceId = "/subscriptions/s/resourceGroups/rg/providers/Microsoft.MachineLearningServices/workspaces/ml-workspace"
	testSubnetId    = "/subscriptions/s/resourceGroups/rg/providers/Microsoft.Network/virtualNetworks/vnet/subnets/ml-subnet"
	testIdentityId  = "/subscriptions/s/resourceGroups/rg/providers/Microsoft.ManagedIdentity/userAssignedIdentities/ml-uai"
)

// validResource returns a minimal valid scale-to-zero cluster that
// individual cases mutate into the shape under test.
func validResource() *AzureMachineLearningComputeCluster {
	return &AzureMachineLearningComputeCluster{
		ApiVersion: "azure.planton.dev/v1alpha1",
		Kind:       "AzureMachineLearningComputeCluster",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-ml-compute-cluster",
		},
		Spec: &AzureMachineLearningComputeClusterSpec{
			WorkspaceId: literal(testWorkspaceId),
			Name:        "cpu-cluster",
			Region:      "eastus",
			VmSize:      "STANDARD_DS2_V2",
			VmPriority:  AzureMachineLearningComputeClusterVmPriority_DEDICATED,
			ScaleSettings: &AzureMachineLearningComputeClusterScaleSettings{
				MinNodeCount:                    0,
				MaxNodeCount:                    1,
				ScaleDownNodesAfterIdleDuration: "PT30M",
			},
		},
	}
}

var _ = ginkgo.Describe("AzureMachineLearningComputeClusterSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("azure_machine_learning_compute_cluster", func() {

			ginkgo.It("should not return a validation error for a minimal scale-to-zero cluster", func() {
				err := protovalidate.Validate(validResource())
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a low-priority cluster with a wide scale range", func() {
				input := validResource()
				input.Spec.VmPriority = AzureMachineLearningComputeClusterVmPriority_LOW_PRIORITY
				input.Spec.ScaleSettings = &AzureMachineLearningComputeClusterScaleSettings{
					MinNodeCount:                    0,
					MaxNodeCount:                    16,
					ScaleDownNodesAfterIdleDuration: "PT2M",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a system-assigned identity", func() {
				input := validResource()
				input.Spec.Identity = &AzureMachineLearningComputeClusterIdentity{
					Type: AzureMachineLearningComputeClusterIdentityType_SYSTEM_ASSIGNED,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a user-assigned identity with identity ids", func() {
				input := validResource()
				input.Spec.Identity = &AzureMachineLearningComputeClusterIdentity{
					Type:        AzureMachineLearningComputeClusterIdentityType_USER_ASSIGNED,
					IdentityIds: []*foreignkeyv1.StringValueOrRef{literal(testIdentityId)},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept an ssh block with only a public key", func() {
				input := validResource()
				input.Spec.Ssh = &AzureMachineLearningComputeClusterSsh{
					AdminUsername: "azureuser",
					KeyValue:      "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQ example",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept an ssh block with only a password", func() {
				input := validResource()
				input.Spec.Ssh = &AzureMachineLearningComputeClusterSsh{
					AdminUsername: "azureuser",
					AdminPassword: literal("a-password-from-a-secret"),
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a VNet-only cluster with a subnet and no public IPs", func() {
				input := validResource()
				input.Spec.SubnetId = literal(testSubnetId)
				nodePublicIp := false
				input.Spec.NodePublicIpEnabled = &nodePublicIp
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a keyless cluster with local auth disabled", func() {
				input := validResource()
				localAuth := false
				input.Spec.LocalAuthEnabled = &localAuth
				input.Spec.Identity = &AzureMachineLearningComputeClusterIdentity{
					Type: AzureMachineLearningComputeClusterIdentityType_SYSTEM_ASSIGNED,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a 32-character name at the upper bound", func() {
				input := validResource()
				input.Spec.Name = "a123456789012345678901234567890b"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("azure_machine_learning_compute_cluster", func() {

			ginkgo.It("should reject a missing workspace reference", func() {
				input := validResource()
				input.Spec.WorkspaceId = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a name starting with a digit", func() {
				input := validResource()
				input.Spec.Name = "1cpu-cluster"
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a 2-character name below the lower bound", func() {
				input := validResource()
				input.Spec.Name = "ab"
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a name ending with a hyphen", func() {
				input := validResource()
				input.Spec.Name = "cpu-cluster-"
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a name with an underscore", func() {
				input := validResource()
				input.Spec.Name = "cpu_cluster"
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a missing region", func() {
				input := validResource()
				input.Spec.Region = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a missing vm size", func() {
				input := validResource()
				input.Spec.VmSize = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject an unspecified vm priority", func() {
				input := validResource()
				input.Spec.VmPriority = AzureMachineLearningComputeClusterVmPriority_azure_machine_learning_compute_cluster_vm_priority_unspecified
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject missing scale settings", func() {
				input := validResource()
				input.Spec.ScaleSettings = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a max node count below the min node count", func() {
				input := validResource()
				input.Spec.ScaleSettings = &AzureMachineLearningComputeClusterScaleSettings{
					MinNodeCount:                    4,
					MaxNodeCount:                    2,
					ScaleDownNodesAfterIdleDuration: "PT30M",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a negative min node count", func() {
				input := validResource()
				input.Spec.ScaleSettings.MinNodeCount = -1
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a non-ISO-8601 idle duration", func() {
				input := validResource()
				input.Spec.ScaleSettings.ScaleDownNodesAfterIdleDuration = "30 minutes"
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a missing idle duration", func() {
				input := validResource()
				input.Spec.ScaleSettings.ScaleDownNodesAfterIdleDuration = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject an ssh block with neither credential", func() {
				input := validResource()
				input.Spec.Ssh = &AzureMachineLearningComputeClusterSsh{
					AdminUsername: "azureuser",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject an ssh block without an admin username", func() {
				input := validResource()
				input.Spec.Ssh = &AzureMachineLearningComputeClusterSsh{
					KeyValue: "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQ example",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a user-assigned identity without identity ids", func() {
				input := validResource()
				input.Spec.Identity = &AzureMachineLearningComputeClusterIdentity{
					Type: AzureMachineLearningComputeClusterIdentityType_USER_ASSIGNED,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a system-assigned identity carrying identity ids", func() {
				input := validResource()
				input.Spec.Identity = &AzureMachineLearningComputeClusterIdentity{
					Type:        AzureMachineLearningComputeClusterIdentityType_SYSTEM_ASSIGNED,
					IdentityIds: []*foreignkeyv1.StringValueOrRef{literal(testIdentityId)},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})
	})
})
