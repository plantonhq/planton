package module

import (
	"github.com/pkg/errors"
	azureuserassignedidentityv1alpha1 "github.com/plantonhq/planton/catalog/azure/azureuserassignedidentity/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/pulumiazureprovider"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/authorization"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *azureuserassignedidentityv1alpha1.AzureUserAssignedIdentityIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the Azure provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static client secret, keyless web identity, or ambient chain).
	azureProvider, err := pulumiazureprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzureUserAssignedIdentity.Spec

	// Lifecycle notes worth knowing before operating this resource:
	// - tags and isolation_scope update IN PLACE; name, region, and resource
	//   group are the identity's ARM identity, so changing any of them
	//   replaces it -- which mints a NEW principal and client ID, silently
	//   invalidating every existing grant and federated trust rule that
	//   referenced the old ones. Composed environments recover automatically
	//   (references re-resolve); externally-wired consumers do not.
	// - The identity is deliberately just the identity: grants live in
	//   AzureRoleAssignment and keyless trust rules in
	//   AzureFederatedIdentityCredential, both referencing this identity's
	//   outputs.
	identityArgs := &authorization.UserAssignedIdentityArgs{
		Name:              pulumi.String(spec.Name),
		Location:          pulumi.String(spec.Region),
		ResourceGroupName: pulumi.String(locals.ResourceGroupName),
		Tags:              pulumi.ToStringMap(locals.AzureTags),
	}

	// Omitted means ARM's default (no isolation -- usable from any region);
	// only the opt-in "Regional" mode is ever sent, so an unspecified spec
	// and Azure's default deploy identically on both engines.
	if locals.IsolationScope != "" {
		identityArgs.IsolationScope = pulumi.String(locals.IsolationScope)
	}

	identity, err := authorization.NewUserAssignedIdentity(ctx,
		spec.Name,
		identityArgs,
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create user-assigned managed identity %s", spec.Name)
	}

	// Export outputs from the created resource. principal_id is what
	// role assignments grant to; client_id is what workloads present to
	// authenticate as the identity; identity_id is what consuming resources
	// (and federated credentials) attach to.
	ctx.Export(OpIdentityId, identity.ID())
	ctx.Export(OpPrincipalId, identity.PrincipalId)
	ctx.Export(OpClientId, identity.ClientId)
	ctx.Export(OpTenantId, identity.TenantId)

	return nil
}
