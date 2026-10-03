package module

import (
	azurecosmosdbsqldatabasev1alpha1 "github.com/plantonhq/planton/catalog/azure/azurecosmosdbsqldatabase/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AzureCosmosdbSqlDatabase *azurecosmosdbsqldatabasev1alpha1.AzureCosmosdbSqlDatabase
	CosmosdbAccountId        string
}

func initializeLocals(ctx *pulumi.Context, iacInput *azurecosmosdbsqldatabasev1alpha1.AzureCosmosdbSqlDatabaseIacInput) *Locals {
	locals := &Locals{}

	locals.AzureCosmosdbSqlDatabase = iacInput.Target
	locals.CosmosdbAccountId = iacInput.Target.Spec.CosmosdbAccountId.GetValue()

	// No Azure tags: ARM does not support tags on Cosmos child
	// resources, so the platform's identity tags live on the account.

	return locals
}
