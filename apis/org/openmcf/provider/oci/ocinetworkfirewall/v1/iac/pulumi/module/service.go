package module

import (
	"fmt"

	"github.com/pkg/errors"
	ocinetworkfirewallv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/oci/ocinetworkfirewall/v1"
	"github.com/pulumi/pulumi-oci/sdk/v4/go/oci"
	"github.com/pulumi/pulumi-oci/sdk/v4/go/oci/networkfirewall"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func serviceResources(
	ctx *pulumi.Context,
	locals *Locals,
	provider *oci.Provider,
	policy *networkfirewall.NetworkFirewallPolicy,
) ([]pulumi.Resource, error) {
	spec := locals.OciNetworkFirewall.Spec
	var resources []pulumi.Resource

	for _, svc := range spec.Policy.Services {
		args := &networkfirewall.NetworkFirewallPolicyServiceArgs{
			NetworkFirewallPolicyId: policy.ID(),
			Name:                   pulumi.String(svc.Name),
			Type:                   pulumi.String(serviceTypeMap[svc.Type]),
			PortRanges:             buildPortRanges(svc.PortRanges),
		}

		if svc.Description != "" {
			args.Description = pulumi.String(svc.Description)
		}

		resourceName := fmt.Sprintf("service-%s", svc.Name)
		created, err := networkfirewall.NewNetworkFirewallPolicyService(
			ctx,
			resourceName,
			args,
			pulumiOciOpt(provider),
			pulumi.DependsOn([]pulumi.Resource{policy}),
		)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to create service %s", svc.Name)
		}

		resources = append(resources, created)
	}

	return resources, nil
}

func serviceListResources(
	ctx *pulumi.Context,
	locals *Locals,
	provider *oci.Provider,
	policy *networkfirewall.NetworkFirewallPolicy,
	serviceDeps []pulumi.Resource,
) ([]pulumi.Resource, error) {
	spec := locals.OciNetworkFirewall.Spec
	var resources []pulumi.Resource

	deps := []pulumi.Resource{policy}
	deps = append(deps, serviceDeps...)

	for _, sl := range spec.Policy.ServiceLists {
		args := &networkfirewall.NetworkFirewallPolicyServiceListArgs{
			NetworkFirewallPolicyId: policy.ID(),
			Name:                   pulumi.String(sl.Name),
			Services:               pulumi.ToStringArray(sl.Services),
		}

		if sl.Description != "" {
			args.Description = pulumi.String(sl.Description)
		}

		resourceName := fmt.Sprintf("service-list-%s", sl.Name)
		created, err := networkfirewall.NewNetworkFirewallPolicyServiceList(
			ctx,
			resourceName,
			args,
			pulumiOciOpt(provider),
			pulumi.DependsOn(deps),
		)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to create service list %s", sl.Name)
		}

		resources = append(resources, created)
	}

	return resources, nil
}

func buildPortRanges(portRanges []*ocinetworkfirewallv1.OciNetworkFirewallSpec_PortRange) networkfirewall.NetworkFirewallPolicyServicePortRangeArray {
	result := make(networkfirewall.NetworkFirewallPolicyServicePortRangeArray, len(portRanges))

	for i, pr := range portRanges {
		prArgs := &networkfirewall.NetworkFirewallPolicyServicePortRangeArgs{
			MinimumPort: pulumi.Int(int(pr.MinimumPort)),
		}

		if pr.MaximumPort != nil {
			prArgs.MaximumPort = pulumi.Int(int(*pr.MaximumPort))
		}

		result[i] = prArgs
	}

	return result
}
