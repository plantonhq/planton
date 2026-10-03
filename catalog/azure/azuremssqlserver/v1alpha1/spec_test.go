package azuremssqlserverv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	"github.com/plantonhq/planton/shared/catalogkind"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestAzureMssqlServerSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureMssqlServerSpec Validation Tests")
}

func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

const uaiId = "/subscriptions/s/resourceGroups/rg/providers/Microsoft.ManagedIdentity/userAssignedIdentities/sql-uai"

// minimal valid spec: a SQL-auth server on the public endpoint.
func minimalSpec() *AzureMssqlServer {
	return &AzureMssqlServer{
		ApiVersion: "azure.planton.dev/v1alpha1",
		Kind:       "AzureMssqlServer",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-mssql",
		},
		Spec: &AzureMssqlServerSpec{
			Region:                "eastus",
			ResourceGroup:         literal("my-rg"),
			ServerName:            "test-mssql-server",
			AdministratorLogin:    "sqladmin",
			AdministratorPassword: literal("P@ssw0rd1234!"),
		},
	}
}

func aadAdmin(aadOnly bool) *AzureMssqlServerAzureadAdministrator {
	return &AzureMssqlServerAzureadAdministrator{
		LoginUsername:             "dba-team@contoso.com",
		ObjectId:                  literal("11111111-2222-3333-4444-555555555555"),
		AzureadAuthenticationOnly: aadOnly,
	}
}

var _ = ginkgo.Describe("AzureMssqlServerSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {

		ginkgo.It("should accept a minimal SQL-auth server", func() {
			gomega.Expect(protovalidate.Validate(minimalSpec())).To(gomega.BeNil())
		})

		ginkgo.It("should accept both supported versions", func() {
			for _, v := range []string{"2.0", "12.0"} {
				ver := v
				input := minimalSpec()
				input.Spec.Version = &ver
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			}
		})

		ginkgo.It("should accept an Entra-only server without SQL credentials", func() {
			input := minimalSpec()
			input.Spec.AdministratorLogin = ""
			input.Spec.AdministratorPassword = nil
			input.Spec.AzureadAdministrator = aadAdmin(true)
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept mixed auth: SQL credentials plus an Entra administrator", func() {
			input := minimalSpec()
			input.Spec.AzureadAdministrator = aadAdmin(false)
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept a system-assigned identity", func() {
			input := minimalSpec()
			input.Spec.Identity = &AzureMssqlServerIdentity{
				Type: AzureMssqlServerIdentityType_SYSTEM_ASSIGNED,
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept a user-assigned identity with the primary pinned and TDE CMK", func() {
			input := minimalSpec()
			input.Spec.Identity = &AzureMssqlServerIdentity{
				Type:        AzureMssqlServerIdentityType_USER_ASSIGNED,
				IdentityIds: []*foreignkeyv1.StringValueOrRef{literal(uaiId)},
			}
			input.Spec.PrimaryUserAssignedIdentityId = literal(uaiId)
			input.Spec.TransparentDataEncryptionKeyVaultKeyId = literal("https://vault.vault.azure.net/keys/sql-tde/0123456789abcdef0123456789abcdef")
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept every connection policy", func() {
			for _, policy := range []AzureMssqlServerConnectionPolicy{
				AzureMssqlServerConnectionPolicy_DEFAULT,
				AzureMssqlServerConnectionPolicy_PROXY,
				AzureMssqlServerConnectionPolicy_REDIRECT,
			} {
				input := minimalSpec()
				input.Spec.ConnectionPolicy = policy
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			}
		})

		ginkgo.It("should accept outbound restriction with FQDN rules", func() {
			input := minimalSpec()
			input.Spec.OutboundNetworkRestrictionEnabled = true
			input.Spec.OutboundFirewallRules = []string{"peer.database.windows.net"}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept firewall and virtual network rules", func() {
			input := minimalSpec()
			input.Spec.FirewallRules = []*AzureMssqlServerFirewallRule{
				{Name: "allow-azure", StartIpAddress: "0.0.0.0", EndIpAddress: "0.0.0.0"},
				{Name: "allow-office", StartIpAddress: "203.0.113.0", EndIpAddress: "203.0.113.255"},
			}
			input.Spec.VirtualNetworkRules = []*AzureMssqlServerVirtualNetworkRule{
				{
					Name:     "allow-app-subnet",
					SubnetId: literal("/subscriptions/s/resourceGroups/rg/providers/Microsoft.Network/virtualNetworks/vnet/subnets/app"),
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept extended auditing to blob storage", func() {
			retention := int32(90)
			input := minimalSpec()
			input.Spec.ExtendedAuditing = &AzureMssqlServerExtendedAuditing{
				StorageEndpoint:         "https://auditlogs.blob.core.windows.net",
				StorageAccountAccessKey: literal("key=="),
				RetentionInDays:         &retention,
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept Azure-Monitor-only auditing", func() {
			input := minimalSpec()
			input.Spec.ExtendedAuditing = &AzureMssqlServerExtendedAuditing{}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept a Defender security-alert policy", func() {
			input := minimalSpec()
			input.Spec.SecurityAlertPolicy = &AzureMssqlServerSecurityAlertPolicy{
				State: AzureMssqlServerSecurityAlertPolicyState_ENABLED,
				DisabledAlerts: []AzureMssqlServerSecurityAlertType{
					AzureMssqlServerSecurityAlertType_UNSAFE_ACTION,
				},
				EmailAccountAdmins: true,
				EmailAddresses:     []string{"secops@contoso.com"},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept express vulnerability assessment and tags", func() {
			input := minimalSpec()
			input.Spec.ExpressVulnerabilityAssessmentEnabled = true
			input.Spec.Tags = map[string]string{"team": "data"}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept valueFrom references for resource group", func() {
			input := minimalSpec()
			input.Spec.ResourceGroup = &foreignkeyv1.StringValueOrRef{
				LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
					ValueFrom: &foreignkeyv1.ValueFromRef{
						Kind:      catalogkind.CatalogKind_AzureResourceGroup,
						Name:      "shared-rg",
						FieldPath: "status.outputs.resource_group_name",
					},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {

		ginkgo.It("should reject a missing region", func() {
			input := minimalSpec()
			input.Spec.Region = ""
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject a missing resource group", func() {
			input := minimalSpec()
			input.Spec.ResourceGroup = nil
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject a server name with uppercase or edge hyphens", func() {
			for _, name := range []string{"Invalid-Name", "-bad", "bad-"} {
				input := minimalSpec()
				input.Spec.ServerName = name
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil(), "name %q must be rejected", name)
			}
		})

		ginkgo.It("should reject an unsupported version", func() {
			bad := "11.0"
			input := minimalSpec()
			input.Spec.Version = &bad
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject a reserved administrator login", func() {
			for _, login := range []string{"admin", "SA", "root", "dbmanager", "public"} {
				input := minimalSpec()
				input.Spec.AdministratorLogin = login
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil(), "login %q must be rejected", login)
			}
		})

		ginkgo.It("should reject a server with no authentication mechanism at all", func() {
			input := minimalSpec()
			input.Spec.AdministratorLogin = ""
			input.Spec.AdministratorPassword = nil
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject an Entra administrator without authentication-only when SQL credentials are absent", func() {
			input := minimalSpec()
			input.Spec.AdministratorLogin = ""
			input.Spec.AdministratorPassword = nil
			input.Spec.AzureadAdministrator = aadAdmin(false)
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject SQL credentials on an Entra-only server", func() {
			input := minimalSpec()
			input.Spec.AzureadAdministrator = aadAdmin(true)
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject a login without a password", func() {
			input := minimalSpec()
			input.Spec.AdministratorPassword = nil
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject an Entra administrator with a malformed object id", func() {
			input := minimalSpec()
			admin := aadAdmin(false)
			admin.ObjectId = nil
			input.Spec.AzureadAdministrator = admin
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject an Entra administrator with a malformed tenant", func() {
			tenant := "not-a-uuid"
			input := minimalSpec()
			admin := aadAdmin(false)
			admin.TenantId = &tenant
			input.Spec.AzureadAdministrator = admin
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject a primary user-assigned identity without an identity block", func() {
			input := minimalSpec()
			input.Spec.PrimaryUserAssignedIdentityId = literal(uaiId)
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject a primary user-assigned identity with a SYSTEM_ASSIGNED-only identity", func() {
			input := minimalSpec()
			input.Spec.Identity = &AzureMssqlServerIdentity{
				Type: AzureMssqlServerIdentityType_SYSTEM_ASSIGNED,
			}
			input.Spec.PrimaryUserAssignedIdentityId = literal(uaiId)
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject TDE CMK without an identity", func() {
			input := minimalSpec()
			input.Spec.TransparentDataEncryptionKeyVaultKeyId = literal("https://vault.vault.azure.net/keys/sql-tde/0123456789abcdef0123456789abcdef")
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject a SYSTEM_ASSIGNED identity carrying identity_ids", func() {
			input := minimalSpec()
			input.Spec.Identity = &AzureMssqlServerIdentity{
				Type:        AzureMssqlServerIdentityType_SYSTEM_ASSIGNED,
				IdentityIds: []*foreignkeyv1.StringValueOrRef{literal(uaiId)},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject a USER_ASSIGNED identity without identity_ids", func() {
			input := minimalSpec()
			input.Spec.Identity = &AzureMssqlServerIdentity{
				Type: AzureMssqlServerIdentityType_USER_ASSIGNED,
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject a TLS floor other than 1.2", func() {
			bad := "1.0"
			input := minimalSpec()
			input.Spec.MinimumTlsVersion = &bad
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject outbound FQDN rules without the restriction toggle", func() {
			input := minimalSpec()
			input.Spec.OutboundFirewallRules = []string{"peer.database.windows.net"}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject a firewall rule with a malformed IP", func() {
			input := minimalSpec()
			input.Spec.FirewallRules = []*AzureMssqlServerFirewallRule{
				{Name: "bad", StartIpAddress: "not-an-ip", EndIpAddress: "0.0.0.0"},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject a virtual network rule without a subnet", func() {
			input := minimalSpec()
			input.Spec.VirtualNetworkRules = []*AzureMssqlServerVirtualNetworkRule{
				{Name: "bad"},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject an auditing access key without a storage endpoint", func() {
			input := minimalSpec()
			input.Spec.ExtendedAuditing = &AzureMssqlServerExtendedAuditing{
				StorageAccountAccessKey: literal("key=="),
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject an auditing endpoint that is not https", func() {
			input := minimalSpec()
			input.Spec.ExtendedAuditing = &AzureMssqlServerExtendedAuditing{
				StorageEndpoint: "http://auditlogs.blob.core.windows.net",
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject auditing retention beyond 3285 days", func() {
			retention := int32(4000)
			input := minimalSpec()
			input.Spec.ExtendedAuditing = &AzureMssqlServerExtendedAuditing{
				RetentionInDays: &retention,
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject a security-alert policy without a state", func() {
			input := minimalSpec()
			input.Spec.SecurityAlertPolicy = &AzureMssqlServerSecurityAlertPolicy{}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject a security-alert storage endpoint without its access key", func() {
			input := minimalSpec()
			input.Spec.SecurityAlertPolicy = &AzureMssqlServerSecurityAlertPolicy{
				State:           AzureMssqlServerSecurityAlertPolicyState_ENABLED,
				StorageEndpoint: "https://export.blob.core.windows.net",
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})
	})
})
