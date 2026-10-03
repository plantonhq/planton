package azureeventhubdisasterrecoveryconfigv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	"github.com/plantonhq/planton/shared/catalogkind"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestAzureEventHubDisasterRecoveryConfigSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureEventHubDisasterRecoveryConfigSpec Validation Tests")
}

func literal(v string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: v},
	}
}

func minimalPairing() *AzureEventHubDisasterRecoveryConfig {
	return &AzureEventHubDisasterRecoveryConfig{
		ApiVersion: "azure.planton.dev/v1alpha1",
		Kind:       "AzureEventHubDisasterRecoveryConfig",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-geo-dr",
		},
		Spec: &AzureEventHubDisasterRecoveryConfigSpec{
			AliasName:          "myapp-hub-alias",
			PrimaryNamespaceId: literal("/subscriptions/s/resourceGroups/rg/providers/Microsoft.EventHub/namespaces/myapp-ehns-eastus"),
			PartnerNamespaceId: literal("/subscriptions/s/resourceGroups/rg/providers/Microsoft.EventHub/namespaces/myapp-ehns-westus"),
		},
	}
}

var _ = ginkgo.Describe("AzureEventHubDisasterRecoveryConfigSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("azure_event_hub_disaster_recovery_config", func() {

			ginkgo.It("should accept a minimal pairing", func() {
				gomega.Expect(protovalidate.Validate(minimalPairing())).To(gomega.BeNil())
			})

			ginkgo.It("should accept a single-character alias name", func() {
				input := minimalPairing()
				input.Spec.AliasName = "a"
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept namespace references by valueFrom", func() {
				input := minimalPairing()
				input.Spec.PrimaryNamespaceId = &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
						ValueFrom: &foreignkeyv1.ValueFromRef{
							Kind:      catalogkind.CatalogKind_AzureEventHubNamespace,
							Name:      "primary-ehns",
							FieldPath: "status.outputs.namespace_id",
						},
					},
				}
				input.Spec.PartnerNamespaceId = &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
						ValueFrom: &foreignkeyv1.ValueFromRef{
							Kind:      catalogkind.CatalogKind_AzureEventHubNamespace,
							Name:      "partner-ehns",
							FieldPath: "status.outputs.namespace_id",
						},
					},
				}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("azure_event_hub_disaster_recovery_config", func() {

			ginkgo.It("should reject a missing alias name", func() {
				input := minimalPairing()
				input.Spec.AliasName = ""
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject an alias name over 60 characters", func() {
				input := minimalPairing()
				input.Spec.AliasName = "a-very-long-event-hub-disaster-recovery-alias-name-over-caps"
				input.Spec.AliasName += "x"
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject an alias name with invalid characters", func() {
				input := minimalPairing()
				input.Spec.AliasName = "bad alias!"
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject a missing primary namespace", func() {
				input := minimalPairing()
				input.Spec.PrimaryNamespaceId = nil
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject a missing partner namespace", func() {
				input := minimalPairing()
				input.Spec.PartnerNamespaceId = nil
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject a missing metadata block", func() {
				input := minimalPairing()
				input.Metadata = nil
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject an incorrect kind", func() {
				input := minimalPairing()
				input.Kind = "WrongKind"
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})
		})
	})
})
