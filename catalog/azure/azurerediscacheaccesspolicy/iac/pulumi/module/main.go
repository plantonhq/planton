package module

import (
	"github.com/pkg/errors"
	azurerediscacheaccesspolicyv1alpha1 "github.com/plantonhq/planton/catalog/azure/azurerediscacheaccesspolicy/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/pulumiazureprovider"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/redis"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *azurerediscacheaccesspolicyv1alpha1.AzureRedisCacheAccessPolicyIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the Azure provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static client secret, keyless web identity, or ambient chain).
	azureProvider, err := pulumiazureprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzureRedisCacheAccessPolicy.Spec

	// A CUSTOM data-plane access policy -- a named permission set in
	// Redis's own ACL syntax that AzureRedisCacheAccessPolicyAssignment
	// grants to Entra identities. The Redis analog of a custom role
	// definition: this says WHAT is allowed; the assignment says WHO gets
	// it.
	//
	// The three built-in policies ("Data Owner", "Data Contributor",
	// "Data Reader") need no policy resource -- assignments reference
	// them by name; a custom policy exists for finer grants (one key
	// prefix, no admin commands, single commands). Permissions are
	// updatable in place; the name and cache are fixed at creation. No
	// tags: ARM does not support tags on access policies (cache
	// children).
	createdAccessPolicy, err := redis.NewCacheAccessPolicy(ctx,
		spec.PolicyName,
		&redis.CacheAccessPolicyArgs{
			Name:         pulumi.String(spec.PolicyName),
			RedisCacheId: pulumi.String(locals.RedisCacheId),
			Permissions:  pulumi.String(spec.Permissions),
		},
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create redis cache access policy %s", spec.PolicyName)
	}

	// Export outputs. The name is what assignments reference to
	// grant this policy.
	ctx.Export(OpAccessPolicyId, createdAccessPolicy.ID())
	ctx.Export(OpAccessPolicyName, createdAccessPolicy.Name)

	return nil
}
