package azurepostgresqlflexibleserverv1alpha1

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

func TestAzurePostgresqlFlexibleServerSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzurePostgresqlFlexibleServerSpec Validation Tests")
}

func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

// minimal valid spec: a fresh public-access server with password auth.
func minimalSpec() *AzurePostgresqlFlexibleServer {
	return &AzurePostgresqlFlexibleServer{
		ApiVersion: "azure.planton.dev/v1alpha1",
		Kind:       "AzurePostgresqlFlexibleServer",
		Metadata: &shared.CloudResourceMetadata{
			Name: "test-pg",
		},
		Spec: &AzurePostgresqlFlexibleServerSpec{
			Region:                "eastus",
			ResourceGroup:         literal("my-rg"),
			ServerName:            "test-pg-server",
			AdministratorLogin:    "pgadmin",
			AdministratorPassword: literal("P@ssw0rd1234!"),
			SkuName:               "GP_Standard_D2s_v3",
		},
	}
}

var _ = ginkgo.Describe("AzurePostgresqlFlexibleServerSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {

		ginkgo.It("should accept a minimal public server", func() {
			gomega.Expect(protovalidate.Validate(minimalSpec())).To(gomega.BeNil())
		})

		ginkgo.It("should accept every supported PostgreSQL version", func() {
			for _, v := range []string{"11", "12", "13", "14", "15", "16", "17", "18"} {
				ver := v
				input := minimalSpec()
				input.Spec.Version = &ver
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			}
		})

		ginkgo.It("should accept every storage_mb ladder size", func() {
			for _, mb := range []int32{32768, 65536, 131072, 262144, 524288, 1048576, 2097152, 4193280, 4194304, 8388608, 16777216, 33553408} {
				size := mb
				input := minimalSpec()
				input.Spec.StorageMb = &size
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			}
		})

		ginkgo.It("should accept a storage tier inside the size's valid range", func() {
			size := int32(131072)
			input := minimalSpec()
			input.Spec.StorageMb = &size
			input.Spec.StorageTier = AzurePostgresqlFlexibleServerStorageTier_P30
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept a non-default tier on the default (unset) storage size", func() {
			input := minimalSpec()
			input.Spec.StorageTier = AzurePostgresqlFlexibleServerStorageTier_P15
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept VNet injection with private DNS zone and public access off", func() {
			off := false
			input := minimalSpec()
			input.Spec.DelegatedSubnetId = literal("/subscriptions/s/resourceGroups/rg/providers/Microsoft.Network/virtualNetworks/vnet/subnets/pg")
			input.Spec.PrivateDnsZoneId = literal("/subscriptions/s/resourceGroups/rg/providers/Microsoft.Network/privateDnsZones/privatelink.postgres.database.azure.com")
			input.Spec.PublicNetworkAccessEnabled = &off
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept zone-redundant high availability with pinned zones", func() {
			input := minimalSpec()
			input.Spec.Zone = "1"
			input.Spec.HighAvailability = &AzurePostgresqlFlexibleServerHighAvailability{
				Mode:                    AzurePostgresqlFlexibleServerHighAvailabilityMode_ZONE_REDUNDANT,
				StandbyAvailabilityZone: "2",
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept a maintenance window", func() {
			input := minimalSpec()
			input.Spec.MaintenanceWindow = &AzurePostgresqlFlexibleServerMaintenanceWindow{
				DayOfWeek:   proto.Int32(6),
				StartHour:   proto.Int32(2),
				StartMinute: proto.Int32(30),
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept a replica that inherits sku and credentials from its source", func() {
			input := minimalSpec()
			input.Spec.CreateMode = AzurePostgresqlFlexibleServerCreateMode_REPLICA
			input.Spec.SourceServerId = literal("/subscriptions/s/resourceGroups/rg/providers/Microsoft.DBforPostgreSQL/flexibleServers/primary")
			input.Spec.SkuName = ""
			input.Spec.AdministratorLogin = ""
			input.Spec.AdministratorPassword = nil
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept replica promotion via replication_role NONE", func() {
			input := minimalSpec()
			input.Spec.CreateMode = AzurePostgresqlFlexibleServerCreateMode_REPLICA
			input.Spec.SourceServerId = literal("/subscriptions/s/resourceGroups/rg/providers/Microsoft.DBforPostgreSQL/flexibleServers/primary")
			input.Spec.ReplicationRole = AzurePostgresqlFlexibleServerReplicationRole_NONE
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept a point-in-time restore with source and timestamp", func() {
			input := minimalSpec()
			input.Spec.CreateMode = AzurePostgresqlFlexibleServerCreateMode_POINT_IN_TIME_RESTORE
			input.Spec.SourceServerId = literal("/subscriptions/s/resourceGroups/rg/providers/Microsoft.DBforPostgreSQL/flexibleServers/primary")
			input.Spec.PointInTimeRestoreTimeInUtc = "2026-07-01T08:30:00Z"
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept Entra (AAD) auth with administrators", func() {
			input := minimalSpec()
			input.Spec.Authentication = &AzurePostgresqlFlexibleServerAuthentication{
				ActiveDirectoryAuthEnabled: true,
			}
			input.Spec.AadAdministrators = []*AzurePostgresqlFlexibleServerAadAdministrator{
				{
					ObjectId:      literal("11111111-2222-3333-4444-555555555555"),
					PrincipalName: "dba-team@contoso.com",
					PrincipalType: AzurePostgresqlFlexibleServerAadPrincipalType_GROUP,
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept an Entra-only server with no password credentials", func() {
			off := false
			input := minimalSpec()
			input.Spec.AdministratorLogin = ""
			input.Spec.AdministratorPassword = nil
			input.Spec.Authentication = &AzurePostgresqlFlexibleServerAuthentication{
				ActiveDirectoryAuthEnabled: true,
				PasswordAuthEnabled:        &off,
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept customer-managed-key encryption with a user-assigned identity", func() {
			input := minimalSpec()
			input.Spec.Identity = &AzurePostgresqlFlexibleServerIdentity{
				Type:        AzurePostgresqlFlexibleServerIdentityType_USER_ASSIGNED,
				IdentityIds: []*foreignkeyv1.StringValueOrRef{literal("/subscriptions/s/resourceGroups/rg/providers/Microsoft.ManagedIdentity/userAssignedIdentities/pg-cmk")},
			}
			input.Spec.CustomerManagedKey = &AzurePostgresqlFlexibleServerCustomerManagedKey{
				KeyVaultKeyId:                 literal("https://vault.vault.azure.net/keys/pg-cmk"),
				PrimaryUserAssignedIdentityId: literal("/subscriptions/s/resourceGroups/rg/providers/Microsoft.ManagedIdentity/userAssignedIdentities/pg-cmk"),
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept an elastic cluster on PostgreSQL 17", func() {
			ver := "17"
			input := minimalSpec()
			input.Spec.Version = &ver
			input.Spec.Cluster = &AzurePostgresqlFlexibleServerCluster{Size: 4}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept databases, firewall rules, server parameters, and tags", func() {
			charset := "SQL_ASCII"
			collation := "C"
			input := minimalSpec()
			input.Spec.Databases = []*AzurePostgresqlFlexibleServerDatabase{
				{Name: "myapp"},
				{Name: "legacy", Charset: &charset, Collation: &collation},
			}
			input.Spec.FirewallRules = []*AzurePostgresqlFlexibleServerFirewallRule{
				{Name: "allow-office", StartIpAddress: "203.0.113.0", EndIpAddress: "203.0.113.255"},
				{Name: "allow-azure", StartIpAddress: "0.0.0.0", EndIpAddress: "0.0.0.0"},
			}
			input.Spec.ServerParameters = map[string]string{
				"shared_preload_libraries": "pg_stat_statements",
				"azure.extensions":         "PGCRYPTO,POSTGIS",
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
			for _, login := range []string{"azure_superuser", "admin", "Administrator", "pg_reader"} {
				input := minimalSpec()
				input.Spec.AdministratorLogin = login
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil(), "login %q must be rejected", login)
			}
		})

		ginkgo.It("should reject an unsupported version", func() {
			bad := "10"
			input := minimalSpec()
			input.Spec.Version = &bad
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject a malformed sku_name", func() {
			input := minimalSpec()
			input.Spec.SkuName = "Standard_D2s_v3"
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject an off-ladder storage_mb", func() {
			bad := int32(50000)
			input := minimalSpec()
			input.Spec.StorageMb = &bad
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject a storage tier below the size's floor", func() {
			size := int32(1048576) // 1 TiB starts at P30
			input := minimalSpec()
			input.Spec.StorageMb = &size
			input.Spec.StorageTier = AzurePostgresqlFlexibleServerStorageTier_P15
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject a storage tier above the size's ceiling", func() {
			size := int32(32768) // 32 GiB caps at P50
			input := minimalSpec()
			input.Spec.StorageMb = &size
			input.Spec.StorageTier = AzurePostgresqlFlexibleServerStorageTier_P60
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject an invalid zone", func() {
			input := minimalSpec()
			input.Spec.Zone = "4"
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject an HA block without a mode", func() {
			input := minimalSpec()
			input.Spec.HighAvailability = &AzurePostgresqlFlexibleServerHighAvailability{}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject backup retention outside 7-35", func() {
			for _, days := range []int32{3, 40} {
				d := days
				input := minimalSpec()
				input.Spec.BackupRetentionDays = &d
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			}
		})

		ginkgo.It("should accept a maintenance window at Sunday midnight stated explicitly", func() {
			input := minimalSpec()
			input.Spec.MaintenanceWindow = &AzurePostgresqlFlexibleServerMaintenanceWindow{
				DayOfWeek:   proto.Int32(0),
				StartHour:   proto.Int32(0),
				StartMinute: proto.Int32(0),
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should reject a maintenance window with an out-of-range day", func() {
			input := minimalSpec()
			input.Spec.MaintenanceWindow = &AzurePostgresqlFlexibleServerMaintenanceWindow{DayOfWeek: proto.Int32(7)}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject a replica without source_server_id", func() {
			input := minimalSpec()
			input.Spec.CreateMode = AzurePostgresqlFlexibleServerCreateMode_REPLICA
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject source_server_id on a fresh (DEFAULT) server", func() {
			input := minimalSpec()
			input.Spec.SourceServerId = literal("/subscriptions/s/resourceGroups/rg/providers/Microsoft.DBforPostgreSQL/flexibleServers/other")
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject a point-in-time restore without a timestamp", func() {
			input := minimalSpec()
			input.Spec.CreateMode = AzurePostgresqlFlexibleServerCreateMode_POINT_IN_TIME_RESTORE
			input.Spec.SourceServerId = literal("/subscriptions/s/resourceGroups/rg/providers/Microsoft.DBforPostgreSQL/flexibleServers/primary")
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject a restore timestamp on a replica", func() {
			input := minimalSpec()
			input.Spec.CreateMode = AzurePostgresqlFlexibleServerCreateMode_REPLICA
			input.Spec.SourceServerId = literal("/subscriptions/s/resourceGroups/rg/providers/Microsoft.DBforPostgreSQL/flexibleServers/primary")
			input.Spec.PointInTimeRestoreTimeInUtc = "2026-07-01T08:30:00Z"
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject a malformed restore timestamp", func() {
			input := minimalSpec()
			input.Spec.CreateMode = AzurePostgresqlFlexibleServerCreateMode_POINT_IN_TIME_RESTORE
			input.Spec.SourceServerId = literal("/subscriptions/s/resourceGroups/rg/providers/Microsoft.DBforPostgreSQL/flexibleServers/primary")
			input.Spec.PointInTimeRestoreTimeInUtc = "yesterday"
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject replication_role on a non-replica server", func() {
			input := minimalSpec()
			input.Spec.ReplicationRole = AzurePostgresqlFlexibleServerReplicationRole_NONE
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject a fresh password-auth server without credentials", func() {
			input := minimalSpec()
			input.Spec.AdministratorLogin = ""
			input.Spec.AdministratorPassword = nil
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject credentials when password auth is disabled", func() {
			off := false
			input := minimalSpec()
			input.Spec.Authentication = &AzurePostgresqlFlexibleServerAuthentication{
				ActiveDirectoryAuthEnabled: true,
				PasswordAuthEnabled:        &off,
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject disabling both authentication mechanisms", func() {
			off := false
			input := minimalSpec()
			input.Spec.AdministratorLogin = ""
			input.Spec.AdministratorPassword = nil
			input.Spec.Authentication = &AzurePostgresqlFlexibleServerAuthentication{
				ActiveDirectoryAuthEnabled: false,
				PasswordAuthEnabled:        &off,
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject a tenant_id without Entra auth enabled", func() {
			tenant := "11111111-2222-3333-4444-555555555555"
			input := minimalSpec()
			input.Spec.Authentication = &AzurePostgresqlFlexibleServerAuthentication{
				TenantId: &tenant,
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject a fresh server without sku_name", func() {
			input := minimalSpec()
			input.Spec.SkuName = ""
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject aad_administrators without Entra auth enabled", func() {
			input := minimalSpec()
			input.Spec.AadAdministrators = []*AzurePostgresqlFlexibleServerAadAdministrator{
				{
					ObjectId:      literal("11111111-2222-3333-4444-555555555555"),
					PrincipalName: "dba",
					PrincipalType: AzurePostgresqlFlexibleServerAadPrincipalType_USER,
				},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject an AAD administrator without a principal type", func() {
			input := minimalSpec()
			input.Spec.Authentication = &AzurePostgresqlFlexibleServerAuthentication{ActiveDirectoryAuthEnabled: true}
			input.Spec.AadAdministrators = []*AzurePostgresqlFlexibleServerAadAdministrator{
				{ObjectId: literal("11111111-2222-3333-4444-555555555555"), PrincipalName: "dba"},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject CMK without a user-assigned identity", func() {
			input := minimalSpec()
			input.Spec.CustomerManagedKey = &AzurePostgresqlFlexibleServerCustomerManagedKey{
				KeyVaultKeyId:                 literal("https://vault.vault.azure.net/keys/pg-cmk"),
				PrimaryUserAssignedIdentityId: literal("/subscriptions/s/resourceGroups/rg/providers/Microsoft.ManagedIdentity/userAssignedIdentities/pg-cmk"),
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject CMK without primary_user_assigned_identity_id", func() {
			input := minimalSpec()
			input.Spec.Identity = &AzurePostgresqlFlexibleServerIdentity{
				Type:        AzurePostgresqlFlexibleServerIdentityType_USER_ASSIGNED,
				IdentityIds: []*foreignkeyv1.StringValueOrRef{literal("/subscriptions/s/resourceGroups/rg/providers/Microsoft.ManagedIdentity/userAssignedIdentities/pg-cmk")},
			}
			input.Spec.CustomerManagedKey = &AzurePostgresqlFlexibleServerCustomerManagedKey{
				KeyVaultKeyId: literal("https://vault.vault.azure.net/keys/pg-cmk"),
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject a geo-backup key without its identity", func() {
			input := minimalSpec()
			input.Spec.Identity = &AzurePostgresqlFlexibleServerIdentity{
				Type:        AzurePostgresqlFlexibleServerIdentityType_USER_ASSIGNED,
				IdentityIds: []*foreignkeyv1.StringValueOrRef{literal("/subscriptions/s/resourceGroups/rg/providers/Microsoft.ManagedIdentity/userAssignedIdentities/pg-cmk")},
			}
			input.Spec.CustomerManagedKey = &AzurePostgresqlFlexibleServerCustomerManagedKey{
				KeyVaultKeyId:                 literal("https://vault.vault.azure.net/keys/pg-cmk"),
				PrimaryUserAssignedIdentityId: literal("/subscriptions/s/resourceGroups/rg/providers/Microsoft.ManagedIdentity/userAssignedIdentities/pg-cmk"),
				GeoBackupKeyVaultKeyId:        literal("https://vault2.vault.azure.net/keys/pg-cmk-geo"),
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject a SYSTEM_ASSIGNED identity carrying identity_ids", func() {
			input := minimalSpec()
			input.Spec.Identity = &AzurePostgresqlFlexibleServerIdentity{
				Type:        AzurePostgresqlFlexibleServerIdentityType_SYSTEM_ASSIGNED,
				IdentityIds: []*foreignkeyv1.StringValueOrRef{literal("/subscriptions/s/resourceGroups/rg/providers/Microsoft.ManagedIdentity/userAssignedIdentities/x")},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject an elastic cluster without version 17+", func() {
			input := minimalSpec()
			input.Spec.Cluster = &AzurePostgresqlFlexibleServerCluster{Size: 2}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject an elastic cluster on a replica", func() {
			ver := "17"
			input := minimalSpec()
			input.Spec.Version = &ver
			input.Spec.CreateMode = AzurePostgresqlFlexibleServerCreateMode_REPLICA
			input.Spec.SourceServerId = literal("/subscriptions/s/resourceGroups/rg/providers/Microsoft.DBforPostgreSQL/flexibleServers/primary")
			input.Spec.Cluster = &AzurePostgresqlFlexibleServerCluster{Size: 2}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject a cluster size above 20", func() {
			ver := "17"
			input := minimalSpec()
			input.Spec.Version = &ver
			input.Spec.Cluster = &AzurePostgresqlFlexibleServerCluster{Size: 21}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject VNet injection without a private DNS zone", func() {
			off := false
			input := minimalSpec()
			input.Spec.DelegatedSubnetId = literal("/subscriptions/s/resourceGroups/rg/providers/Microsoft.Network/virtualNetworks/vnet/subnets/pg")
			input.Spec.PublicNetworkAccessEnabled = &off
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject VNet injection without public access explicitly off", func() {
			input := minimalSpec()
			input.Spec.DelegatedSubnetId = literal("/subscriptions/s/resourceGroups/rg/providers/Microsoft.Network/virtualNetworks/vnet/subnets/pg")
			input.Spec.PrivateDnsZoneId = literal("/subscriptions/s/resourceGroups/rg/providers/Microsoft.Network/privateDnsZones/privatelink.postgres.database.azure.com")
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject a database name starting with a digit", func() {
			input := minimalSpec()
			input.Spec.Databases = []*AzurePostgresqlFlexibleServerDatabase{{Name: "1app"}}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject a firewall rule with a malformed IP", func() {
			input := minimalSpec()
			input.Spec.FirewallRules = []*AzurePostgresqlFlexibleServerFirewallRule{
				{Name: "bad", StartIpAddress: "not-an-ip", EndIpAddress: "0.0.0.0"},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject a firewall rule name with illegal characters", func() {
			input := minimalSpec()
			input.Spec.FirewallRules = []*AzurePostgresqlFlexibleServerFirewallRule{
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
