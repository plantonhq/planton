package module

import (
	"github.com/pkg/errors"
	azuremanagedredisaccesspolicyassignmentv1alpha1 "github.com/plantonhq/planton/catalog/azure/azuremanagedredisaccesspolicyassignment/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/pulumiazureprovider"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/managedredis"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *azuremanagedredisaccesspolicyassignmentv1alpha1.AzureManagedRedisAccessPolicyAssignmentIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the Azure provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static client secret, keyless web identity, or ambient chain).
	azureProvider, err := pulumiazureprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	// The data-plane grant: assigns Managed Redis's built-in "default"
	// access policy to a Microsoft Entra identity -- the Redis analog of
	// a role assignment, and the grant half of the keyless-by-default
	// story (access keys are off unless enabled, so grants are how
	// clients connect at all). The granted identity presents its object
	// ID as the Redis username and an Entra token as the password.
	//
	// Azure names the assignment after the object ID, so an identity is
	// granted at most once per database -- there is nothing else to
	// name. Every argument is ForceNew: replacing the assignment
	// momentarily revokes and re-grants, which is safe for the grant
	// class. No tags: ARM does not support tags on access policy
	// assignments (database children).
	createdAssignment, err := managedredis.NewAccessPolicyAssignment(ctx,
		locals.AzureManagedRedisAccessPolicyAssignment.Metadata.Name,
		&managedredis.AccessPolicyAssignmentArgs{
			ManagedRedisId: pulumi.String(locals.ManagedRedisId),

			// For a managed identity this must be the PRINCIPAL id --
			// granting the client id fails at connect time, not at
			// deploy time.
			ObjectId: pulumi.String(locals.ObjectId),
		},
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create managed redis access policy assignment for %s", locals.ObjectId)
	}

	// Export outputs. Azure names the assignment after the granted
	// object ID.
	ctx.Export(OpAccessPolicyAssignmentId, createdAssignment.ID())
	ctx.Export(OpAccessPolicyAssignmentName, createdAssignment.ObjectId)

	return nil
}
