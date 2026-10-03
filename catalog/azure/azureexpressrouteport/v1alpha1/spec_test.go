package azureexpressrouteportv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestAzureExpressRoutePortSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureExpressRoutePortSpec Validation Tests")
}

// literal builds a StringValueOrRef carrying a literal value.
func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

// billingType returns a pointer to the given billing-type enum value
// (the field is optional, so the generated Go type is a pointer).
func billingType(value AzureExpressRoutePortBillingType) *AzureExpressRoutePortBillingType {
	return &value
}

// macsecCipher returns a pointer to the given cipher enum value.
func macsecCipher(value AzureExpressRoutePortMacsecCipher) *AzureExpressRoutePortMacsecCipher {
	return &value
}

// validResource returns a minimal valid 10 Gbps Dot1Q port that
// individual cases mutate into the shape under test.
func validResource() *AzureExpressRoutePort {
	return &AzureExpressRoutePort{
		ApiVersion: "azure.planton.dev/v1alpha1",
		Kind:       "AzureExpressRoutePort",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-erport",
		},
		Spec: &AzureExpressRoutePortSpec{
			Region:          "eastus",
			ResourceGroup:   literal("test-rg"),
			Name:            "hq-port",
			PeeringLocation: "Equinix-Ashburn-DC2",
			BandwidthInGbps: 10,
			Encapsulation:   AzureExpressRoutePortEncapsulation_DOT1Q,
		},
	}
}

var _ = ginkgo.Describe("AzureExpressRoutePortSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("azure_express_route_port", func() {

			ginkgo.It("should not return a validation error for a minimal Dot1Q port", func() {
				err := protovalidate.Validate(validResource())
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a 100 Gbps QinQ unlimited port", func() {
				input := validResource()
				input.Spec.BandwidthInGbps = 100
				input.Spec.Encapsulation = AzureExpressRoutePortEncapsulation_QINQ
				input.Spec.BillingType = billingType(AzureExpressRoutePortBillingType_UNLIMITED_DATA)
				input.Spec.Tags = map[string]string{"cost-center": "networking"}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a system-assigned identity", func() {
				input := validResource()
				input.Spec.Identity = &AzureExpressRoutePortIdentity{
					Type: AzureExpressRoutePortIdentityType_SYSTEM_ASSIGNED,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept MACsec links backed by a user-assigned identity", func() {
				input := validResource()
				input.Spec.Identity = &AzureExpressRoutePortIdentity{
					Type:        AzureExpressRoutePortIdentityType_USER_ASSIGNED,
					IdentityIds: []*foreignkeyv1.StringValueOrRef{literal("/subscriptions/sub/resourceGroups/rg/providers/Microsoft.ManagedIdentity/userAssignedIdentities/macsec-identity")},
				}
				input.Spec.Link1 = &AzureExpressRoutePortLink{
					AdminEnabled:              true,
					MacsecCipher:              macsecCipher(AzureExpressRoutePortMacsecCipher_GCM_AES_256),
					MacsecCknKeyvaultSecretId: "https://kv.vault.azure.net/secrets/ckn/v1",
					MacsecCakKeyvaultSecretId: "https://kv.vault.azure.net/secrets/cak/v1",
					MacsecSciEnabled:          true,
				}
				input.Spec.Link2 = &AzureExpressRoutePortLink{
					AdminEnabled:              true,
					MacsecCknKeyvaultSecretId: "https://kv.vault.azure.net/secrets/ckn/v1",
					MacsecCakKeyvaultSecretId: "https://kv.vault.azure.net/secrets/cak/v1",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept enabled links without MACsec and no identity", func() {
				input := validResource()
				input.Spec.Link1 = &AzureExpressRoutePortLink{AdminEnabled: true}
				input.Spec.Link2 = &AzureExpressRoutePortLink{AdminEnabled: true}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept issued authorizations", func() {
				input := validResource()
				input.Spec.Authorizations = []*AzureExpressRoutePortAuthorization{
					{Name: "partner-team"},
					{Name: "dr-site"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("azure_express_route_port", func() {

			ginkgo.It("should return a validation error when region is missing", func() {
				input := validResource()
				input.Spec.Region = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when resource_group is missing", func() {
				input := validResource()
				input.Spec.ResourceGroup = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when name is missing", func() {
				input := validResource()
				input.Spec.Name = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a name starting with an underscore", func() {
				input := validResource()
				input.Spec.Name = "_hq-port"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a name ending in a hyphen", func() {
				input := validResource()
				input.Spec.Name = "hq-port-"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a name over 80 characters", func() {
				input := validResource()
				name := "p"
				for len(name) < 81 {
					name += "x"
				}
				input.Spec.Name = name
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when peering_location is missing", func() {
				input := validResource()
				input.Spec.PeeringLocation = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when bandwidth is missing", func() {
				input := validResource()
				input.Spec.BandwidthInGbps = 0
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when encapsulation is unspecified", func() {
				input := validResource()
				input.Spec.Encapsulation = AzureExpressRoutePortEncapsulation_azure_express_route_port_encapsulation_unspecified
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for an undefined billing type", func() {
				input := validResource()
				input.Spec.BillingType = billingType(AzureExpressRoutePortBillingType(99))
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for MACsec with only the CKN key", func() {
				input := validResource()
				input.Spec.Identity = &AzureExpressRoutePortIdentity{
					Type:        AzureExpressRoutePortIdentityType_USER_ASSIGNED,
					IdentityIds: []*foreignkeyv1.StringValueOrRef{literal("/subscriptions/sub/resourceGroups/rg/providers/Microsoft.ManagedIdentity/userAssignedIdentities/macsec-identity")},
				}
				input.Spec.Link1 = &AzureExpressRoutePortLink{
					MacsecCknKeyvaultSecretId: "https://kv.vault.azure.net/secrets/ckn/v1",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for MACsec without a user-assigned identity", func() {
				input := validResource()
				input.Spec.Link1 = &AzureExpressRoutePortLink{
					MacsecCknKeyvaultSecretId: "https://kv.vault.azure.net/secrets/ckn/v1",
					MacsecCakKeyvaultSecretId: "https://kv.vault.azure.net/secrets/cak/v1",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for MACsec with only a system-assigned identity", func() {
				input := validResource()
				input.Spec.Identity = &AzureExpressRoutePortIdentity{
					Type: AzureExpressRoutePortIdentityType_SYSTEM_ASSIGNED,
				}
				input.Spec.Link2 = &AzureExpressRoutePortLink{
					MacsecCknKeyvaultSecretId: "https://kv.vault.azure.net/secrets/ckn/v1",
					MacsecCakKeyvaultSecretId: "https://kv.vault.azure.net/secrets/cak/v1",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a user-assigned identity without identity_ids", func() {
				input := validResource()
				input.Spec.Identity = &AzureExpressRoutePortIdentity{
					Type: AzureExpressRoutePortIdentityType_USER_ASSIGNED,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a system-assigned identity carrying identity_ids", func() {
				input := validResource()
				input.Spec.Identity = &AzureExpressRoutePortIdentity{
					Type:        AzureExpressRoutePortIdentityType_SYSTEM_ASSIGNED,
					IdentityIds: []*foreignkeyv1.StringValueOrRef{literal("/subscriptions/sub/resourceGroups/rg/providers/Microsoft.ManagedIdentity/userAssignedIdentities/extra")},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for duplicate authorization names", func() {
				input := validResource()
				input.Spec.Authorizations = []*AzureExpressRoutePortAuthorization{
					{Name: "partner-team"},
					{Name: "partner-team"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for an authorization without a name", func() {
				input := validResource()
				input.Spec.Authorizations = []*AzureExpressRoutePortAuthorization{
					{Name: ""},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})
	})
})
