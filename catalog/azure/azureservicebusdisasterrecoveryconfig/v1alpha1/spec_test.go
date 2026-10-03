package azureservicebusdisasterrecoveryconfigv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	"github.com/plantonhq/planton/shared/catalogkind"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestAzureServiceBusDisasterRecoveryConfigSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureServiceBusDisasterRecoveryConfigSpec Validation Tests")
}

func literal(v string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: v},
	}
}

func minimalPairing() *AzureServiceBusDisasterRecoveryConfig {
	return &AzureServiceBusDisasterRecoveryConfig{
		ApiVersion: "azure.planton.dev/v1alpha1",
		Kind:       "AzureServiceBusDisasterRecoveryConfig",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-geo-dr",
		},
		Spec: &AzureServiceBusDisasterRecoveryConfigSpec{
			AliasName:          "myapp-bus-alias",
			PrimaryNamespaceId: literal("/subscriptions/s/resourceGroups/rg/providers/Microsoft.ServiceBus/namespaces/myapp-bus-eastus"),
			PartnerNamespaceId: literal("/subscriptions/s/resourceGroups/rg/providers/Microsoft.ServiceBus/namespaces/myapp-bus-westus"),
		},
	}
}

var _ = ginkgo.Describe("AzureServiceBusDisasterRecoveryConfigSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("azure_service_bus_disaster_recovery_config", func() {

			ginkgo.It("should accept a minimal pairing", func() {
				gomega.Expect(protovalidate.Validate(minimalPairing())).To(gomega.BeNil())
			})

			ginkgo.It("should accept namespace references by valueFrom", func() {
				input := minimalPairing()
				input.Spec.PrimaryNamespaceId = &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
						ValueFrom: &foreignkeyv1.ValueFromRef{
							Kind:      catalogkind.CatalogKind_AzureServiceBusNamespace,
							Name:      "primary-bus",
							FieldPath: "status.outputs.namespace_id",
						},
					},
				}
				input.Spec.PartnerNamespaceId = &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
						ValueFrom: &foreignkeyv1.ValueFromRef{
							Kind:      catalogkind.CatalogKind_AzureServiceBusNamespace,
							Name:      "partner-bus",
							FieldPath: "status.outputs.namespace_id",
						},
					},
				}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept a scoped alias authorization rule", func() {
				input := minimalPairing()
				input.Spec.AliasAuthorizationRuleId = &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
						ValueFrom: &foreignkeyv1.ValueFromRef{
							Kind:      catalogkind.CatalogKind_AzureServiceBusAuthorizationRule,
							Name:      "dr-clients",
							FieldPath: "status.outputs.authorization_rule_id",
						},
					},
				}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("azure_service_bus_disaster_recovery_config", func() {

			ginkgo.It("should reject a missing alias name", func() {
				input := minimalPairing()
				input.Spec.AliasName = ""
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject an alias name over 50 characters", func() {
				input := minimalPairing()
				input.Spec.AliasName = "a-very-long-disaster-recovery-alias-name-over-caps"
				input.Spec.AliasName += "x"
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject an alias name under 6 characters", func() {
				input := minimalPairing()
				input.Spec.AliasName = "bus1"
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject an alias name starting with a number", func() {
				input := minimalPairing()
				input.Spec.AliasName = "1myapp-bus-alias"
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject an alias name ending with a hyphen", func() {
				input := minimalPairing()
				input.Spec.AliasName = "myapp-bus-alias-"
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject an alias name with characters outside the namespace grammar", func() {
				input := minimalPairing()
				input.Spec.AliasName = "myapp_bus.alias"
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject an alias name ending in a reserved suffix", func() {
				input := minimalPairing()
				input.Spec.AliasName = "myapp-alias-sb"
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
				input.Spec.AliasName = "myapp-alias-mgmt"
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
