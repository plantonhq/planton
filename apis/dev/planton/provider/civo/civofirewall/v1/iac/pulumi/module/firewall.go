package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-civo/sdk/v2/go/civo"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// firewall provisions the firewall and exports its ID.
func firewall(
	ctx *pulumi.Context,
	locals *Locals,
	civoProvider *civo.Provider,
) (*civo.Firewall, error) {

	// 1. Translate inbound rules.
	inboundRules := make(civo.FirewallIngressRuleArray, 0, len(locals.CivoFirewall.Spec.InboundRules))
	for _, rule := range locals.CivoFirewall.Spec.InboundRules {
		inboundRules = append(inboundRules, civo.FirewallIngressRuleArgs{
			Protocol:  pulumi.String(rule.Protocol),
			PortRange: pulumi.String(rule.PortRange),
			Cidrs:     pulumi.ToStringArray(rule.Cidrs),
			Action:    pulumi.String(rule.Action),
			Label:     pulumi.String(rule.Label),
		})
	}

	// 2. Translate outbound rules.
	outboundRules := make(civo.FirewallEgressRuleArray, 0, len(locals.CivoFirewall.Spec.OutboundRules))
	for _, rule := range locals.CivoFirewall.Spec.OutboundRules {
		outboundRules = append(outboundRules, civo.FirewallEgressRuleArgs{
			Protocol:  pulumi.String(rule.Protocol),
			PortRange: pulumi.String(rule.PortRange),
			Cidrs:     pulumi.ToStringArray(rule.Cidrs),
			Action:    pulumi.String(rule.Action),
			Label:     pulumi.String(rule.Label),
		})
	}

	// 3. Build firewall args.
	firewallArgs := &civo.FirewallArgs{
		Name:         pulumi.String(locals.CivoFirewall.Spec.Name),
		NetworkId:    pulumi.String(locals.CivoFirewall.Spec.NetworkId.GetValue()),
		IngressRules: inboundRules,
		EgressRules:  outboundRules,
	}

	// 4. Create the firewall.
	createdFirewall, err := civo.NewFirewall(
		ctx,
		"firewall",
		firewallArgs,
		pulumi.Provider(civoProvider),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create Civo firewall")
	}

	// 5. Export output.
	ctx.Export(OpFirewallId, createdFirewall.ID())

	return createdFirewall, nil
}
