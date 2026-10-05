package module

import (
	"fmt"

	"github.com/pkg/errors"
	azureprivatednsresolverforwardingrulesetv1alpha1 "github.com/plantonhq/planton/catalog/azure/azureprivatednsresolverforwardingruleset/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/pulumiazureprovider"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/privatedns"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *azureprivatednsresolverforwardingrulesetv1alpha1.AzurePrivateDnsResolverForwardingRulesetIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the Azure provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static client secret, keyless web identity, or ambient chain).
	azureProvider, err := pulumiazureprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzurePrivateDnsResolverForwardingRuleset.Spec

	// The ruleset binds the outbound endpoint(s) of a DNS Private
	// Resolver (at most 2, both from the SAME resolver -- Azure enforces
	// it at deploy time). The endpoint list and tags update in place;
	// name, resource group, and region replace the ruleset.
	createdRuleset, err := privatedns.NewResolverDnsForwardingRuleset(ctx,
		spec.Name,
		&privatedns.ResolverDnsForwardingRulesetArgs{
			Name:                                  pulumi.String(spec.Name),
			ResourceGroupName:                     pulumi.String(locals.ResourceGroupName),
			Location:                              pulumi.String(spec.Region),
			PrivateDnsResolverOutboundEndpointIds: pulumi.ToStringArray(locals.OutboundEndpointIds),
			Tags:                                  pulumi.ToStringMap(locals.AzureTags),
		},
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create dns forwarding ruleset %s", spec.Name)
	}

	// The forwarding rules -- one per captured domain. Everything on a
	// rule updates in place EXCEPT domain_name, which replaces that rule.
	for _, rule := range spec.ForwardingRules {
		targetDnsServers := privatedns.ResolverForwardingRuleTargetDnsServerArray{}
		for _, server := range rule.TargetDnsServers {
			// The port is always sent explicitly -- 53 (the standard DNS
			// port and ARM's default) when the spec leaves it unset -- so
			// both engines send identical wire shapes.
			port := 53
			if server.Port != nil {
				port = int(*server.Port)
			}
			targetDnsServers = append(targetDnsServers, &privatedns.ResolverForwardingRuleTargetDnsServerArgs{
				IpAddress: pulumi.String(server.IpAddress),
				Port:      pulumi.Int(port),
			})
		}

		enabled := true
		if rule.Enabled != nil {
			enabled = *rule.Enabled
		}

		ruleArgs := &privatedns.ResolverForwardingRuleArgs{
			Name:                   pulumi.String(rule.Name),
			DnsForwardingRulesetId: createdRuleset.ID(),
			// ARM stores domains as fully qualified names WITH the
			// trailing dot ("corp.contoso.com.") -- write them that way
			// in the spec.
			DomainName:       pulumi.String(rule.DomainName),
			Enabled:          pulumi.Bool(enabled),
			TargetDnsServers: targetDnsServers,
		}
		// ARM's free-form annotation map on the rule itself (rules carry
		// no tags).
		if len(rule.Metadata) > 0 {
			ruleArgs.Metadata = pulumi.ToStringMap(rule.Metadata)
		}

		if _, err := privatedns.NewResolverForwardingRule(ctx,
			fmt.Sprintf("%s-rule-%s", spec.Name, rule.Name),
			ruleArgs,
			pulumi.Provider(azureProvider)); err != nil {
			return errors.Wrapf(err, "failed to create forwarding rule %s", rule.Name)
		}
	}

	ctx.Export(OpDnsForwardingRulesetId, createdRuleset.ID())
	ctx.Export(OpDnsForwardingRulesetName, createdRuleset.Name)

	return nil
}
