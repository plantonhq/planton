package module

import (
	"fmt"

	"github.com/pkg/errors"
	gcphierarchicalfirewallpolicyv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcphierarchicalfirewallpolicy/v1alpha1"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/compute"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// firewallPolicyRules provisions one rule resource per spec.rules entry.
// A rule is keyed by its priority in Google and in this module (the
// resource name carries it), so changing a rule's priority recreates that
// one rule while every other rule is untouched -- the provider's own
// semantics, mirrored by the Terraform module's for_each key.
//
// Optional inputs are sent only when set so the provider's defaults stay
// the provider's: the two Optional+Computed network-context enums would
// otherwise fight the API's read-back on re-plan. Booleans are sent as the
// spec states them (false is the API's own default and reads back as
// such).
func firewallPolicyRules(ctx *pulumi.Context, locals *Locals, gcpProvider *gcp.Provider,
	createdPolicy *compute.FirewallPolicy) error {
	spec := locals.GcpHierarchicalFirewallPolicy.Spec

	for _, rule := range spec.Rules {
		priority := int(rule.GetPriority())
		args := &compute.FirewallPolicyRuleArgs{
			FirewallPolicy: createdPolicy.Name,
			Priority:       pulumi.Int(priority),
			Action:         pulumi.String(rule.Action),
			Direction:      pulumi.String(rule.Direction),
			Disabled:       pulumi.BoolPtr(rule.Disabled),
			EnableLogging:  pulumi.BoolPtr(rule.EnableLogging),
			Match:          ruleMatchArgs(rule.Match),
		}
		if rule.Description != "" {
			args.Description = pulumi.StringPtr(rule.Description)
		}
		if targets := refListValues(rule.TargetResources); len(targets) > 0 {
			args.TargetResources = targets
		}
		if tags := secureTagArgs(rule.TargetSecureTags); len(tags) > 0 {
			args.TargetSecureTags = tags
		}
		if accounts := refListValues(rule.TargetServiceAccounts); len(accounts) > 0 {
			args.TargetServiceAccounts = accounts
		}
		// Present exactly when action is apply_security_profile_group
		// (spec CEL); tls_inspect rides along with that action only.
		if rule.SecurityProfileGroup != "" {
			args.SecurityProfileGroup = pulumi.StringPtr(rule.SecurityProfileGroup)
			args.TlsInspect = pulumi.BoolPtr(rule.TlsInspect)
		}
		if spec.DeletionPolicy != "" {
			args.DeletionPolicy = pulumi.StringPtr(spec.DeletionPolicy)
		}

		if _, err := compute.NewFirewallPolicyRule(ctx, fmt.Sprintf("rule-%d", priority), args,
			pulumi.Provider(gcpProvider), pulumi.Parent(createdPolicy)); err != nil {
			return errors.Wrapf(err, "failed to create rule at priority %d", priority)
		}
	}

	return nil
}

// ruleMatchArgs maps the spec's match message onto the provider's block.
// Every list is sent only when non-empty; the two network-context enums
// only when set (Optional+Computed).
func ruleMatchArgs(match *gcphierarchicalfirewallpolicyv1alpha1.GcpHierarchicalFirewallPolicyRuleMatch) *compute.FirewallPolicyRuleMatchArgs {
	layer4 := compute.FirewallPolicyRuleMatchLayer4ConfigArray{}
	for _, config := range match.Layer4Configs {
		configArgs := &compute.FirewallPolicyRuleMatchLayer4ConfigArgs{
			IpProtocol: pulumi.String(config.IpProtocol),
		}
		if len(config.Ports) > 0 {
			configArgs.Ports = pulumi.ToStringArray(config.Ports)
		}
		layer4 = append(layer4, configArgs)
	}

	args := &compute.FirewallPolicyRuleMatchArgs{
		Layer4Configs: layer4,
	}
	if len(match.SrcIpRanges) > 0 {
		args.SrcIpRanges = pulumi.ToStringArray(match.SrcIpRanges)
	}
	if len(match.DestIpRanges) > 0 {
		args.DestIpRanges = pulumi.ToStringArray(match.DestIpRanges)
	}
	if len(match.SrcAddressGroups) > 0 {
		args.SrcAddressGroups = pulumi.ToStringArray(match.SrcAddressGroups)
	}
	if len(match.DestAddressGroups) > 0 {
		args.DestAddressGroups = pulumi.ToStringArray(match.DestAddressGroups)
	}
	if len(match.SrcFqdns) > 0 {
		args.SrcFqdns = pulumi.ToStringArray(match.SrcFqdns)
	}
	if len(match.DestFqdns) > 0 {
		args.DestFqdns = pulumi.ToStringArray(match.DestFqdns)
	}
	if len(match.SrcRegionCodes) > 0 {
		args.SrcRegionCodes = pulumi.ToStringArray(match.SrcRegionCodes)
	}
	if len(match.DestRegionCodes) > 0 {
		args.DestRegionCodes = pulumi.ToStringArray(match.DestRegionCodes)
	}
	if len(match.SrcThreatIntelligences) > 0 {
		args.SrcThreatIntelligences = pulumi.ToStringArray(match.SrcThreatIntelligences)
	}
	if len(match.DestThreatIntelligences) > 0 {
		args.DestThreatIntelligences = pulumi.ToStringArray(match.DestThreatIntelligences)
	}
	if tags := matchSecureTagArgs(match.SrcSecureTags); len(tags) > 0 {
		args.SrcSecureTags = tags
	}
	if networks := refListValues(match.SrcNetworks); len(networks) > 0 {
		args.SrcNetworks = networks
	}
	if match.SrcNetworkContext != nil && match.GetSrcNetworkContext() != "" {
		args.SrcNetworkContext = pulumi.StringPtr(match.GetSrcNetworkContext())
	}
	if match.DestNetworkContext != nil && match.GetDestNetworkContext() != "" {
		args.DestNetworkContext = pulumi.StringPtr(match.GetDestNetworkContext())
	}
	return args
}

// secureTagArgs renders the rule's target secure tags: each reference
// resolves to a tag value's `tagValues/{id}` name, which is the one input
// of the provider's block (its `state` is read-only).
func secureTagArgs(refs []*foreignkeyv1.StringValueOrRef) compute.FirewallPolicyRuleTargetSecureTagArray {
	tags := compute.FirewallPolicyRuleTargetSecureTagArray{}
	for _, name := range refListValues(refs) {
		tags = append(tags, &compute.FirewallPolicyRuleTargetSecureTagArgs{Name: name})
	}
	return tags
}

// matchSecureTagArgs is secureTagArgs for the match block's source tags,
// which the provider types separately.
func matchSecureTagArgs(refs []*foreignkeyv1.StringValueOrRef) compute.FirewallPolicyRuleMatchSrcSecureTagArray {
	tags := compute.FirewallPolicyRuleMatchSrcSecureTagArray{}
	for _, name := range refListValues(refs) {
		tags = append(tags, &compute.FirewallPolicyRuleMatchSrcSecureTagArgs{Name: name})
	}
	return tags
}

// refListValues flattens a list of references to their resolved values,
// dropping empties (the manifest loader has already resolved every
// valueFrom by the time the module runs).
func refListValues(refs []*foreignkeyv1.StringValueOrRef) pulumi.StringArray {
	values := pulumi.StringArray{}
	for _, ref := range refs {
		if ref.GetValue() != "" {
			values = append(values, pulumi.String(ref.GetValue()))
		}
	}
	return values
}
