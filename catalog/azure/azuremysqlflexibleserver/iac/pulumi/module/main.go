package module

import (
	"fmt"

	"github.com/pkg/errors"
	azuremysqlflexibleserverv1alpha1 "github.com/plantonhq/planton/catalog/azure/azuremysqlflexibleserver/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/pulumiazureprovider"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/core"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/mysql"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *azuremysqlflexibleserverv1alpha1.AzureMysqlFlexibleServerStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	// Build the Azure provider from the stack input via the shared builder, which resolves
	// the right credential mechanism (static client secret, keyless web identity, or ambient chain).
	azureProvider, err := pulumiazureprovider.Get(ctx, stackInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzureMysqlFlexibleServer.Spec

	// The deploying credential's context -- the tenant fallback for the
	// Microsoft Entra administrator grant when the spec does not pin one.
	clientConfig, err := core.GetClientConfig(ctx, pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrap(err, "failed to read the azure client configuration")
	}

	serverArgs := &mysql.FlexibleServerArgs{
		Name:              pulumi.String(spec.ServerName),
		Location:          pulumi.String(spec.Region),
		ResourceGroupName: pulumi.String(locals.ResourceGroupName),
		Tags:              pulumi.ToStringMap(locals.AzureTags),
	}

	// Lifecycle: how the server comes into existence. An unspecified mode
	// means a fresh (DEFAULT) server and is not sent at all -- azurerm
	// treats an omitted create_mode identically, keeping both engines'
	// payloads aligned. The replica/restore modes consume the source
	// server ID (and, for point-in-time restore, the timestamp); all
	// fixed at creation. GEO_RESTORE takes no timestamp -- it restores
	// the latest geo-replicated backup.
	isDefaultMode := true
	if spec.CreateMode != azuremysqlflexibleserverv1alpha1.AzureMysqlFlexibleServerCreateMode_azure_mysql_flexible_server_create_mode_unspecified {
		serverArgs.CreateMode = pulumi.String(createModeStrings[spec.CreateMode])
		isDefaultMode = spec.CreateMode == azuremysqlflexibleserverv1alpha1.AzureMysqlFlexibleServerCreateMode_DEFAULT
	}
	if spec.SourceServerId.GetValue() != "" {
		serverArgs.SourceServerId = pulumi.String(spec.SourceServerId.GetValue())
	}
	if spec.PointInTimeRestoreTimeInUtc != "" {
		serverArgs.PointInTimeRestoreTimeInUtc = pulumi.String(spec.PointInTimeRestoreTimeInUtc)
	}

	// Replica promotion (day-2 only): Azure rejects replication_role at
	// creation, and the only legal update value is "None" -- which breaks
	// replication and promotes the replica to a standalone primary.
	if spec.ReplicationRole == azuremysqlflexibleserverv1alpha1.AzureMysqlFlexibleServerReplicationRole_NONE {
		serverArgs.ReplicationRole = pulumi.String("None")
	}

	// Password-auth credentials -- omitted on replicas/restores, which
	// inherit from the source. MySQL always keeps password auth on
	// (unlike PostgreSQL Flexible Server, it cannot be disabled). The
	// login is fixed once set; the password rotates in place.
	if spec.AdministratorLogin != "" {
		serverArgs.AdministratorLogin = pulumi.String(spec.AdministratorLogin)
	}
	if spec.AdministratorPassword.GetValue() != "" {
		serverArgs.AdministratorPassword = pulumi.String(spec.AdministratorPassword.GetValue())
	}

	// Version is only sent for a fresh server: replicas and restores
	// inherit the source's version. Presence-guarded to the spec default
	// (8.0.21) -- stack inputs built from a manifest do NOT materialize
	// proto defaults.
	if isDefaultMode {
		if spec.Version != nil {
			serverArgs.Version = pulumi.String(spec.GetVersion())
		} else {
			serverArgs.Version = pulumi.String("8.0.21")
		}
	}

	// Compute. A replica left unset inherits the source's SKU (a replica
	// may legitimately override its compute by setting one).
	if spec.SkuName != "" {
		serverArgs.SkuName = pulumi.String(spec.SkuName)
	}

	// Storage: capacity only grows (shrinking replaces the server), and
	// iops is mutually exclusive with elastic IO scaling -- enforced by
	// spec validation before the deploy ever runs. Every optional dial is
	// presence-guarded so an unset field is simply not sent.
	if spec.Storage != nil {
		storageArgs := &mysql.FlexibleServerStorageArgs{}
		if spec.Storage.SizeGb != nil {
			storageArgs.SizeGb = pulumi.Int(int(spec.Storage.GetSizeGb()))
		}
		if spec.Storage.Iops != nil {
			storageArgs.Iops = pulumi.Int(int(spec.Storage.GetIops()))
		}
		if spec.Storage.AutoGrowEnabled != nil {
			storageArgs.AutoGrowEnabled = pulumi.Bool(spec.Storage.GetAutoGrowEnabled())
		} else {
			storageArgs.AutoGrowEnabled = pulumi.Bool(true)
		}
		storageArgs.IoScalingEnabled = pulumi.Bool(spec.Storage.IoScalingEnabled)
		storageArgs.LogOnDiskEnabled = pulumi.Bool(spec.Storage.LogOnDiskEnabled)
		serverArgs.Storage = storageArgs
	}

	// Networking: unset public_network_access is not sent at all, letting
	// Azure derive the value (Enabled publicly, Disabled when
	// VNet-injected) instead of the module guessing and fighting the
	// service; the injection pair is fixed at creation.
	if spec.PublicNetworkAccess != azuremysqlflexibleserverv1alpha1.AzureMysqlFlexibleServerPublicNetworkAccess_azure_mysql_flexible_server_public_network_access_unspecified {
		serverArgs.PublicNetworkAccess = pulumi.String(publicNetworkAccessStrings[spec.PublicNetworkAccess])
	}
	if spec.DelegatedSubnetId.GetValue() != "" {
		serverArgs.DelegatedSubnetId = pulumi.String(spec.DelegatedSubnetId.GetValue())
	}
	if spec.PrivateDnsZoneId.GetValue() != "" {
		serverArgs.PrivateDnsZoneId = pulumi.String(spec.PrivateDnsZoneId.GetValue())
	}

	// The primary's zone can only change via a planned failover that
	// swaps zone and standby_availability_zone -- Azure rejects an
	// independent zone change.
	if spec.Zone != "" {
		serverArgs.Zone = pulumi.String(spec.Zone)
	}

	// Backups. Retention is presence-guarded to Azure's default (7 days).
	if spec.BackupRetentionDays != nil {
		serverArgs.BackupRetentionDays = pulumi.Int(int(spec.GetBackupRetentionDays()))
	} else {
		serverArgs.BackupRetentionDays = pulumi.Int(7)
	}
	serverArgs.GeoRedundantBackupEnabled = pulumi.Bool(spec.GeoRedundantBackupEnabled)

	// High availability: a standby with synchronous replication and
	// automatic failover. Not supported on burstable SKUs or replicas.
	if spec.HighAvailability != nil {
		haArgs := &mysql.FlexibleServerHighAvailabilityArgs{
			Mode: pulumi.String(haModeStrings[spec.HighAvailability.Mode]),
		}
		if spec.HighAvailability.StandbyAvailabilityZone != "" {
			haArgs.StandbyAvailabilityZone = pulumi.String(spec.HighAvailability.StandbyAvailabilityZone)
		}
		serverArgs.HighAvailability = haArgs
	}

	// The weekly patching window. Azure applies it via a secondary update
	// right after creation; omitting the block leaves the window
	// system-managed.
	if spec.MaintenanceWindow != nil {
		serverArgs.MaintenanceWindow = &mysql.FlexibleServerMaintenanceWindowArgs{
			DayOfWeek:   pulumi.Int(int(spec.MaintenanceWindow.GetDayOfWeek())),
			StartHour:   pulumi.Int(int(spec.MaintenanceWindow.GetStartHour())),
			StartMinute: pulumi.Int(int(spec.MaintenanceWindow.GetStartMinute())),
		}
	}

	// The server's identities. MySQL Flexible Server supports
	// user-assigned identities only -- they unwrap the CMK and back the
	// Entra administrator grant.
	if len(spec.UserAssignedIdentityIds) > 0 {
		identityIds := make([]string, 0, len(spec.UserAssignedIdentityIds))
		for _, identityId := range spec.UserAssignedIdentityIds {
			identityIds = append(identityIds, identityId.GetValue())
		}
		serverArgs.Identity = &mysql.FlexibleServerIdentityArgs{
			Type:        pulumi.String("UserAssigned"),
			IdentityIds: pulumi.ToStringArray(identityIds),
		}
	}

	// Customer-managed-key encryption. The geo-backup pair encrypts the
	// paired-region backup data and is only meaningful with geo-redundant
	// backups.
	if spec.CustomerManagedKey != nil {
		cmkArgs := &mysql.FlexibleServerCustomerManagedKeyArgs{
			KeyVaultKeyId: pulumi.String(spec.CustomerManagedKey.KeyVaultKeyId.GetValue()),
		}
		if spec.CustomerManagedKey.PrimaryUserAssignedIdentityId.GetValue() != "" {
			cmkArgs.PrimaryUserAssignedIdentityId = pulumi.String(spec.CustomerManagedKey.PrimaryUserAssignedIdentityId.GetValue())
		}
		if spec.CustomerManagedKey.GeoBackupKeyVaultKeyId.GetValue() != "" {
			cmkArgs.GeoBackupKeyVaultKeyId = pulumi.String(spec.CustomerManagedKey.GeoBackupKeyVaultKeyId.GetValue())
		}
		if spec.CustomerManagedKey.GeoBackupUserAssignedIdentityId.GetValue() != "" {
			cmkArgs.GeoBackupUserAssignedIdentityId = pulumi.String(spec.CustomerManagedKey.GeoBackupUserAssignedIdentityId.GetValue())
		}
		serverArgs.CustomerManagedKey = cmkArgs
	}

	server, err := mysql.NewFlexibleServer(ctx,
		spec.ServerName,
		serverArgs,
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create mysql flexible server %s", spec.ServerName)
	}

	// Databases, one Azure sub-resource each -- fixed at creation;
	// changing any field replaces just that database, never the server.
	// Charset and collation are presence-guarded to their spec defaults.
	// MySQL's database resource addresses its parent by name + resource
	// group (not by server ID like PostgreSQL's).
	databaseIdMap := pulumi.StringMap{}
	for _, database := range spec.Databases {
		charset := "utf8mb4"
		if database.Charset != nil {
			charset = database.GetCharset()
		}
		collation := "utf8mb4_0900_ai_ci"
		if database.Collation != nil {
			collation = database.GetCollation()
		}
		createdDatabase, err := mysql.NewFlexibleDatabase(ctx,
			fmt.Sprintf("%s-%s", spec.ServerName, database.Name),
			&mysql.FlexibleDatabaseArgs{
				Name:              pulumi.String(database.Name),
				ResourceGroupName: pulumi.String(locals.ResourceGroupName),
				ServerName:        server.Name,
				Charset:           pulumi.String(charset),
				Collation:         pulumi.String(collation),
			},
			pulumi.Provider(azureProvider),
			pulumi.Parent(server))
		if err != nil {
			return errors.Wrapf(err, "failed to create database %s", database.Name)
		}
		databaseIdMap[database.Name] = createdDatabase.ID()
	}

	// Public-endpoint firewall rules. Only meaningful while the public
	// endpoint is enabled; Azure ignores them on a VNet-injected server.
	for _, rule := range spec.FirewallRules {
		if _, err := mysql.NewFlexibleServerFirewallRule(ctx,
			fmt.Sprintf("%s-%s", spec.ServerName, rule.Name),
			&mysql.FlexibleServerFirewallRuleArgs{
				Name:              pulumi.String(rule.Name),
				ResourceGroupName: pulumi.String(locals.ResourceGroupName),
				ServerName:        server.Name,
				StartIpAddress:    pulumi.String(rule.StartIpAddress),
				EndIpAddress:      pulumi.String(rule.EndIpAddress),
			},
			pulumi.Provider(azureProvider),
			pulumi.Parent(server)); err != nil {
			return errors.Wrapf(err, "failed to create firewall rule %s", rule.Name)
		}
	}

	// Server-parameter overrides. Azure applies each as a user override on
	// the per-SKU default; destroying the resource resets the parameter to
	// its default rather than deleting anything. Static (non-dynamic)
	// parameters report "pending restart" until the server restarts.
	for parameterName, parameterValue := range spec.ServerParameters {
		if _, err := mysql.NewFlexibleServerConfiguration(ctx,
			fmt.Sprintf("%s-%s", spec.ServerName, parameterName),
			&mysql.FlexibleServerConfigurationArgs{
				Name:              pulumi.String(parameterName),
				ResourceGroupName: pulumi.String(locals.ResourceGroupName),
				ServerName:        server.Name,
				Value:             pulumi.String(parameterValue),
			},
			pulumi.Provider(azureProvider),
			pulumi.Parent(server)); err != nil {
			return errors.Wrapf(err, "failed to set server parameter %s", parameterName)
		}
	}

	// The single Microsoft Entra administrator (MySQL supports exactly
	// one per server). The grant is backed by a user-assigned identity
	// attached to the server, which Azure uses to read directory objects
	// when validating Entra logins; the tenant falls back to the
	// deploying credential's.
	if spec.AadAdministrator != nil {
		tenantId := clientConfig.TenantId
		if spec.AadAdministrator.TenantId != nil && spec.AadAdministrator.GetTenantId() != "" {
			tenantId = spec.AadAdministrator.GetTenantId()
		}
		if _, err := mysql.NewFlexibleServerActiveDirectoryAdministratory(ctx,
			fmt.Sprintf("%s-aad-admin", spec.ServerName),
			&mysql.FlexibleServerActiveDirectoryAdministratoryArgs{
				ServerId:   server.ID(),
				IdentityId: pulumi.String(spec.AadAdministrator.IdentityId.GetValue()),
				Login:      pulumi.String(spec.AadAdministrator.Login),
				ObjectId:   pulumi.String(spec.AadAdministrator.ObjectId.GetValue()),
				TenantId:   pulumi.String(tenantId),
			},
			pulumi.Provider(azureProvider),
			pulumi.Parent(server)); err != nil {
			return errors.Wrapf(err, "failed to create entra administrator %s", spec.AadAdministrator.Login)
		}
	}

	// Export stack outputs from the created resources.
	ctx.Export(OpServerId, server.ID())
	ctx.Export(OpServerName, server.Name)
	ctx.Export(OpFqdn, server.Fqdn)
	ctx.Export(OpAdministratorLogin, server.AdministratorLogin)
	ctx.Export(OpDatabaseIds, databaseIdMap)
	ctx.Export(OpReplicaCapacity, server.ReplicaCapacity)

	return nil
}
