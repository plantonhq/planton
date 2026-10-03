package module

import (
	azurecosmosdbsqlroleassignmentv1alpha1 "github.com/plantonhq/planton/catalog/azure/azurecosmosdbsqlroleassignment/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AzureCosmosdbSqlRoleAssignment *azurecosmosdbsqlroleassignmentv1alpha1.AzureCosmosdbSqlRoleAssignment
	CosmosdbAccountId              string
	RoleDefinitionId               string
	PrincipalId                    string
	Scope                          string
}

func initializeLocals(ctx *pulumi.Context, iacInput *azurecosmosdbsqlroleassignmentv1alpha1.AzureCosmosdbSqlRoleAssignmentIacInput) *Locals {
	locals := &Locals{}

	locals.AzureCosmosdbSqlRoleAssignment = iacInput.Target
	locals.CosmosdbAccountId = iacInput.Target.Spec.CosmosdbAccountId.GetValue()
	locals.RoleDefinitionId = iacInput.Target.Spec.RoleDefinitionId.GetValue()
	locals.PrincipalId = iacInput.Target.Spec.PrincipalId.GetValue()
	locals.Scope = iacInput.Target.Spec.Scope.GetValue()

	// No Azure tags: ARM does not support tags on Cosmos child
	// resources, so the platform's identity tags live on the account.

	return locals
}
