package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-oci/sdk/v4/go/oci"
	"github.com/pulumi/pulumi-oci/sdk/v4/go/oci/networkfirewall"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func policyResource(ctx *pulumi.Context, locals *Locals, provider *oci.Provider) (*networkfirewall.NetworkFirewallPolicy, error) {
	spec := locals.OciNetworkFirewall.Spec

	args := &networkfirewall.NetworkFirewallPolicyArgs{
		CompartmentId: pulumi.String(spec.CompartmentId.GetValue()),
		DisplayName:   pulumi.String(locals.PolicyDisplayName),
		FreeformTags:  pulumi.ToStringMap(locals.FreeformTags),
	}

	if spec.Policy.Description != "" {
		args.Description = pulumi.String(spec.Policy.Description)
	}

	createdPolicy, err := networkfirewall.NewNetworkFirewallPolicy(
		ctx,
		locals.PolicyDisplayName,
		args,
		pulumiOciOpt(provider),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create firewall policy")
	}

	ctx.Export(OpPolicyId, createdPolicy.ID())

	return createdPolicy, nil
}
