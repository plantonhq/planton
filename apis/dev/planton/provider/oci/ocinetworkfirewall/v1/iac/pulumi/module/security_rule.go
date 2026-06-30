package module

import (
	"fmt"

	"github.com/pkg/errors"
	ocinetworkfirewallv1 "github.com/plantonhq/planton/apis/dev/planton/provider/oci/ocinetworkfirewall/v1"
	"github.com/pulumi/pulumi-oci/sdk/v4/go/oci"
	"github.com/pulumi/pulumi-oci/sdk/v4/go/oci/networkfirewall"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func securityRuleResources(
	ctx *pulumi.Context,
	locals *Locals,
	provider *oci.Provider,
	policy *networkfirewall.NetworkFirewallPolicy,
	subResourceDeps []pulumi.Resource,
) ([]pulumi.Resource, error) {
	spec := locals.OciNetworkFirewall.Spec
	var resources []pulumi.Resource

	deps := []pulumi.Resource{policy}
	deps = append(deps, subResourceDeps...)

	for i, rule := range spec.Policy.SecurityRules {
		args := &networkfirewall.NetworkFirewallPolicySecurityRuleArgs{
			NetworkFirewallPolicyId: policy.ID(),
			Name:                    pulumi.String(rule.Name),
			Action:                  pulumi.String(actionMap[rule.Action]),
			Condition:               buildSecurityRuleCondition(rule.Condition),
			PriorityOrder:           pulumi.String(fmt.Sprintf("%d", i+1)),
		}

		if rule.Inspection != ocinetworkfirewallv1.OciNetworkFirewallSpec_SecurityRule_inspection_unspecified {
			args.Inspection = pulumi.String(inspectionMap[rule.Inspection])
		}

		if rule.Description != "" {
			args.Description = pulumi.String(rule.Description)
		}

		resourceName := fmt.Sprintf("security-rule-%s", rule.Name)
		created, err := networkfirewall.NewNetworkFirewallPolicySecurityRule(
			ctx,
			resourceName,
			args,
			pulumiOciOpt(provider),
			pulumi.DependsOn(deps),
		)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to create security rule %s", rule.Name)
		}

		resources = append(resources, created)
	}

	return resources, nil
}

func buildSecurityRuleCondition(cond *ocinetworkfirewallv1.OciNetworkFirewallSpec_SecurityRuleCondition) *networkfirewall.NetworkFirewallPolicySecurityRuleConditionArgs {
	args := &networkfirewall.NetworkFirewallPolicySecurityRuleConditionArgs{}

	if len(cond.SourceAddresses) > 0 {
		args.SourceAddresses = pulumi.ToStringArray(cond.SourceAddresses)
	}

	if len(cond.DestinationAddresses) > 0 {
		args.DestinationAddresses = pulumi.ToStringArray(cond.DestinationAddresses)
	}

	if len(cond.Services) > 0 {
		args.Services = pulumi.ToStringArray(cond.Services)
	}

	if len(cond.Urls) > 0 {
		args.Urls = pulumi.ToStringArray(cond.Urls)
	}

	return args
}
