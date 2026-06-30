package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-digitalocean/sdk/v4/go/digitalocean"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// firewall provisions the firewall and exports its ID.
func firewall(
	ctx *pulumi.Context,
	locals *Locals,
	digitalOceanProvider *digitalocean.Provider,
) (*digitalocean.Firewall, error) {

	// 1. Translate inbound rules.
	inboundRules := make(digitalocean.FirewallInboundRuleArray, 0, len(locals.DigitalOceanFirewall.Spec.InboundRules))
	for _, rule := range locals.DigitalOceanFirewall.Spec.InboundRules {
		inboundRules = append(inboundRules, digitalocean.FirewallInboundRuleArgs{
			Protocol:               pulumi.String(rule.Protocol),
			PortRange:              pulumi.String(rule.PortRange),
			SourceAddresses:        pulumi.ToStringArray(rule.SourceAddresses),
			SourceDropletIds:       int64SliceToIntArray(rule.SourceDropletIds),
			SourceTags:             pulumi.ToStringArray(rule.SourceTags),
			SourceKubernetesIds:    pulumi.ToStringArray(rule.SourceKubernetesIds),
			SourceLoadBalancerUids: pulumi.ToStringArray(rule.SourceLoadBalancerUids),
		})
	}

	// 2. Translate outbound rules.
	outboundRules := make(digitalocean.FirewallOutboundRuleArray, 0, len(locals.DigitalOceanFirewall.Spec.OutboundRules))
	for _, rule := range locals.DigitalOceanFirewall.Spec.OutboundRules {
		outboundRules = append(outboundRules, digitalocean.FirewallOutboundRuleArgs{
			Protocol:                    pulumi.String(rule.Protocol),
			PortRange:                   pulumi.String(rule.PortRange),
			DestinationAddresses:        pulumi.ToStringArray(rule.DestinationAddresses),
			DestinationDropletIds:       int64SliceToIntArray(rule.DestinationDropletIds),
			DestinationTags:             pulumi.ToStringArray(rule.DestinationTags),
			DestinationKubernetesIds:    pulumi.ToStringArray(rule.DestinationKubernetesIds),
			DestinationLoadBalancerUids: pulumi.ToStringArray(rule.DestinationLoadBalancerUids),
		})
	}

	// 3. Build firewall args.
	firewallArgs := &digitalocean.FirewallArgs{
		Name:          pulumi.String(locals.DigitalOceanFirewall.Spec.Name),
		InboundRules:  inboundRules,
		OutboundRules: outboundRules,
		DropletIds:    int64SliceToIntArray(locals.DigitalOceanFirewall.Spec.DropletIds),
		Tags:          pulumi.ToStringArray(locals.DigitalOceanFirewall.Spec.Tags),
	}

	// 4. Create the firewall using the provider.
	createdFirewall, err := digitalocean.NewFirewall(
		ctx,
		"firewall",
		firewallArgs,
		pulumi.Provider(digitalOceanProvider),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create digitalocean firewall")
	}

	// 5. Export output.
	ctx.Export(OpFirewallId, createdFirewall.ID())

	return createdFirewall, nil
}

// int64SliceToIntArray converts []int64 → pulumi.IntArrayInput.
func int64SliceToIntArray(values []int64) pulumi.IntArray {
	intInputs := make(pulumi.IntArray, 0, len(values))
	for _, v := range values {
		intInputs = append(intInputs, pulumi.Int(int(v)))
	}
	return intInputs
}
