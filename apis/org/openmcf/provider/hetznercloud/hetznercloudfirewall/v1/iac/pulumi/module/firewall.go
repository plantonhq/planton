package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-hcloud/sdk/go/hcloud"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// firewall provisions the Hetzner Cloud firewall with inline rules and exports its ID.
func firewall(
	ctx *pulumi.Context,
	locals *Locals,
	hcloudProvider *hcloud.Provider,
) error {
	var rules hcloud.FirewallRuleArray

	for _, rule := range locals.HetznerCloudFirewall.Spec.Rules {
		args := &hcloud.FirewallRuleArgs{
			Direction: pulumi.String(rule.Direction.String()),
			Protocol:  pulumi.String(rule.Protocol.String()),
		}

		if rule.Port != "" {
			args.Port = pulumi.String(rule.Port)
		}

		if len(rule.SourceIps) > 0 {
			args.SourceIps = pulumi.ToStringArray(rule.SourceIps)
		}

		if len(rule.DestinationIps) > 0 {
			args.DestinationIps = pulumi.ToStringArray(rule.DestinationIps)
		}

		if rule.Description != "" {
			args.Description = pulumi.String(rule.Description)
		}

		rules = append(rules, args)
	}

	createdFirewall, err := hcloud.NewFirewall(
		ctx,
		"firewall",
		&hcloud.FirewallArgs{
			Name:   pulumi.String(locals.HetznerCloudFirewall.Metadata.Name),
			Labels: pulumi.ToStringMap(locals.Labels),
			Rules:  rules,
		},
		pulumi.Provider(hcloudProvider),
	)
	if err != nil {
		return errors.Wrap(err, "failed to create hetzner cloud firewall")
	}

	ctx.Export(OpFirewallId, createdFirewall.ID())

	return nil
}
