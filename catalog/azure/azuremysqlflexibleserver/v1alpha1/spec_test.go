package azuremysqlflexibleserverv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	"github.com/plantonhq/planton/shared/cloudresourcekind"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
	"google.golang.org/protobuf/proto"
)

func TestAzureMysqlFlexibleServerSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureMysqlFlexibleServerSpec Validation Tests")
}

func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

// minimal valid spec: a fresh public-access server with password auth.
func minimalSpec() *AzureMysqlFlexibleServer {
	return &AzureMysqlFlexibleServer{
		ApiVersion: "azure.planton.dev/v1alpha1",
		Kind:       "AzureMysqlFlexibleServer",
		Metadata: &shared.CloudResourceMetadata{
			Name: "test-mysql",
		},
		Spec: &AzureMysqlFlexibleServerSpec{
			Region:                "eastus",
			ResourceGroup:         literal("my-rg"),
			ServerName:            "test-mysql-server",
			AdministratorLogin:    "mysqladmin",
			AdministratorPassword: literal("P@ssw0rd1234!"),
			SkuName:               "GP_Standard_D2ds_v4",
		},
	}
}

const uaiId = "/subscriptions/s/resourceGroups/rg/providers/Microsoft.ManagedIdentity/userAssignedIdentities/mysql-uai"

var _ = ginkgo.Describe("AzureMysqlFlexibleServerSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {

		ginkgo.It("should accept a minimal public server", func() {
			gomega.Expect(protovalidate.Validate(minimalSpec())).To(gomega.BeNil())
		})

		ginkgo.It("should accept every supported MySQL version", func() {
			for _, v := range []string{"5.7", "8.0.21", "8.4"} {
				ver := v
				input := minimalSpec()
				input.Spec.Version = &ver
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			}
		})

		ginkgo.It("should accept each SKU tier prefix", func() {
			for _, sku := range []string{"B_Standard_B1ms", "GP_Standard_D4ads_v5", "MO_Standard_E4ds_v4"} {
				input := minimalSpec()
				input.Spec.SkuName = sku
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			}
		})

		ginkgo.It("should accept a full storage profile with provisioned IOPS", func() {
			size := int32(128)
			iops := int32(1000)
			autoGrow := false
			input := minimalSpec()
			input.Spec.Storage = &AzureMysqlFlexibleServerStorage{
				SizeGb:           &size,
				Iops:             &iops,
				AutoGrowEnabled:  &autoGrow,
				LogOnDiskEnabled: true,
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept elastic IO scaling without provisioned IOPS", func() {
			input := minimalSpec()
			input.Spec.Storage = &AzureMysqlFlexibleServerStorage{
				IoScalingEnabled: true,
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept VNet injection with a private DNS zone", func() {
			input := minimalSpec()
			input.Spec.DelegatedSubnetId = literal("/subscriptions/s/resourceGroups/rg/providers/Microsoft.Network/virtualNetworks/vnet/subnets/mysql")
			input.Spec.PrivateDnsZoneId = literal("/subscriptions/s/resourceGroups/rg/providers/Microsoft.Network/privateDnsZones/mysql.database.azure.com")
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept zone-redundant high availability with pinned zones", func() {
			input := minimalSpec()
			input.Spec.Zone = "1"
			input.Spec.HighAvailability = &AzureMysqlFlexibleServerHighAvailability{
				Mode:                    AzureMysqlFlexibleServerHighAvailabilityMode_ZONE_REDUNDANT,
				StandbyAvailabilityZone: "2",
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept a maintenance window", func() {
			input := minimalSpec()
			input.Spec.MaintenanceWindow = &AzureMysqlFlexibleServerMaintenanceWindow{
				DayOfWeek:   proto.Int32(6),
				StartHour:   proto.Int32(2),
				StartMinute: proto.Int32(30),
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept a replica that inherits sku and credentials from its source", func() {
			input := minimalSpec()
			input.Spec.CreateMode = AzureMysqlFlexibleServerCreateMode_REPLICA
			input.Spec.SourceServerId = literal("/subscriptions/s/resourceGroups/rg/providers/Microsoft.DBforMySQL/flexibleServers/primary")
			input.Spec.SkuName = ""
			input.Spec.AdministratorLogin = ""
			input.Spec.AdministratorPassword = nil
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept replica promotion via replication_role NONE", func() {
			input := minimalSpec()
			input.Spec.CreateMode = AzureMysqlFlexibleServerCreateMode_REPLICA
			input.Spec.SourceServerId = literal("/subscriptions/s/resourceGroups/rg/providers/Microsoft.DBforMySQL/flexibleServers/primary")
			input.Spec.ReplicationRole = AzureMysqlFlexibleServerReplicationRole_NONE
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept a point-in-time restore with source and timestamp", func() {
			input := minimalSpec()
			input.Spec.CreateMode = AzureMysqlFlexibleServerCreateMode_POINT_IN_TIME_RESTORE
			input.Spec.SourceServerId = literal("/subscriptions/s/resourceGroups/rg/providers/Microsoft.DBforMySQL/flexibleServers/primary")
			input.Spec.PointInTimeRestoreTimeInUtc = "2026-07-01T08:30:00Z"
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept a geo-restore with a source and no timestamp", func() {
			input := minimalSpec()
			input.Spec.CreateMode = AzureMysqlFlexibleServerCreateMode_GEO_RESTORE
			input.Spec.SourceServerId = literal("/subscriptions/s/resourceGroups/rg/providers/Microsoft.DBforMySQL/flexibleServers/primary")
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept customer-managed-key encryption with a user-assigned identity", func() {
			input := minimalSpec()
			input.Spec.UserAssignedIdentityIds = []*foreignkeyv1.StringValueOrRef{literal(uaiId)}
			input.Spec.CustomerManagedKey = &AzureMysqlFlexibleServerCustomerManagedKey{
				KeyVaultKeyId:                 literal("https://vault.vault.azure.net/keys/mysql-cmk"),
				PrimaryUserAssignedIdentityId: literal(uaiId),
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept CMK with the geo-backup pair set together", func() {
			input := minimalSpec()
			input.Spec.GeoRedundantBackupEnabled = true
			input.Spec.UserAssignedIdentityIds = []*foreignkeyv1.StringValueOrRef{literal(uaiId)}
			input.Spec.CustomerManagedKey = &AzureMysqlFlexibleServerCustomerManagedKey{
				KeyVaultKeyId:                   literal("https://vault.vault.azure.net/keys/mysql-cmk"),
				PrimaryUserAssignedIdentityId:   literal(uaiId),
				GeoBackupKeyVaultKeyId:          literal("https://vault2.vault.azure.net/keys/mysql-cmk-geo"),
				GeoBackupUserAssignedIdentityId: literal(uaiId + "-geo"),
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept an Entra administrator backed by an attached identity", func() {
			input := minimalSpec()
			input.Spec.UserAssignedIdentityIds = []*foreignkeyv1.StringValueOrRef{literal(uaiId)}
			input.Spec.AadAdministrator = &AzureMysqlFlexibleServerAadAdministrator{
				IdentityId: literal(uaiId),
				Login:      "dba-team@contoso.com",
				ObjectId:   literal("11111111-2222-3333-4444-555555555555"),
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept databases, firewall rules, server parameters, and tags", func() {
			charset := "latin1"
			collation := "latin1_swedish_ci"
			input := minimalSpec()
			input.Spec.Databases = []*AzureMysqlFlexibleServerDatabase{
				{Name: "myapp"},
				{Name: "legacy", Charset: &charset, Collation: &collation},
			}
			input.Spec.FirewallRules = []*AzureMysqlFlexibleServerFirewallRule{
				{Name: "allow-office", StartIpAddress: "203.0.113.0", EndIpAddress: "203.0.113.255"},
				{Name: "allow-azure", StartIpAddress: "0.0.0.0", EndIpAddress: "0.0.0.0"},
			}
			input.Spec.ServerParameters = map[string]string{
				"max_connections":          "500",
				"require_secure_transport": "ON",
			}
			input.Spec.Tags = map[string]string{"team": "data"}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept valueFrom references for resource group and password", func() {
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

		ginkgo.It("should reject a server name with uppercase characters", func() {
			input := minimalSpec()
			input.Spec.ServerName = "Invalid-Name"
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject a server name ending with a hyphen", func() {
			input := minimalSpec()
			input.Spec.ServerName = "bad-name-"
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject a reserved administrator login", func() {
			for _, login := range []string{"azure_superuser", "admin", "Administrator", "root", "guest", "public"} {
				input := minimalSpec()
				input.Spec.AdministratorLogin = login
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil(), "login %q must be rejected", login)
			}
		})

		ginkgo.It("should reject an administrator login with illegal characters", func() {
			input := minimalSpec()
			input.Spec.AdministratorLogin = "my-admin"
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject an unsupported version", func() {
			for _, v := range []string{"8.0", "5.6", "9.0"} {
				bad := v
				input := minimalSpec()
				input.Spec.Version = &bad
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil(), "version %q must be rejected", v)
			}
		})

		ginkgo.It("should reject a malformed sku_name", func() {
			input := minimalSpec()
			input.Spec.SkuName = "Standard_D2ds_v4"
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject storage size outside 20-16384", func() {
			for _, gb := range []int32{10, 20000} {
				size := gb
				input := minimalSpec()
				input.Spec.Storage = &AzureMysqlFlexibleServerStorage{SizeGb: &size}
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			}
		})

		ginkgo.It("should reject IOPS outside 360-48000", func() {
			for _, v := range []int32{100, 50000} {
				iops := v
				input := minimalSpec()
				input.Spec.Storage = &AzureMysqlFlexibleServerStorage{Iops: &iops}
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			}
		})

		ginkgo.It("should reject provisioned IOPS combined with elastic IO scaling", func() {
			iops := int32(1000)
			input := minimalSpec()
			input.Spec.Storage = &AzureMysqlFlexibleServerStorage{
				Iops:             &iops,
				IoScalingEnabled: true,
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject an invalid zone", func() {
			input := minimalSpec()
			input.Spec.Zone = "4"
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject an HA block without a mode", func() {
			input := minimalSpec()
			input.Spec.HighAvailability = &AzureMysqlFlexibleServerHighAvailability{}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject an invalid standby zone", func() {
			input := minimalSpec()
			input.Spec.HighAvailability = &AzureMysqlFlexibleServerHighAvailability{
				Mode:                    AzureMysqlFlexibleServerHighAvailabilityMode_SAME_ZONE,
				StandbyAvailabilityZone: "5",
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject backup retention outside 1-35", func() {
			for _, days := range []int32{0, 40} {
				d := days
				input := minimalSpec()
				input.Spec.BackupRetentionDays = &d
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			}
		})

		ginkgo.It("should accept a maintenance window at Sunday midnight stated explicitly", func() {
			input := minimalSpec()
			input.Spec.MaintenanceWindow = &AzureMysqlFlexibleServerMaintenanceWindow{
				DayOfWeek:   proto.Int32(0),
				StartHour:   proto.Int32(0),
				StartMinute: proto.Int32(0),
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should reject a maintenance window with an out-of-range day", func() {
			input := minimalSpec()
			input.Spec.MaintenanceWindow = &AzureMysqlFlexibleServerMaintenanceWindow{DayOfWeek: proto.Int32(7)}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject a replica without source_server_id", func() {
			input := minimalSpec()
			input.Spec.CreateMode = AzureMysqlFlexibleServerCreateMode_REPLICA
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject source_server_id on a fresh (DEFAULT) server", func() {
			input := minimalSpec()
			input.Spec.SourceServerId = literal("/subscriptions/s/resourceGroups/rg/providers/Microsoft.DBforMySQL/flexibleServers/other")
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject a point-in-time restore without a timestamp", func() {
			input := minimalSpec()
			input.Spec.CreateMode = AzureMysqlFlexibleServerCreateMode_POINT_IN_TIME_RESTORE
			input.Spec.SourceServerId = literal("/subscriptions/s/resourceGroups/rg/providers/Microsoft.DBforMySQL/flexibleServers/primary")
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject a restore timestamp on a geo-restore", func() {
			input := minimalSpec()
			input.Spec.CreateMode = AzureMysqlFlexibleServerCreateMode_GEO_RESTORE
			input.Spec.SourceServerId = literal("/subscriptions/s/resourceGroups/rg/providers/Microsoft.DBforMySQL/flexibleServers/primary")
			input.Spec.PointInTimeRestoreTimeInUtc = "2026-07-01T08:30:00Z"
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject a malformed restore timestamp", func() {
			input := minimalSpec()
			input.Spec.CreateMode = AzureMysqlFlexibleServerCreateMode_POINT_IN_TIME_RESTORE
			input.Spec.SourceServerId = literal("/subscriptions/s/resourceGroups/rg/providers/Microsoft.DBforMySQL/flexibleServers/primary")
			input.Spec.PointInTimeRestoreTimeInUtc = "yesterday"
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject replication_role on a non-replica server", func() {
			input := minimalSpec()
			input.Spec.ReplicationRole = AzureMysqlFlexibleServerReplicationRole_NONE
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject a fresh server without credentials", func() {
			input := minimalSpec()
			input.Spec.AdministratorLogin = ""
			input.Spec.AdministratorPassword = nil
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject a fresh server without sku_name", func() {
			input := minimalSpec()
			input.Spec.SkuName = ""
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject VNet injection without a private DNS zone", func() {
			input := minimalSpec()
			input.Spec.DelegatedSubnetId = literal("/subscriptions/s/resourceGroups/rg/providers/Microsoft.Network/virtualNetworks/vnet/subnets/mysql")
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject ENABLED public access on a VNet-injected server", func() {
			input := minimalSpec()
			input.Spec.DelegatedSubnetId = literal("/subscriptions/s/resourceGroups/rg/providers/Microsoft.Network/virtualNetworks/vnet/subnets/mysql")
			input.Spec.PrivateDnsZoneId = literal("/subscriptions/s/resourceGroups/rg/providers/Microsoft.Network/privateDnsZones/mysql.database.azure.com")
			input.Spec.PublicNetworkAccess = AzureMysqlFlexibleServerPublicNetworkAccess_ENABLED
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject CMK without any attached user-assigned identity", func() {
			input := minimalSpec()
			input.Spec.CustomerManagedKey = &AzureMysqlFlexibleServerCustomerManagedKey{
				KeyVaultKeyId:                 literal("https://vault.vault.azure.net/keys/mysql-cmk"),
				PrimaryUserAssignedIdentityId: literal(uaiId),
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject CMK without primary_user_assigned_identity_id", func() {
			input := minimalSpec()
			input.Spec.UserAssignedIdentityIds = []*foreignkeyv1.StringValueOrRef{literal(uaiId)}
			input.Spec.CustomerManagedKey = &AzureMysqlFlexibleServerCustomerManagedKey{
				KeyVaultKeyId: literal("https://vault.vault.azure.net/keys/mysql-cmk"),
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject a geo-backup key without its identity", func() {
			input := minimalSpec()
			input.Spec.UserAssignedIdentityIds = []*foreignkeyv1.StringValueOrRef{literal(uaiId)}
			input.Spec.CustomerManagedKey = &AzureMysqlFlexibleServerCustomerManagedKey{
				KeyVaultKeyId:                 literal("https://vault.vault.azure.net/keys/mysql-cmk"),
				PrimaryUserAssignedIdentityId: literal(uaiId),
				GeoBackupKeyVaultKeyId:        literal("https://vault2.vault.azure.net/keys/mysql-cmk-geo"),
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject an Entra administrator without any attached identity", func() {
			input := minimalSpec()
			input.Spec.AadAdministrator = &AzureMysqlFlexibleServerAadAdministrator{
				IdentityId: literal(uaiId),
				Login:      "dba-team@contoso.com",
				ObjectId:   literal("11111111-2222-3333-4444-555555555555"),
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject an Entra administrator missing its identity_id", func() {
			input := minimalSpec()
			input.Spec.UserAssignedIdentityIds = []*foreignkeyv1.StringValueOrRef{literal(uaiId)}
			input.Spec.AadAdministrator = &AzureMysqlFlexibleServerAadAdministrator{
				Login:    "dba-team@contoso.com",
				ObjectId: literal("11111111-2222-3333-4444-555555555555"),
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject an Entra administrator with a malformed tenant", func() {
			tenant := "not-a-uuid"
			input := minimalSpec()
			input.Spec.UserAssignedIdentityIds = []*foreignkeyv1.StringValueOrRef{literal(uaiId)}
			input.Spec.AadAdministrator = &AzureMysqlFlexibleServerAadAdministrator{
				IdentityId: literal(uaiId),
				Login:      "dba-team@contoso.com",
				ObjectId:   literal("11111111-2222-3333-4444-555555555555"),
				TenantId:   &tenant,
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject a database name longer than 64 characters", func() {
			input := minimalSpec()
			input.Spec.Databases = []*AzureMysqlFlexibleServerDatabase{
				{Name: "a123456789a123456789a123456789a123456789a123456789a123456789a1234"},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject a firewall rule with a malformed IP", func() {
			input := minimalSpec()
			input.Spec.FirewallRules = []*AzureMysqlFlexibleServerFirewallRule{
				{Name: "bad", StartIpAddress: "not-an-ip", EndIpAddress: "0.0.0.0"},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject a firewall rule name with illegal characters", func() {
			input := minimalSpec()
			input.Spec.FirewallRules = []*AzureMysqlFlexibleServerFirewallRule{
				{Name: "bad rule!", StartIpAddress: "0.0.0.0", EndIpAddress: "0.0.0.0"},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject an empty server parameter value", func() {
			input := minimalSpec()
			input.Spec.ServerParameters = map[string]string{"max_connections": ""}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})
	})
})
