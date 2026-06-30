package module

import (
	"fmt"

	"github.com/pkg/errors"
	ocinetworkfirewallv1 "github.com/plantonhq/planton/apis/dev/planton/provider/oci/ocinetworkfirewall/v1"
	"github.com/pulumi/pulumi-oci/sdk/v4/go/oci"
	"github.com/pulumi/pulumi-oci/sdk/v4/go/oci/networkfirewall"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func urlListResources(
	ctx *pulumi.Context,
	locals *Locals,
	provider *oci.Provider,
	policy *networkfirewall.NetworkFirewallPolicy,
) ([]pulumi.Resource, error) {
	spec := locals.OciNetworkFirewall.Spec
	var resources []pulumi.Resource

	for _, ul := range spec.Policy.UrlLists {
		args := &networkfirewall.NetworkFirewallPolicyUrlListArgs{
			NetworkFirewallPolicyId: policy.ID(),
			Name:                    pulumi.String(ul.Name),
			Urls:                    buildUrls(ul.Urls),
		}

		if ul.Description != "" {
			args.Description = pulumi.String(ul.Description)
		}

		resourceName := fmt.Sprintf("url-list-%s", ul.Name)
		created, err := networkfirewall.NewNetworkFirewallPolicyUrlList(
			ctx,
			resourceName,
			args,
			pulumiOciOpt(provider),
			pulumi.DependsOn([]pulumi.Resource{policy}),
		)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to create url list %s", ul.Name)
		}

		resources = append(resources, created)
	}

	return resources, nil
}

func buildUrls(urls []*ocinetworkfirewallv1.OciNetworkFirewallSpec_UrlPattern) networkfirewall.NetworkFirewallPolicyUrlListUrlArray {
	result := make(networkfirewall.NetworkFirewallPolicyUrlListUrlArray, len(urls))

	for i, u := range urls {
		result[i] = &networkfirewall.NetworkFirewallPolicyUrlListUrlArgs{
			Pattern: pulumi.String(u.Pattern),
			Type:    pulumi.String("SIMPLE"),
		}
	}

	return result
}
