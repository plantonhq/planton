package azuremachinelearningonlineendpointv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestAzureMachineLearningOnlineEndpointSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureMachineLearningOnlineEndpointSpec Validation Tests")
}

// literal builds a StringValueOrRef carrying a literal value.
func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

const (
	testWorkspaceId = "/subscriptions/s/resourceGroups/rg/providers/Microsoft.MachineLearningServices/workspaces/ml-workspace"
	testIdentityId  = "/subscriptions/s/resourceGroups/rg/providers/Microsoft.ManagedIdentity/userAssignedIdentities/ml-uai"
)

// validResource returns a minimal valid key-auth endpoint that
// individual cases mutate into the shape under test.
func validResource() *AzureMachineLearningOnlineEndpoint {
	return &AzureMachineLearningOnlineEndpoint{
		ApiVersion: "azure.planton.dev/v1alpha1",
		Kind:       "AzureMachineLearningOnlineEndpoint",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-ml-online-endpoint",
		},
		Spec: &AzureMachineLearningOnlineEndpointSpec{
			WorkspaceId: literal(testWorkspaceId),
			Name:        "fraud-scoring",
			Region:      "eastus",
			AuthMode:    "Key",
			Identity: &AzureMachineLearningOnlineEndpointIdentity{
				Type: AzureMachineLearningOnlineEndpointIdentityType_SYSTEM_ASSIGNED,
			},
		},
	}
}

var _ = ginkgo.Describe("AzureMachineLearningOnlineEndpointSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("azure_machine_learning_online_endpoint", func() {

			ginkgo.It("should not return a validation error for a minimal key-auth endpoint", func() {
				err := protovalidate.Validate(validResource())
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept every service auth mode", func() {
				for _, mode := range []string{"Key", "AMLToken", "AADToken"} {
					input := validResource()
					input.Spec.AuthMode = mode
					err := protovalidate.Validate(input)
					gomega.Expect(err).To(gomega.BeNil())
				}
			})

			ginkgo.It("should accept a blue/green traffic split with mirrored shadow traffic", func() {
				input := validResource()
				input.Spec.Traffic = map[string]int32{"blue": 90, "green": 10}
				input.Spec.MirrorTraffic = map[string]int32{"shadow": 25}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a user-assigned identity with identity ids", func() {
				input := validResource()
				input.Spec.Identity = &AzureMachineLearningOnlineEndpointIdentity{
					Type:        AzureMachineLearningOnlineEndpointIdentityType_USER_ASSIGNED,
					IdentityIds: []*foreignkeyv1.StringValueOrRef{literal(testIdentityId)},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept bring-your-own initial keys with only a primary key", func() {
				input := validResource()
				input.Spec.InitialAuthKeys = &AzureMachineLearningOnlineEndpointInitialAuthKeys{
					PrimaryKey: literal("primary-key-from-vault"),
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept names on ARM's rule: digit start, underscores and hyphens", func() {
				for _, name := range []string{"0endpoint", "fraud_scoring-v2", "a"} {
					input := validResource()
					input.Spec.Name = name
					err := protovalidate.Validate(input)
					gomega.Expect(err).To(gomega.BeNil())
				}
			})

			ginkgo.It("should accept a private endpoint with public network access disabled", func() {
				input := validResource()
				disabled := false
				input.Spec.PublicNetworkAccessEnabled = &disabled
				input.Spec.Description = "private scoring endpoint"
				input.Spec.Properties = map[string]string{"team": "ml-platform"}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("azure_machine_learning_online_endpoint", func() {

			ginkgo.It("should reject a missing workspace reference", func() {
				input := validResource()
				input.Spec.WorkspaceId = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject an auth mode outside the service vocabulary", func() {
				input := validResource()
				input.Spec.AuthMode = "Token"
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a missing identity block (the recorded tightening)", func() {
				input := validResource()
				input.Spec.Identity = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a user-assigned identity without identity ids", func() {
				input := validResource()
				input.Spec.Identity = &AzureMachineLearningOnlineEndpointIdentity{
					Type: AzureMachineLearningOnlineEndpointIdentityType_USER_ASSIGNED,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a system-assigned identity carrying identity ids", func() {
				input := validResource()
				input.Spec.Identity = &AzureMachineLearningOnlineEndpointIdentity{
					Type:        AzureMachineLearningOnlineEndpointIdentityType_SYSTEM_ASSIGNED,
					IdentityIds: []*foreignkeyv1.StringValueOrRef{literal(testIdentityId)},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a traffic percentage above 100", func() {
				input := validResource()
				input.Spec.Traffic = map[string]int32{"blue": 101}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a negative traffic percentage", func() {
				input := validResource()
				input.Spec.Traffic = map[string]int32{"blue": -1}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a mirror-traffic percentage above the service's 50 cap", func() {
				input := validResource()
				input.Spec.MirrorTraffic = map[string]int32{"shadow": 51}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject names ARM rejects", func() {
				for _, name := range []string{"-endpoint", "_endpoint", "fraud scoring", "fraud.scoring"} {
					input := validResource()
					input.Spec.Name = name
					err := protovalidate.Validate(input)
					gomega.Expect(err).NotTo(gomega.BeNil())
				}
			})

			ginkgo.It("should reject an initial-keys block carrying no key", func() {
				input := validResource()
				input.Spec.InitialAuthKeys = &AzureMachineLearningOnlineEndpointInitialAuthKeys{}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a missing region", func() {
				input := validResource()
				input.Spec.Region = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a wrong kind literal", func() {
				input := validResource()
				input.Kind = "AzureMachineLearningEndpoint"
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})
	})
})
