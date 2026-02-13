package module

import (
	"fmt"

	"github.com/pkg/errors"
	azureuserassignedidentityv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/azure/azureuserassignedidentity/v1"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/authorization"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *azureuserassignedidentityv1.AzureUserAssignedIdentityStackInput) error {
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

	spec := locals.AzureUserAssignedIdentity.Spec

	// Create the User-Assigned Managed Identity
	identity, err := authorization.NewUserAssignedIdentity(ctx,
		spec.Name,
		&authorization.UserAssignedIdentityArgs{
			Name:              pulumi.String(spec.Name),
			Location:          pulumi.String(spec.Region),
			ResourceGroupName: pulumi.String(locals.ResourceGroupName),
			Tags:              pulumi.ToStringMap(locals.AzureTags),
		},
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create User-Assigned Managed Identity %s", spec.Name)
	}

	// Create role assignments for each entry in the spec.
	// Each role assignment binds the identity to a specific role at a specific scope.
	for i, ra := range spec.RoleAssignments {
		resolvedScope := locals.ResolvedScopes[i]

		_, err := authorization.NewAssignment(ctx,
			fmt.Sprintf("%s-ra-%d", spec.Name, i),
			&authorization.AssignmentArgs{
				PrincipalId:                 identity.PrincipalId,
				Scope:                       pulumi.String(resolvedScope),
				RoleDefinitionName:          pulumi.String(ra.RoleDefinitionName),
				SkipServicePrincipalAadCheck: pulumi.Bool(true),
			},
			pulumi.Provider(azureProvider),
			pulumi.DependsOn([]pulumi.Resource{identity}))
		if err != nil {
			return errors.Wrapf(err, "failed to create role assignment %d (%s at %s)",
				i, ra.RoleDefinitionName, resolvedScope)
		}
	}

	// Export stack outputs
	ctx.Export(OpIdentityId, identity.ID())
	ctx.Export(OpPrincipalId, identity.PrincipalId)
	ctx.Export(OpClientId, identity.ClientId)
	ctx.Export(OpTenantId, identity.TenantId)

	return nil
}
