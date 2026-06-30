package module

import (
	"fmt"

	"github.com/pkg/errors"
	azuremssqlserverv1 "github.com/plantonhq/planton/apis/dev/planton/provider/azure/azuremssqlserver/v1"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/mssql"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *azuremssqlserverv1.AzureMssqlServerStackInput) error {
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

	spec := locals.AzureMssqlServer.Spec

	// Build the SQL Server arguments.
	// Azure SQL uses a logical server model: the server is an administrative
	// container with no compute or storage. Compute lives on each database.
	serverArgs := &mssql.ServerArgs{
		Name:                       pulumi.String(spec.Name),
		Location:                   pulumi.String(spec.Region),
		ResourceGroupName:          pulumi.String(locals.ResourceGroupName),
		AdministratorLogin:         pulumi.StringPtr(spec.AdministratorLogin),
		AdministratorLoginPassword: pulumi.StringPtr(spec.AdministratorPassword.GetValue()),
		Version:                    pulumi.String(spec.GetVersion()),
		MinimumTlsVersion:          pulumi.StringPtr(spec.GetMinimumTlsVersion()),
		PublicNetworkAccessEnabled: pulumi.BoolPtr(spec.GetPublicNetworkAccessEnabled()),
		ConnectionPolicy:           pulumi.StringPtr(spec.GetConnectionPolicy()),
		Tags:                       pulumi.ToStringMap(locals.AzureTags),
	}

	// Create the SQL Server.
	server, err := mssql.NewServer(ctx,
		spec.Name,
		serverArgs,
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create SQL Server %s", spec.Name)
	}

	// Create databases.
	// MSSQL databases carry their own compute SKU and storage (unlike PG/MySQL
	// where the server defines compute). Each database is an independent resource.
	databaseIdMap := make(map[string]pulumi.StringOutput)
	for _, db := range spec.Databases {
		dbArgs := &mssql.DatabaseArgs{
			Name:               pulumi.String(db.Name),
			ServerId:           server.ID(),
			SkuName:            pulumi.StringPtr(db.SkuName),
			Collation:          pulumi.StringPtr(db.GetCollation()),
			StorageAccountType: pulumi.StringPtr(db.GetStorageAccountType()),
		}

		// max_size_gb is optional; if not set, Azure uses the SKU default.
		// Pulumi SDK uses Float64 for this field (Azure supports fractional GB).
		if db.MaxSizeGb != nil {
			dbArgs.MaxSizeGb = pulumi.Float64Ptr(float64(*db.MaxSizeGb))
		}

		// Zone redundancy (supported on Premium DTU and Business Critical vCore).
		if db.ZoneRedundant != nil {
			dbArgs.ZoneRedundant = pulumi.BoolPtr(*db.ZoneRedundant)
		}

		// License type for Azure Hybrid Benefit.
		if db.LicenseType != nil {
			dbArgs.LicenseType = pulumi.StringPtr(*db.LicenseType)
		}

		database, err := mssql.NewDatabase(ctx,
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
	// Firewall rules use ServerId (same pattern as PostgreSQL).
	for _, rule := range spec.FirewallRules {
		_, err := mssql.NewFirewallRule(ctx,
			fmt.Sprintf("%s-%s", spec.Name, rule.Name),
			&mssql.FirewallRuleArgs{
				Name:           pulumi.String(rule.Name),
				ServerId:       server.ID(),
				StartIpAddress: pulumi.String(rule.StartIpAddress),
				EndIpAddress:   pulumi.String(rule.EndIpAddress),
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
	ctx.Export(OpFqdn, server.FullyQualifiedDomainName)
	ctx.Export(OpAdministratorLogin, pulumi.String(spec.AdministratorLogin))

	if len(databaseIdMap) > 0 {
		dbIdMapOutput := pulumi.StringMap{}
		for name, id := range databaseIdMap {
			dbIdMapOutput[name] = id
		}
		ctx.Export(OpDatabaseIds, dbIdMapOutput)
	}

	return nil
}
