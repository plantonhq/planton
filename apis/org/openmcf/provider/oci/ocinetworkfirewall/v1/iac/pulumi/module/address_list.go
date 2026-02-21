package module

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-oci/sdk/v4/go/oci"
	"github.com/pulumi/pulumi-oci/sdk/v4/go/oci/networkfirewall"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func addressListResources(
	ctx *pulumi.Context,
	locals *Locals,
	provider *oci.Provider,
	policy *networkfirewall.NetworkFirewallPolicy,
) ([]pulumi.Resource, error) {
	spec := locals.OciNetworkFirewall.Spec
	var resources []pulumi.Resource

	for _, al := range spec.Policy.AddressLists {
		args := &networkfirewall.NetworkFirewallPolicyAddressListArgs{
			NetworkFirewallPolicyId: policy.ID(),
			Name:                    pulumi.String(al.Name),
			Type:                    pulumi.String(addressListTypeMap[al.Type]),
			Addresses:               pulumi.ToStringArray(al.Addresses),
		}

		if al.Description != "" {
			args.Description = pulumi.String(al.Description)
		}

		resourceName := fmt.Sprintf("address-list-%s", al.Name)
		created, err := networkfirewall.NewNetworkFirewallPolicyAddressList(
			ctx,
			resourceName,
			args,
			pulumiOciOpt(provider),
			pulumi.DependsOn([]pulumi.Resource{policy}),
		)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to create address list %s", al.Name)
		}

		resources = append(resources, created)
	}

	return resources, nil
}
