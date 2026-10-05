package module

import (
	azurecosmosdbsqlcontainerv1alpha1 "github.com/plantonhq/planton/catalog/azure/azurecosmosdbsqlcontainer/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AzureCosmosdbSqlContainer *azurecosmosdbsqlcontainerv1alpha1.AzureCosmosdbSqlContainer
	SqlDatabaseId             string
}

// partitionKeyKindStrings maps the spec's kind enum to ARM's wire
// values. Unspecified materializes Hash in main.go -- azurerm's own
// default.
var partitionKeyKindStrings = map[azurecosmosdbsqlcontainerv1alpha1.AzureCosmosdbSqlContainerPartitionKeyKind]string{
	azurecosmosdbsqlcontainerv1alpha1.AzureCosmosdbSqlContainerPartitionKeyKind_HASH:       "Hash",
	azurecosmosdbsqlcontainerv1alpha1.AzureCosmosdbSqlContainerPartitionKeyKind_MULTI_HASH: "MultiHash",
}

// indexingModeStrings maps the indexing-mode enum to ARM's wire values
// (lowercase -- the provider's contract). Unspecified materializes
// consistent in main.go.
var indexingModeStrings = map[azurecosmosdbsqlcontainerv1alpha1.AzureCosmosdbSqlContainerIndexingMode]string{
	azurecosmosdbsqlcontainerv1alpha1.AzureCosmosdbSqlContainerIndexingMode_CONSISTENT: "consistent",
	azurecosmosdbsqlcontainerv1alpha1.AzureCosmosdbSqlContainerIndexingMode_NONE:       "none",
}

// compositeIndexOrderStrings maps the composite-index order enum to
// ARM's wire values. Unspecified materializes Ascending in main.go.
var compositeIndexOrderStrings = map[azurecosmosdbsqlcontainerv1alpha1.AzureCosmosdbSqlContainerCompositeIndexOrder]string{
	azurecosmosdbsqlcontainerv1alpha1.AzureCosmosdbSqlContainerCompositeIndexOrder_ASCENDING:  "Ascending",
	azurecosmosdbsqlcontainerv1alpha1.AzureCosmosdbSqlContainerCompositeIndexOrder_DESCENDING: "Descending",
}

// conflictResolutionModeStrings maps the conflict-resolution mode enum
// to ARM's wire values.
var conflictResolutionModeStrings = map[azurecosmosdbsqlcontainerv1alpha1.AzureCosmosdbSqlContainerConflictResolutionMode]string{
	azurecosmosdbsqlcontainerv1alpha1.AzureCosmosdbSqlContainerConflictResolutionMode_LAST_WRITER_WINS: "LastWriterWins",
	azurecosmosdbsqlcontainerv1alpha1.AzureCosmosdbSqlContainerConflictResolutionMode_CUSTOM:           "Custom",
}

func initializeLocals(ctx *pulumi.Context, iacInput *azurecosmosdbsqlcontainerv1alpha1.AzureCosmosdbSqlContainerIacInput) *Locals {
	locals := &Locals{}

	locals.AzureCosmosdbSqlContainer = iacInput.Target
	locals.SqlDatabaseId = iacInput.Target.Spec.SqlDatabaseId.GetValue()

	// No Azure tags: ARM does not support tags on Cosmos child
	// resources, so the platform's identity tags live on the account.

	return locals
}
