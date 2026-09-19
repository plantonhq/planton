package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/compute"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// firewallPolicy provisions the hierarchical firewall policy container:
// the organization- or folder-owned object the rules and associations
// hang off. Google names it with a server-assigned numeric ID (the
// policy's `name`), which is what the rules and associations reference and
// what the policy_id output carries.
//
// parent and short_name are immutable: changing either recreates the
// policy, and with it every rule and association. description changes in
// place. deletion_policy is sent only when set so the provider's default
// (DELETE) stays the provider's.
func firewallPolicy(ctx *pulumi.Context, locals *Locals, gcpProvider *gcp.Provider) (*compute.FirewallPolicy, error) {
	spec := locals.GcpHierarchicalFirewallPolicy.Spec

	args := &compute.FirewallPolicyArgs{
		Parent:    pulumi.String(locals.Parent),
		ShortName: pulumi.String(locals.ShortName),
	}
	if spec.Description != "" {
		args.Description = pulumi.StringPtr(spec.Description)
	}
	if spec.DeletionPolicy != "" {
		args.DeletionPolicy = pulumi.StringPtr(spec.DeletionPolicy)
	}

	createdPolicy, err := compute.NewFirewallPolicy(ctx, "policy", args, pulumi.Provider(gcpProvider))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create firewall policy")
	}

	ctx.Export(OpPolicyId, createdPolicy.Name)
	ctx.Export(OpShortName, createdPolicy.ShortName)
	ctx.Export(OpSelfLink, createdPolicy.SelfLink)
	ctx.Export(OpParent, createdPolicy.Parent)
	ctx.Export(OpRuleTupleCount, createdPolicy.RuleTupleCount)
	ctx.Export(OpAssociationNames, pulumi.ToStringArray(locals.AssociationNames))

	return createdPolicy, nil
}
