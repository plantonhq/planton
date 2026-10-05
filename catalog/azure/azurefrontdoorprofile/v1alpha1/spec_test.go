package azurefrontdoorprofilev1alpha1

import (
	"strings"
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
	"google.golang.org/protobuf/proto"
)

func TestAzureFrontDoorProfileSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureFrontDoorProfileSpec Validation Tests")
}

func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

const identityId = "/subscriptions/s/resourceGroups/rg/providers/Microsoft.ManagedIdentity/userAssignedIdentities/cert-reader"

// minimal valid spec: a Standard profile (sku unspecified deploys
// STANDARD).
func minimalSpec() *AzureFrontDoorProfile {
	return &AzureFrontDoorProfile{
		ApiVersion: "azure.planton.dev/v1alpha1",
		Kind:       "AzureFrontDoorProfile",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-front-door-profile",
		},
		Spec: &AzureFrontDoorProfileSpec{
			ResourceGroup: literal("test-rg"),
			ProfileName:   "test-profile",
		},
	}
}

var _ = ginkgo.Describe("AzureFrontDoorProfileSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {

		ginkgo.It("should accept a minimal Standard profile", func() {
			gomega.Expect(protovalidate.Validate(minimalSpec())).To(gomega.BeNil())
		})

		ginkgo.It("should accept both explicit sku tiers", func() {
			for _, sku := range []AzureFrontDoorProfileSku{
				AzureFrontDoorProfileSku_STANDARD,
				AzureFrontDoorProfileSku_PREMIUM,
			} {
				input := minimalSpec()
				input.Spec.Sku = sku
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil(), "sku %v must be accepted", sku)
			}
		})

		ginkgo.It("should accept a resource-group reference via valueFrom", func() {
			input := minimalSpec()
			input.Spec.ResourceGroup = &foreignkeyv1.StringValueOrRef{
				LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
					ValueFrom: &foreignkeyv1.ValueFromRef{Name: "my-rg"},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept response timeout boundaries 16, 120, and 240", func() {
			for _, timeout := range []int32{16, 120, 240} {
				input := minimalSpec()
				input.Spec.ResponseTimeoutSeconds = proto.Int32(timeout)
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil(), "timeout %d must be accepted", timeout)
			}
		})

		ginkgo.It("should accept profile name boundaries (2 and 90 characters)", func() {
			input := minimalSpec()
			input.Spec.ProfileName = "ab"
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			input.Spec.ProfileName = "a" + strings.Repeat("b", 88) + "c"
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept a system-assigned identity", func() {
			input := minimalSpec()
			input.Spec.Identity = &AzureFrontDoorProfileIdentity{
				Type: AzureFrontDoorProfileIdentityType_SYSTEM_ASSIGNED,
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept a user-assigned identity with identity ids", func() {
			input := minimalSpec()
			input.Spec.Identity = &AzureFrontDoorProfileIdentity{
				Type:                    AzureFrontDoorProfileIdentityType_USER_ASSIGNED,
				UserAssignedIdentityIds: []*foreignkeyv1.StringValueOrRef{literal(identityId)},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept a combined identity with identity ids", func() {
			input := minimalSpec()
			input.Spec.Identity = &AzureFrontDoorProfileIdentity{
				Type:                    AzureFrontDoorProfileIdentityType_SYSTEM_AND_USER_ASSIGNED,
				UserAssignedIdentityIds: []*foreignkeyv1.StringValueOrRef{literal(identityId)},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept all three log-scrubbing variables together", func() {
			input := minimalSpec()
			input.Spec.LogScrubbingVariables = []AzureFrontDoorProfileLogScrubbingVariable{
				AzureFrontDoorProfileLogScrubbingVariable_QUERY_STRING_ARG_NAMES,
				AzureFrontDoorProfileLogScrubbingVariable_REQUEST_IP_ADDRESS,
				AzureFrontDoorProfileLogScrubbingVariable_REQUEST_URI,
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept user tags", func() {
			input := minimalSpec()
			input.Spec.Tags = map[string]string{"cost-center": "platform", "owner": "web-team"}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {

		ginkgo.It("should reject a missing resource group", func() {
			input := minimalSpec()
			input.Spec.ResourceGroup = nil
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a missing profile name", func() {
			input := minimalSpec()
			input.Spec.ProfileName = ""
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a single-character profile name", func() {
			input := minimalSpec()
			input.Spec.ProfileName = "a"
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a profile name over 90 characters", func() {
			input := minimalSpec()
			input.Spec.ProfileName = "a" + strings.Repeat("b", 89) + "c"
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a profile name with a leading hyphen", func() {
			input := minimalSpec()
			input.Spec.ProfileName = "-profile"
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a profile name with a trailing hyphen", func() {
			input := minimalSpec()
			input.Spec.ProfileName = "profile-"
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a profile name with invalid characters", func() {
			input := minimalSpec()
			input.Spec.ProfileName = "my_profile"
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject an undefined sku enum number", func() {
			input := minimalSpec()
			input.Spec.Sku = AzureFrontDoorProfileSku(99)
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject response timeout below 16 seconds", func() {
			input := minimalSpec()
			input.Spec.ResponseTimeoutSeconds = proto.Int32(15)
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject response timeout above 240 seconds", func() {
			input := minimalSpec()
			input.Spec.ResponseTimeoutSeconds = proto.Int32(241)
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject an identity without a type", func() {
			input := minimalSpec()
			input.Spec.Identity = &AzureFrontDoorProfileIdentity{}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a user-assigned identity without identity ids", func() {
			input := minimalSpec()
			input.Spec.Identity = &AzureFrontDoorProfileIdentity{
				Type: AzureFrontDoorProfileIdentityType_USER_ASSIGNED,
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a system-assigned identity carrying identity ids", func() {
			input := minimalSpec()
			input.Spec.Identity = &AzureFrontDoorProfileIdentity{
				Type:                    AzureFrontDoorProfileIdentityType_SYSTEM_ASSIGNED,
				UserAssignedIdentityIds: []*foreignkeyv1.StringValueOrRef{literal(identityId)},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject an unspecified log-scrubbing variable", func() {
			input := minimalSpec()
			input.Spec.LogScrubbingVariables = []AzureFrontDoorProfileLogScrubbingVariable{
				AzureFrontDoorProfileLogScrubbingVariable_azure_front_door_profile_log_scrubbing_variable_unspecified,
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject duplicate log-scrubbing variables", func() {
			input := minimalSpec()
			input.Spec.LogScrubbingVariables = []AzureFrontDoorProfileLogScrubbingVariable{
				AzureFrontDoorProfileLogScrubbingVariable_REQUEST_URI,
				AzureFrontDoorProfileLogScrubbingVariable_REQUEST_URI,
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a wrong api version", func() {
			input := minimalSpec()
			input.ApiVersion = "azure.planton.dev/v2"
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a wrong kind", func() {
			input := minimalSpec()
			input.Kind = "AzureFrontDoor"
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject missing metadata", func() {
			input := minimalSpec()
			input.Metadata = nil
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a missing spec", func() {
			input := minimalSpec()
			input.Spec = nil
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})
	})
})
