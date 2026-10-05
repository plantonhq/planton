package module

import (
	"github.com/pkg/errors"
	azurerediscacheaccesspolicyassignmentv1alpha1 "github.com/plantonhq/planton/catalog/azure/azurerediscacheaccesspolicyassignment/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/pulumiazureprovider"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/redis"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *azurerediscacheaccesspolicyassignmentv1alpha1.AzureRedisCacheAccessPolicyAssignmentIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the Azure provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static client secret, keyless web identity, or ambient chain).
	azureProvider, err := pulumiazureprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzureRedisCacheAccessPolicyAssignment.Spec

	// The data-plane grant: assigns an access policy (built-in or an
	// AzureRedisCacheAccessPolicy) to a Microsoft Entra identity on the
	// cache -- the Redis analog of a role assignment. The granted
	// identity connects with its object ID (or the alias below) as the
	// Redis username and an Entra token as the password; Entra auth must
	// be enabled on the cache for the grant to matter.
	//
	// Every argument is ForceNew: replacing the assignment momentarily
	// revokes and re-grants, which is safe for the grant class. No tags:
	// ARM does not support tags on access policy assignments (cache
	// children).
	createdAssignment, err := redis.NewCacheAccessPolicyAssignment(ctx,
		spec.AssignmentName,
		&redis.CacheAccessPolicyAssignmentArgs{
			Name:         pulumi.String(spec.AssignmentName),
			RedisCacheId: pulumi.String(locals.RedisCacheId),

			AccessPolicyName: pulumi.String(spec.AccessPolicyName.GetValue()),

			// For a managed identity this must be the PRINCIPAL id --
			// granting the client id fails at connect time, not at
			// deploy time.
			ObjectId:      pulumi.String(spec.ObjectId.GetValue()),
			ObjectIdAlias: pulumi.String(spec.ObjectIdAlias),
		},
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create redis cache access policy assignment %s", spec.AssignmentName)
	}

	// Export outputs.
	ctx.Export(OpAccessPolicyAssignmentId, createdAssignment.ID())
	ctx.Export(OpAccessPolicyAssignmentName, createdAssignment.Name)

	return nil
}
