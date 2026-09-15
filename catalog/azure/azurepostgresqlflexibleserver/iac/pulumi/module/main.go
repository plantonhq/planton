package module

import (
	"fmt"

	"github.com/pkg/errors"
	azurepostgresqlflexibleserverv1alpha1 "github.com/plantonhq/planton/catalog/azure/azurepostgresqlflexibleserver/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/pulumiazureprovider"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/core"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/postgresql"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *azurepostgresqlflexibleserverv1alpha1.AzurePostgresqlFlexibleServerStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	// Build the Azure provider from the stack input via the shared builder, which resolves
	// the right credential mechanism (static client secret, keyless web identity, or ambient chain).
	azureProvider, err := pulumiazureprovider.Get(ctx, stackInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzurePostgresqlFlexibleServer.Spec

	// The deploying credential's context -- the tenant fallback for Entra
	// (AAD) authentication and administrator grants when the spec does not
	// pin one.
	clientConfig, err := core.GetClientConfig(ctx, pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrap(err, "failed to read the azure client configuration")
	}

	// The Entra tenant for AAD auth and administrator grants: the spec's
	// explicit tenant wins, otherwise the deploying credential's tenant.
	aadTenantId := clientConfig.TenantId
	if spec.Authentication != nil && spec.Authentication.GetTenantId() != "" {
		aadTenantId = spec.Authentication.GetTenantId()
	}
	aadAuthEnabled := spec.Authentication != nil && spec.Authentication.ActiveDirectoryAuthEnabled

	serverArgs := &postgresql.FlexibleServerArgs{
		Name:              pulumi.String(spec.ServerName),
		Location:          pulumi.String(spec.Region),
		ResourceGroupName: pulumi.String(locals.ResourceGroupName),
		Tags:              pulumi.ToStringMap(locals.AzureTags),
	}

	// Lifecycle: how the server comes into existence. An unspecified mode
	// means a fresh (DEFAULT) server and is not sent at all -- azurerm
	// treats an omitted create_mode identically, keeping both engines'
	// payloads aligned. The replica/restore modes consume the source
	// server ID (and, for restores, the timestamp); all fixed at creation.
	isDefaultMode := true
	if spec.CreateMode != azurepostgresqlflexibleserverv1alpha1.AzurePostgresqlFlexibleServerCreateMode_azure_postgresql_flexible_server_create_mode_unspecified {
		serverArgs.CreateMode = pulumi.String(createModeStrings[spec.CreateMode])
		isDefaultMode = spec.CreateMode == azurepostgresqlflexibleserverv1alpha1.AzurePostgresqlFlexibleServerCreateMode_DEFAULT
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
	if spec.ReplicationRole == azurepostgresqlflexibleserverv1alpha1.AzurePostgresqlFlexibleServerReplicationRole_NONE {
		serverArgs.ReplicationRole = pulumi.String("None")
	}

	// Password-auth credentials -- omitted on Entra-only servers and on
	// replicas/restores, which inherit from the source.
	if spec.AdministratorLogin != "" {
		serverArgs.AdministratorLogin = pulumi.String(spec.AdministratorLogin)
	}
	if spec.AdministratorPassword.GetValue() != "" {
		serverArgs.AdministratorPassword = pulumi.String(spec.AdministratorPassword.GetValue())
	}

	// Version is only sent for a fresh server: replicas and restores
	// inherit the source's version. Presence-guarded to the spec default
	// (16) -- stack inputs built from a manifest do NOT materialize proto
	// defaults.
	if isDefaultMode {
		if spec.Version != nil {
			serverArgs.Version = pulumi.String(spec.GetVersion())
		} else {
			serverArgs.Version = pulumi.String("16")
		}
	}

	// Compute and storage. A replica left unset inherits the source's SKU
	// and size; storage_tier buys IOPS within the size's valid tier range
	// (spec validation mirrors Azure's matrix before the deploy runs).
	if spec.SkuName != "" {
		serverArgs.SkuName = pulumi.String(spec.SkuName)
	}
	if spec.StorageMb != nil {
		serverArgs.StorageMb = pulumi.Int(int(spec.GetStorageMb()))
	}
	if spec.StorageTier != azurepostgresqlflexibleserverv1alpha1.AzurePostgresqlFlexibleServerStorageTier_azure_postgresql_flexible_server_storage_tier_unspecified {
		serverArgs.StorageTier = pulumi.String(spec.StorageTier.String())
	}
	serverArgs.AutoGrowEnabled = pulumi.Bool(spec.AutoGrowEnabled)

	// Networking: the public endpoint dial and the VNet-injection pair.
	// public_network_access_enabled is presence-guarded to Azure's default
	// (true); spec validation already forces it explicitly false on a
	// VNet-injected server.
	if spec.PublicNetworkAccessEnabled != nil {
		serverArgs.PublicNetworkAccessEnabled = pulumi.Bool(spec.GetPublicNetworkAccessEnabled())
	} else {
		serverArgs.PublicNetworkAccessEnabled = pulumi.Bool(true)
	}
	if spec.DelegatedSubnetId.GetValue() != "" {
		serverArgs.DelegatedSubnetId = pulumi.String(spec.DelegatedSubnetId.GetValue())
	}
	if spec.PrivateDnsZoneId.GetValue() != "" {
		serverArgs.PrivateDnsZoneId = pulumi.String(spec.PrivateDnsZoneId.GetValue())
	}

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

	// Authentication mechanisms. Omitting the block applies Azure's
	// default (password on, Entra off). password_auth_enabled is
	// presence-guarded to its spec default (true); the tenant is only sent
	// when Entra auth is on.
	if spec.Authentication != nil {
		authArgs := &postgresql.FlexibleServerAuthenticationArgs{
			ActiveDirectoryAuthEnabled: pulumi.Bool(spec.Authentication.ActiveDirectoryAuthEnabled),
		}
		if spec.Authentication.PasswordAuthEnabled != nil {
			authArgs.PasswordAuthEnabled = pulumi.Bool(spec.Authentication.GetPasswordAuthEnabled())
		} else {
			authArgs.PasswordAuthEnabled = pulumi.Bool(true)
		}
		if aadAuthEnabled {
			authArgs.TenantId = pulumi.String(aadTenantId)
		}
		serverArgs.Authentication = authArgs
	}

	// High availability: a standby with synchronous replication and
	// automatic failover. The standby zone is fixed at creation.
	if spec.HighAvailability != nil {
		haArgs := &postgresql.FlexibleServerHighAvailabilityArgs{
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
		serverArgs.MaintenanceWindow = &postgresql.FlexibleServerMaintenanceWindowArgs{
			DayOfWeek:   pulumi.Int(int(spec.MaintenanceWindow.GetDayOfWeek())),
			StartHour:   pulumi.Int(int(spec.MaintenanceWindow.GetStartHour())),
			StartMinute: pulumi.Int(int(spec.MaintenanceWindow.GetStartMinute())),
		}
	}

	// The server's managed identity. A user-assigned identity is required
	// for customer-managed-key encryption (it unwraps the key).
	if spec.Identity != nil {
		identityIds := make([]string, 0, len(spec.Identity.IdentityIds))
		for _, identityId := range spec.Identity.IdentityIds {
			identityIds = append(identityIds, identityId.GetValue())
		}
		identityArgs := &postgresql.FlexibleServerIdentityArgs{
			Type: pulumi.String(identityTypeStrings[spec.Identity.Type]),
		}
		if len(identityIds) > 0 {
			identityArgs.IdentityIds = pulumi.ToStringArray(identityIds)
		}
		serverArgs.Identity = identityArgs
	}

	// Customer-managed-key encryption (fixed at creation). The geo-backup
	// pair encrypts the paired-region backup data.
	if spec.CustomerManagedKey != nil {
		cmkArgs := &postgresql.FlexibleServerCustomerManagedKeyArgs{
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

	// Elastic cluster (PG 17+): a sharded, citus-based cluster of the
	// declared node count. Fixed at creation; the size grows in place but
	// never shrinks.
	if spec.Cluster != nil {
		clusterArgs := &postgresql.FlexibleServerClusterArgs{
			Size: pulumi.Int(int(spec.Cluster.Size)),
		}
		if spec.Cluster.DefaultDatabaseName != nil {
			clusterArgs.DefaultDatabaseName = pulumi.String(spec.Cluster.GetDefaultDatabaseName())
		}
		serverArgs.Cluster = clusterArgs
	}

	server, err := postgresql.NewFlexibleServer(ctx,
		spec.ServerName,
		serverArgs,
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create postgresql flexible server %s", spec.ServerName)
	}

	// Databases, one Azure sub-resource each -- fixed at creation;
	// changing any field replaces just that database, never the server.
	// Charset and collation are presence-guarded to their spec defaults.
	databaseIdMap := pulumi.StringMap{}
	for _, database := range spec.Databases {
		charset := "UTF8"
		if database.Charset != nil {
			charset = database.GetCharset()
		}
		collation := "en_US.utf8"
		if database.Collation != nil {
			collation = database.GetCollation()
		}
		createdDatabase, err := postgresql.NewFlexibleServerDatabase(ctx,
			fmt.Sprintf("%s-%s", spec.ServerName, database.Name),
			&postgresql.FlexibleServerDatabaseArgs{
				Name:      pulumi.String(database.Name),
				ServerId:  server.ID(),
				Charset:   pulumi.String(charset),
				Collation: pulumi.String(collation),
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
		if _, err := postgresql.NewFlexibleServerFirewallRule(ctx,
			fmt.Sprintf("%s-%s", spec.ServerName, rule.Name),
			&postgresql.FlexibleServerFirewallRuleArgs{
				Name:           pulumi.String(rule.Name),
				ServerId:       server.ID(),
				StartIpAddress: pulumi.String(rule.StartIpAddress),
				EndIpAddress:   pulumi.String(rule.EndIpAddress),
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
		if _, err := postgresql.NewFlexibleServerConfiguration(ctx,
			fmt.Sprintf("%s-%s", spec.ServerName, parameterName),
			&postgresql.FlexibleServerConfigurationArgs{
				Name:     pulumi.String(parameterName),
				ServerId: server.ID(),
				Value:    pulumi.String(parameterValue),
			},
			pulumi.Provider(azureProvider),
			pulumi.Parent(server)); err != nil {
			return errors.Wrapf(err, "failed to set server parameter %s", parameterName)
		}
	}

	// Microsoft Entra administrator grants, keyed by the principal's
	// object ID. Azure validates principal_type against the directory
	// object, and the grant rides the same tenant as the server's Entra
	// auth configuration.
	for _, administrator := range spec.AadAdministrators {
		if _, err := postgresql.NewFlexibleServerActiveDirectoryAdministrator(ctx,
			fmt.Sprintf("%s-%s", spec.ServerName, administrator.ObjectId.GetValue()),
			&postgresql.FlexibleServerActiveDirectoryAdministratorArgs{
				ServerName:        server.Name,
				ResourceGroupName: pulumi.String(locals.ResourceGroupName),
				TenantId:          pulumi.String(aadTenantId),
				ObjectId:          pulumi.String(administrator.ObjectId.GetValue()),
				PrincipalName:     pulumi.String(administrator.PrincipalName),
				PrincipalType:     pulumi.String(principalTypeStrings[administrator.PrincipalType]),
			},
			pulumi.Provider(azureProvider),
			pulumi.Parent(server)); err != nil {
			return errors.Wrapf(err, "failed to create entra administrator %s", administrator.PrincipalName)
		}
	}

	// Export stack outputs from the created resources.
	ctx.Export(OpServerId, server.ID())
	ctx.Export(OpServerName, server.Name)
	ctx.Export(OpFqdn, server.Fqdn)
	ctx.Export(OpAdministratorLogin, server.AdministratorLogin)
	ctx.Export(OpDatabaseIds, databaseIdMap)

	// The system-assigned identity's principal ID -- empty unless the
	// identity type includes SYSTEM_ASSIGNED, matching the Terraform
	// module's conditional output.
	hasSystemIdentity := spec.Identity != nil &&
		(spec.Identity.Type == azurepostgresqlflexibleserverv1alpha1.AzurePostgresqlFlexibleServerIdentityType_SYSTEM_ASSIGNED ||
			spec.Identity.Type == azurepostgresqlflexibleserverv1alpha1.AzurePostgresqlFlexibleServerIdentityType_SYSTEM_AND_USER_ASSIGNED)
	if hasSystemIdentity {
		ctx.Export(OpIdentityPrincipalId, server.Identity.PrincipalId().Elem())
	} else {
		ctx.Export(OpIdentityPrincipalId, pulumi.String(""))
	}

	return nil
}
