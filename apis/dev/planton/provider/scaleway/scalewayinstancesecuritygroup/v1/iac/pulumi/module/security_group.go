package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumiverse/pulumi-scaleway/sdk/go/scaleway"
	"github.com/pulumiverse/pulumi-scaleway/sdk/go/scaleway/instance"
)

// securityGroup provisions the Scaleway Instance Security Group with inline
// inbound and outbound rules, and exports the security_group_id output.
func securityGroup(
	ctx *pulumi.Context,
	locals *Locals,
	scalewayProvider *scaleway.Provider,
) error {
	spec := locals.ScalewayInstanceSecurityGroup.Spec

	// ── 1. Translate inbound rules ────────────────────────────────────────
	//
	// Map proto InboundRule messages to Pulumi SecurityGroupInboundRuleArgs.
	// We use PortRange (string) as the unified port field -- it handles both
	// single ports ("80") and ranges ("22-23").
	inboundRules := make(instance.SecurityGroupInboundRuleArray, 0, len(spec.InboundRules))
	for _, rule := range spec.InboundRules {
		ruleArgs := instance.SecurityGroupInboundRuleArgs{
			Action: pulumi.String(rule.Action),
		}

		if rule.Protocol != "" {
			ruleArgs.Protocol = pulumi.String(rule.Protocol)
		}

		if rule.PortRange != "" {
			ruleArgs.PortRange = pulumi.String(rule.PortRange)
		}

		if rule.IpRange != "" {
			ruleArgs.IpRange = pulumi.String(rule.IpRange)
		}

		inboundRules = append(inboundRules, ruleArgs)
	}

	// ── 2. Translate outbound rules ───────────────────────────────────────
	//
	// Same mapping logic as inbound rules.
	outboundRules := make(instance.SecurityGroupOutboundRuleArray, 0, len(spec.OutboundRules))
	for _, rule := range spec.OutboundRules {
		ruleArgs := instance.SecurityGroupOutboundRuleArgs{
			Action: pulumi.String(rule.Action),
		}

		if rule.Protocol != "" {
			ruleArgs.Protocol = pulumi.String(rule.Protocol)
		}

		if rule.PortRange != "" {
			ruleArgs.PortRange = pulumi.String(rule.PortRange)
		}

		if rule.IpRange != "" {
			ruleArgs.IpRange = pulumi.String(rule.IpRange)
		}

		outboundRules = append(outboundRules, ruleArgs)
	}

	// ── 3. Build the security group arguments ─────────────────────────────
	sgArgs := &instance.SecurityGroupArgs{
		Name: pulumi.String(locals.ScalewayInstanceSecurityGroup.Metadata.Name),
		Zone: pulumi.String(spec.Zone),
		Tags: pulumi.ToStringArray(locals.ScalewayTags),
	}

	// Description (optional).
	if spec.Description != "" {
		sgArgs.Description = pulumi.String(spec.Description)
	}

	// Stateful flag. The Scaleway default is true; we set it explicitly
	// so the user's intent is always reflected in the Pulumi state.
	sgArgs.Stateful = pulumi.Bool(spec.Stateful)

	// Default policies. Only set when the user specified a non-empty value
	// to avoid overriding Scaleway's defaults with empty strings.
	if spec.InboundDefaultPolicy != "" {
		sgArgs.InboundDefaultPolicy = pulumi.String(spec.InboundDefaultPolicy)
	}
	if spec.OutboundDefaultPolicy != "" {
		sgArgs.OutboundDefaultPolicy = pulumi.String(spec.OutboundDefaultPolicy)
	}

	// SMTP security (enable_default_security).
	sgArgs.EnableDefaultSecurity = pulumi.Bool(spec.EnableDefaultSecurity)

	// Inline rules.
	if len(inboundRules) > 0 {
		sgArgs.InboundRules = inboundRules
	}
	if len(outboundRules) > 0 {
		sgArgs.OutboundRules = outboundRules
	}

	// ── 4. Create the security group ──────────────────────────────────────
	createdSg, err := instance.NewSecurityGroup(
		ctx,
		"security-group",
		sgArgs,
		pulumi.Provider(scalewayProvider),
	)
	if err != nil {
		return errors.Wrap(err, "failed to create scaleway instance security group")
	}

	// ── 5. Export stack output ─────────────────────────────────────────────
	ctx.Export(OpSecurityGroupId, createdSg.ID())

	return nil
}
