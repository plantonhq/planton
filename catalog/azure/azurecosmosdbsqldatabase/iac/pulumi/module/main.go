package module

import (
	"fmt"

	"github.com/pkg/errors"
	azurecosmosdbsqldatabasev1alpha1 "github.com/plantonhq/planton/catalog/azure/azurecosmosdbsqldatabase/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/pulumiazureprovider"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/cosmosdb"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *azurecosmosdbsqldatabasev1alpha1.AzureCosmosdbSqlDatabaseIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the Azure provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static client secret, keyless web identity, or ambient chain).
	azureProvider, err := pulumiazureprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzureCosmosdbSqlDatabase.Spec

	// The provider addresses Cosmos children by the (resource group,
	// account, name) trio rather than an ARM ID, so both are parsed
	// from the resolved account ID -- the spec models a single parent
	// reference and the module derives the rest (no redundant,
	// contradictable state). Parsing matches the Terraform module's
	// anchored regexes, so a malformed ID fails loudly here instead of
	// computing wrong names.
	accountName, resourceGroupName, err := parseCosmosdbAccountId(locals.CosmosdbAccountId)
	if err != nil {
		return err
	}

	// No Azure tags: ARM does not support tags on Cosmos child
	// resources, so the platform's identity tags live on the account.
	databaseArgs := &cosmosdb.SqlDatabaseArgs{
		Name:              pulumi.String(spec.DatabaseName),
		ResourceGroupName: pulumi.String(resourceGroupName),
		AccountName:       pulumi.String(accountName),
	}

	// Shared fixed throughput for the database's containers. Sent only
	// when set: serverless accounts reject provisioned throughput, and
	// unset means containers bring their own. The spec enforces mutual
	// exclusion with autoscale before the plan ever runs.
	if spec.Throughput != nil {
		databaseArgs.Throughput = pulumi.Int(int(spec.GetThroughput()))
	}
	if spec.AutoscaleMaxThroughput != nil {
		databaseArgs.AutoscaleSettings = cosmosdb.SqlDatabaseAutoscaleSettingsArgs{
			MaxThroughput: pulumi.Int(int(spec.GetAutoscaleMaxThroughput())),
		}
	}

	createdDatabase, err := cosmosdb.NewSqlDatabase(ctx,
		fmt.Sprintf("%s-%s", accountName, spec.DatabaseName),
		databaseArgs,
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create cosmosdb sql database %s", spec.DatabaseName)
	}

	// Export outputs. No endpoint or credential outputs on
	// purpose: connectivity and keys live on the ACCOUNT; the database
	// is addressed inside that connection by name.
	ctx.Export(OpSqlDatabaseId, createdDatabase.ID())
	ctx.Export(OpSqlDatabaseName, createdDatabase.Name)
	ctx.Export(OpCosmosdbAccountName, pulumi.String(accountName))

	return nil
}
