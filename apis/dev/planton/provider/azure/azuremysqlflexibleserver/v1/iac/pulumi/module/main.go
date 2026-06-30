package module

import (
	"fmt"

	"github.com/pkg/errors"
	azuremysqlflexibleserverv1 "github.com/plantonhq/planton/apis/dev/planton/provider/azure/azuremysqlflexibleserver/v1"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/mysql"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *azuremysqlflexibleserverv1.AzureMysqlFlexibleServerStackInput) error {
	locals := initializeLocals(ctx, stackInput)
	azureProviderConfig := stackInput.ProviderConfig

	// Create azure provider using the credentials from the input
	azureProvider, err := azure.NewProvider(ctx,
		"azure",
		&azure.ProviderArgs{
			ClientId:       pulumi.String(azureProviderConfig.ClientId),
			ClientSecret:   pulumi.String(azureProviderConfig.ClientSecret),
			SubscriptionId: pulumi.String(azureProviderConfig.SubscriptionId),
			TenantId:       pulumi.String(azureProviderConfig.TenantId),
		})
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzureMysqlFlexibleServer.Spec

	// Build the flexible server arguments.
	// MySQL uses a Storage block (not flat StorageMb/AutoGrowEnabled fields),
	// PublicNetworkAccess string (not PublicNetworkAccessEnabled bool),
	// and has no Authentication block.
	serverArgs := &mysql.FlexibleServerArgs{
		Name:               pulumi.String(spec.Name),
		Location:           pulumi.String(spec.Region),
		ResourceGroupName:  pulumi.String(locals.ResourceGroupName),
		AdministratorLogin: pulumi.StringPtr(spec.AdministratorLogin),
		AdministratorPassword: pulumi.StringPtr(
			spec.AdministratorPassword.GetValue(),
		),
		Version: pulumi.StringPtr(spec.GetVersion()),
		SkuName: pulumi.StringPtr(spec.SkuName),
		Storage: &mysql.FlexibleServerStorageArgs{
			SizeGb:          pulumi.IntPtr(int(spec.StorageSizeGb)),
			AutoGrowEnabled: pulumi.BoolPtr(spec.GetAutoGrowEnabled()),
		},
		BackupRetentionDays:       pulumi.IntPtr(int(spec.GetBackupRetentionDays())),
		GeoRedundantBackupEnabled: pulumi.BoolPtr(spec.GetGeoRedundantBackupEnabled()),
		Tags:                      pulumi.ToStringMap(locals.AzureTags),
	}

	// Network access mode: VNet-integrated (private) vs public.
	// When delegated_subnet_id is set, the server is deployed with private access
	// and public network access is automatically disabled.
	// MySQL uses PublicNetworkAccess string ("Enabled"/"Disabled") instead of a bool.
	if spec.DelegatedSubnetId != nil && spec.DelegatedSubnetId.GetValue() != "" {
		serverArgs.DelegatedSubnetId = pulumi.StringPtr(spec.DelegatedSubnetId.GetValue())
		serverArgs.PublicNetworkAccess = pulumi.StringPtr("Disabled")
	} else {
		serverArgs.PublicNetworkAccess = pulumi.StringPtr("Enabled")
	}

	// Private DNS zone (optional, typically used with VNet integration).
	if spec.PrivateDnsZoneId != nil && spec.PrivateDnsZoneId.GetValue() != "" {
		serverArgs.PrivateDnsZoneId = pulumi.StringPtr(spec.PrivateDnsZoneId.GetValue())
	}

	// Availability zone for primary server.
	if spec.Zone != "" {
		serverArgs.Zone = pulumi.StringPtr(spec.Zone)
	}

	// High availability configuration.
	// If the HA message is present, HA is enabled with the specified mode.
	if spec.HighAvailability != nil {
		haArgs := &mysql.FlexibleServerHighAvailabilityArgs{
			Mode: pulumi.String(spec.HighAvailability.Mode),
		}
		if spec.HighAvailability.StandbyAvailabilityZone != "" {
			haArgs.StandbyAvailabilityZone = pulumi.StringPtr(spec.HighAvailability.StandbyAvailabilityZone)
		}
		serverArgs.HighAvailability = haArgs
	}

	// Create the MySQL Flexible Server.
	server, err := mysql.NewFlexibleServer(ctx,
		spec.Name,
		serverArgs,
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create MySQL Flexible Server %s", spec.Name)
	}

	// Create databases.
	// Each database is a separate resource with an explicit dependency on the server.
	// MySQL uses ServerName + ResourceGroupName (NOT ServerId like PostgreSQL).
	// We collect database IDs for the output map.
	databaseIdMap := make(map[string]pulumi.StringOutput)
	for _, db := range spec.Databases {
		dbArgs := &mysql.FlexibleDatabaseArgs{
			Name:              pulumi.String(db.Name),
			ServerName:        server.Name,
			ResourceGroupName: pulumi.String(locals.ResourceGroupName),
			Charset:           pulumi.String(db.GetCharset()),
			Collation:         pulumi.String(db.GetCollation()),
		}

		database, err := mysql.NewFlexibleDatabase(ctx,
			fmt.Sprintf("%s-%s", spec.Name, db.Name),
			dbArgs,
			pulumi.Provider(azureProvider),
			pulumi.DependsOn([]pulumi.Resource{server}))
		if err != nil {
			return errors.Wrapf(err, "failed to create database %s", db.Name)
		}
		databaseIdMap[db.Name] = database.ID().ToStringOutput()
	}

	// Create firewall rules.
	// Only effective in public access mode (when delegated_subnet_id is not set).
	// In VNet mode, firewall rules are ignored by Azure, but creating them doesn't
	// cause errors -- they simply have no effect.
	// MySQL uses ServerName + ResourceGroupName (NOT ServerId like PostgreSQL).
	for _, rule := range spec.FirewallRules {
		_, err := mysql.NewFlexibleServerFirewallRule(ctx,
			fmt.Sprintf("%s-%s", spec.Name, rule.Name),
			&mysql.FlexibleServerFirewallRuleArgs{
				Name:              pulumi.String(rule.Name),
				ServerName:        server.Name,
				ResourceGroupName: pulumi.String(locals.ResourceGroupName),
				StartIpAddress:    pulumi.String(rule.StartIpAddress),
				EndIpAddress:      pulumi.String(rule.EndIpAddress),
			},
			pulumi.Provider(azureProvider),
			pulumi.DependsOn([]pulumi.Resource{server}))
		if err != nil {
			return errors.Wrapf(err, "failed to create firewall rule %s", rule.Name)
		}
	}

	// Export stack outputs.
	ctx.Export(OpServerId, server.ID())
	ctx.Export(OpServerName, server.Name)
	ctx.Export(OpFqdn, server.Fqdn)
	ctx.Export(OpAdministratorLogin, pulumi.String(spec.AdministratorLogin))

	// Export database ID map.
	if len(databaseIdMap) > 0 {
		dbIdMapOutput := pulumi.StringMap{}
		for name, id := range databaseIdMap {
			dbIdMapOutput[name] = id
		}
		ctx.Export(OpDatabaseIds, dbIdMapOutput)
	}

	return nil
}
