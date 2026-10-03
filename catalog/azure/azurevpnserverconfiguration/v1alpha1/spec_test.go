package azurevpnserverconfigurationv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestAzureVpnServerConfigurationSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureVpnServerConfigurationSpec Validation Tests")
}

// literal builds a StringValueOrRef carrying a literal value.
func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

// testRootCertData is a placeholder base64 X.509 body -- validation
// checks presence, not certificate semantics (ARM does that).
const testRootCertData = "MIIC5zCCAc+gAwIBAgIQEXAMPLEONLYNOTAREALCERT"

// validResource returns a minimal valid certificate-auth configuration
// that individual cases mutate into the shape under test.
func validResource() *AzureVpnServerConfiguration {
	return &AzureVpnServerConfiguration{
		ApiVersion: "azure.planton.dev/v1alpha1",
		Kind:       "AzureVpnServerConfiguration",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-vpn-server-configuration",
		},
		Spec: &AzureVpnServerConfigurationSpec{
			Region:                 "eastus",
			ResourceGroup:          literal("test-rg"),
			Name:                   "remote-workforce",
			VpnAuthenticationTypes: []string{"Certificate"},
			ClientRootCertificates: []*AzureVpnServerConfigurationClientRootCertificate{
				{Name: "corp-root", PublicCertData: testRootCertData},
			},
		},
	}
}

// validAad returns a complete Entra ID authentication block.
func validAad() *AzureVpnServerConfigurationAadAuthentication {
	return &AzureVpnServerConfigurationAadAuthentication{
		Audience: "41b23e61-6c1e-4545-b367-cd054e0ed4b4",
		Issuer:   "https://sts.windows.net/00000000-0000-0000-0000-000000000000/",
		Tenant:   "https://login.microsoftonline.com/00000000-0000-0000-0000-000000000000",
	}
}

// validRadius returns a single-server RADIUS block.
func validRadius() *AzureVpnServerConfigurationRadius {
	return &AzureVpnServerConfigurationRadius{
		Servers: []*AzureVpnServerConfigurationRadiusServer{
			{Address: "10.50.0.4", Secret: literal("radius-shared-secret"), Score: 10},
		},
	}
}

var _ = ginkgo.Describe("AzureVpnServerConfigurationSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("azure_vpn_server_configuration", func() {

			ginkgo.It("should not return a validation error for a minimal certificate-auth configuration", func() {
				err := protovalidate.Validate(validResource())
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept Entra ID authentication with the complete trio", func() {
				input := validResource()
				input.Spec.VpnAuthenticationTypes = []string{"AAD"}
				input.Spec.ClientRootCertificates = nil
				input.Spec.AadAuthentication = validAad()
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept RADIUS authentication with a server", func() {
				input := validResource()
				input.Spec.VpnAuthenticationTypes = []string{"Radius"}
				input.Spec.ClientRootCertificates = nil
				input.Spec.Radius = validRadius()
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept all three authentication types with all three blocks", func() {
				input := validResource()
				input.Spec.VpnAuthenticationTypes = []string{"AAD", "Certificate", "Radius"}
				input.Spec.AadAuthentication = validAad()
				input.Spec.Radius = validRadius()
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a fully pinned IPsec policy", func() {
				input := validResource()
				input.Spec.IpsecPolicy = &AzureVpnServerConfigurationIpsecPolicy{
					DhGroup:             "DHGroup14",
					IkeEncryption:       "AES256",
					IkeIntegrity:        "SHA256",
					IpsecEncryption:     "AES256",
					IpsecIntegrity:      "SHA256",
					PfsGroup:            "PFS2048",
					SaLifetimeSeconds:   27000,
					SaDataSizeKilobytes: 102400000,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept both tunnel protocols", func() {
				input := validResource()
				input.Spec.VpnProtocols = []string{"IkeV2", "OpenVPN"}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept revoked client certificates", func() {
				input := validResource()
				input.Spec.ClientRevokedCertificates = []*AzureVpnServerConfigurationClientRevokedCertificate{
					{Name: "lost-laptop", Thumbprint: "A1B2C3D4E5F6A7B8C9D0E1F2A3B4C5D6E7F8A9B0"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept policy groups with all three member-rule types", func() {
				input := validResource()
				input.Spec.PolicyGroups = []*AzureVpnServerConfigurationPolicyGroup{
					{
						Name:      "engineering",
						IsDefault: true,
						Priority:  0,
						Policies: []*AzureVpnServerConfigurationPolicyGroupPolicy{
							{Name: "eng-entra", Type: "AADGroupId", Value: "11111111-1111-1111-1111-111111111111"},
						},
					},
					{
						Name:     "contractors",
						Priority: 10,
						Policies: []*AzureVpnServerConfigurationPolicyGroupPolicy{
							{Name: "contractor-cn", Type: "CertificateGroupId", Value: "contractor"},
							{Name: "contractor-radius", Type: "RadiusAzureGroupId", Value: "6ad1bd08"},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept RADIUS trust anchors in both directions", func() {
				input := validResource()
				input.Spec.VpnAuthenticationTypes = []string{"Radius"}
				input.Spec.ClientRootCertificates = nil
				radius := validRadius()
				radius.ClientRootCertificates = []*AzureVpnServerConfigurationRadiusClientRootCertificate{
					{Name: "radius-client-root", Thumbprint: "A1B2C3D4E5F6A7B8C9D0E1F2A3B4C5D6E7F8A9B0"},
				}
				radius.ServerRootCertificates = []*AzureVpnServerConfigurationRadiusServerRootCertificate{
					{Name: "radius-server-root", PublicCertData: testRootCertData},
				}
				input.Spec.Radius = radius
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("azure_vpn_server_configuration", func() {

			ginkgo.It("should reject a missing name", func() {
				input := validResource()
				input.Spec.Name = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject empty vpn_authentication_types", func() {
				input := validResource()
				input.Spec.VpnAuthenticationTypes = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject a duplicate authentication type", func() {
				input := validResource()
				input.Spec.VpnAuthenticationTypes = []string{"Certificate", "Certificate"}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject an unknown authentication type", func() {
				input := validResource()
				input.Spec.VpnAuthenticationTypes = []string{"Password"}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject the AAD type without the aad_authentication block", func() {
				input := validResource()
				input.Spec.VpnAuthenticationTypes = []string{"AAD"}
				input.Spec.ClientRootCertificates = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject the Certificate type without a root certificate", func() {
				input := validResource()
				input.Spec.ClientRootCertificates = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject the Radius type without the radius block", func() {
				input := validResource()
				input.Spec.VpnAuthenticationTypes = []string{"Radius"}
				input.Spec.ClientRootCertificates = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject an AAD block missing its tenant", func() {
				input := validResource()
				input.Spec.VpnAuthenticationTypes = []string{"AAD"}
				input.Spec.ClientRootCertificates = nil
				aad := validAad()
				aad.Tenant = ""
				input.Spec.AadAuthentication = aad
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject a root certificate without its public data", func() {
				input := validResource()
				input.Spec.ClientRootCertificates = []*AzureVpnServerConfigurationClientRootCertificate{
					{Name: "corp-root"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject duplicate root certificate names", func() {
				input := validResource()
				input.Spec.ClientRootCertificates = []*AzureVpnServerConfigurationClientRootCertificate{
					{Name: "corp-root", PublicCertData: testRootCertData},
					{Name: "corp-root", PublicCertData: testRootCertData},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject a RADIUS server score outside 1-30", func() {
				input := validResource()
				input.Spec.VpnAuthenticationTypes = []string{"Radius"}
				input.Spec.ClientRootCertificates = nil
				radius := validRadius()
				radius.Servers[0].Score = 31
				input.Spec.Radius = radius
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject a RADIUS server without a secret", func() {
				input := validResource()
				input.Spec.VpnAuthenticationTypes = []string{"Radius"}
				input.Spec.ClientRootCertificates = nil
				radius := validRadius()
				radius.Servers[0].Secret = nil
				input.Spec.Radius = radius
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject an unknown DH group in the IPsec policy", func() {
				input := validResource()
				input.Spec.IpsecPolicy = &AzureVpnServerConfigurationIpsecPolicy{
					DhGroup:             "DHGroup99",
					IkeEncryption:       "AES256",
					IkeIntegrity:        "SHA256",
					IpsecEncryption:     "AES256",
					IpsecIntegrity:      "SHA256",
					PfsGroup:            "PFS2048",
					SaLifetimeSeconds:   27000,
					SaDataSizeKilobytes: 102400000,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject an unknown tunnel protocol", func() {
				input := validResource()
				input.Spec.VpnProtocols = []string{"SSTP"}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject a policy group without member rules", func() {
				input := validResource()
				input.Spec.PolicyGroups = []*AzureVpnServerConfigurationPolicyGroup{
					{Name: "engineering"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject an unknown policy member-rule type", func() {
				input := validResource()
				input.Spec.PolicyGroups = []*AzureVpnServerConfigurationPolicyGroup{
					{
						Name: "engineering",
						Policies: []*AzureVpnServerConfigurationPolicyGroupPolicy{
							{Name: "rule", Type: "IPAddress", Value: "10.0.0.0/8"},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject duplicate policy group names", func() {
				input := validResource()
				group := func() *AzureVpnServerConfigurationPolicyGroup {
					return &AzureVpnServerConfigurationPolicyGroup{
						Name: "engineering",
						Policies: []*AzureVpnServerConfigurationPolicyGroupPolicy{
							{Name: "rule", Type: "AADGroupId", Value: "11111111-1111-1111-1111-111111111111"},
						},
					}
				}
				input.Spec.PolicyGroups = []*AzureVpnServerConfigurationPolicyGroup{group(), group()}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject duplicate policy names within a group", func() {
				input := validResource()
				input.Spec.PolicyGroups = []*AzureVpnServerConfigurationPolicyGroup{
					{
						Name: "engineering",
						Policies: []*AzureVpnServerConfigurationPolicyGroupPolicy{
							{Name: "rule", Type: "AADGroupId", Value: "11111111-1111-1111-1111-111111111111"},
							{Name: "rule", Type: "CertificateGroupId", Value: "engineering"},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})
	})
})
