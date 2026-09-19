package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/compute"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// firewallPolicyAssociations provisions one association resource per
// spec.associations entry, attaching the policy to the organization or a
// folder so every network beneath that node enforces it. Keyed by the
// association's name (the resource name carries it); name and target are
// immutable in Google, so a change to either detaches the old association
// and creates the new one.
func firewallPolicyAssociations(ctx *pulumi.Context, locals *Locals, gcpProvider *gcp.Provider,
	createdPolicy *compute.FirewallPolicy) error {
	spec := locals.GcpHierarchicalFirewallPolicy.Spec

	for i := range spec.Associations {
		name := locals.AssociationNames[i]
		args := &compute.FirewallPolicyAssociationArgs{
			FirewallPolicy:   createdPolicy.Name,
			Name:             pulumi.String(name),
			AttachmentTarget: pulumi.String(locals.AssociationTargets[i]),
		}
		if spec.DeletionPolicy != "" {
			args.DeletionPolicy = pulumi.StringPtr(spec.DeletionPolicy)
		}

		if _, err := compute.NewFirewallPolicyAssociation(ctx, "association-"+name, args,
			pulumi.Provider(gcpProvider), pulumi.Parent(createdPolicy)); err != nil {
			return errors.Wrapf(err, "failed to create association %s", name)
		}
	}

	return nil
}
