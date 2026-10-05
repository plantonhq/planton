package azuremachinelearningcomputeinstancev1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestAzureMachineLearningComputeInstanceSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureMachineLearningComputeInstanceSpec Validation Tests")
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

// validResource returns a minimal valid instance that individual cases
// mutate into the shape under test.
func validResource() *AzureMachineLearningComputeInstance {
	return &AzureMachineLearningComputeInstance{
		ApiVersion: "azure.planton.dev/v1alpha1",
		Kind:       "AzureMachineLearningComputeInstance",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-ml-compute-instance",
		},
		Spec: &AzureMachineLearningComputeInstanceSpec{
			WorkspaceId:        literal(testWorkspaceId),
			Name:               "alice-dev",
			VirtualMachineSize: "STANDARD_DS3_V2",
		},
	}
}

var _ = ginkgo.Describe("AzureMachineLearningComputeInstanceSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("azure_machine_learning_compute_instance", func() {

			ginkgo.It("should not return a validation error for a minimal instance", func() {
				err := protovalidate.Validate(validResource())
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a personal instance assigned to another user", func() {
				input := validResource()
				input.Spec.AuthorizationType = "personal"
				input.Spec.AssignToUser = &AzureMachineLearningComputeInstanceAssignToUser{
					TenantId: "d67d43c0-6b17-4d4a-8e5e-3b8a76f1f1a1",
					ObjectId: "b6870b46-2b6d-4b0e-8bbd-7a4d81f1c1b2",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a system-assigned identity", func() {
				input := validResource()
				input.Spec.Identity = &AzureMachineLearningComputeInstanceIdentity{
					Type: AzureMachineLearningComputeInstanceIdentityType_SYSTEM_ASSIGNED,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a user-assigned identity with identity ids", func() {
				input := validResource()
				input.Spec.Identity = &AzureMachineLearningComputeInstanceIdentity{
					Type:        AzureMachineLearningComputeInstanceIdentityType_USER_ASSIGNED,
					IdentityIds: []*foreignkeyv1.StringValueOrRef{literal(testIdentityId)},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept SSH access with a public key", func() {
				input := validResource()
				input.Spec.Ssh = &AzureMachineLearningComputeInstanceSsh{
					PublicKey: "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQ example",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a private instance with a subnet and no public IP", func() {
				input := validResource()
				input.Spec.SubnetId = literal(testSubnetId)
				nodePublicIp := false
				input.Spec.NodePublicIpEnabled = &nodePublicIp
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a 25-character name at the upper bound", func() {
				input := validResource()
				input.Spec.Name = "a234567890123456789012345"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("azure_machine_learning_compute_instance", func() {

			ginkgo.It("should reject a missing workspace reference", func() {
				input := validResource()
				input.Spec.WorkspaceId = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a 3-character name below the lower bound", func() {
				input := validResource()
				input.Spec.Name = "abc"
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a 26-character name above the upper bound", func() {
				input := validResource()
				input.Spec.Name = "a2345678901234567890123456"
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a name starting with a digit", func() {
				input := validResource()
				input.Spec.Name = "1alice-dev"
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a name with an underscore", func() {
				input := validResource()
				input.Spec.Name = "alice_dev"
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a missing vm size", func() {
				input := validResource()
				input.Spec.VirtualMachineSize = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject an authorization type other than personal", func() {
				input := validResource()
				input.Spec.AuthorizationType = "shared"
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a non-UUID tenant id in assign_to_user", func() {
				input := validResource()
				input.Spec.AssignToUser = &AzureMachineLearningComputeInstanceAssignToUser{
					TenantId: "not-a-uuid",
					ObjectId: "b6870b46-2b6d-4b0e-8bbd-7a4d81f1c1b2",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject an ssh block without a public key", func() {
				input := validResource()
				input.Spec.Ssh = &AzureMachineLearningComputeInstanceSsh{}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a user-assigned identity without identity ids", func() {
				input := validResource()
				input.Spec.Identity = &AzureMachineLearningComputeInstanceIdentity{
					Type: AzureMachineLearningComputeInstanceIdentityType_USER_ASSIGNED,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a system-assigned identity carrying identity ids", func() {
				input := validResource()
				input.Spec.Identity = &AzureMachineLearningComputeInstanceIdentity{
					Type:        AzureMachineLearningComputeInstanceIdentityType_SYSTEM_ASSIGNED,
					IdentityIds: []*foreignkeyv1.StringValueOrRef{literal(testIdentityId)},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})
	})
})
