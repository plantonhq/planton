package module

import (
	azurecosmosdbsqlroledefinitionv1alpha1 "github.com/plantonhq/planton/catalog/azure/azurecosmosdbsqlroledefinition/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AzureCosmosdbSqlRoleDefinition *azurecosmosdbsqlroledefinitionv1alpha1.AzureCosmosdbSqlRoleDefinition
	CosmosdbAccountId              string
}

func initializeLocals(ctx *pulumi.Context, iacInput *azurecosmosdbsqlroledefinitionv1alpha1.AzureCosmosdbSqlRoleDefinitionIacInput) *Locals {
	locals := &Locals{}

	locals.AzureCosmosdbSqlRoleDefinition = iacInput.Target
	locals.CosmosdbAccountId = iacInput.Target.Spec.CosmosdbAccountId.GetValue()

	// No Azure tags: ARM does not support tags on Cosmos child
	// resources, so the platform's identity tags live on the account.

	return locals
}
