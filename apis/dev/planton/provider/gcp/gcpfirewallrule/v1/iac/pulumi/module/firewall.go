package module

import (
	"github.com/pkg/errors"
	gcpfirewallrulev1 "github.com/plantonhq/planton/apis/dev/planton/provider/gcp/gcpfirewallrule/v1"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/compute"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// firewall creates a GCP compute firewall rule.
// The spec's action field ("ALLOW" or "DENY") determines whether Pulumi's Allows or Denies
// argument is populated. Each GcpFirewallProtocolPort maps to one allow/deny block.
func firewall(ctx *pulumi.Context, locals *Locals, gcpProvider *gcp.Provider) error {
	spec := locals.GcpFirewallRule.Spec

	args := &compute.FirewallArgs{
		Name:      pulumi.String(spec.RuleName),
		Network:   pulumi.String(spec.Network.GetValue()),
		Project:   pulumi.String(spec.ProjectId.GetValue()),
		Direction: pulumi.String(spec.Direction),
		Disabled:  pulumi.BoolPtr(spec.Disabled),
	}

	// Priority (default applied by Planton middleware, always present).
	args.Priority = pulumi.IntPtr(int(spec.GetPriority()))

	// Description.
	if spec.Description != "" {
		args.Description = pulumi.StringPtr(spec.Description)
	}

	// Map action + rules to the correct Pulumi allow/deny blocks.
	switch spec.Action {
	case "ALLOW":
		args.Allows = mapToAllowRules(spec.Rules)
	case "DENY":
		args.Denies = mapToDenyRules(spec.Rules)
	}

	// Source ranges (INGRESS).
	if len(spec.SourceRanges) > 0 {
		args.SourceRanges = toPulumiStringArray(spec.SourceRanges)
	}

	// Destination ranges (EGRESS).
	if len(spec.DestinationRanges) > 0 {
		args.DestinationRanges = toPulumiStringArray(spec.DestinationRanges)
	}

	// Source tags.
	if len(spec.SourceTags) > 0 {
		args.SourceTags = toPulumiStringArray(spec.SourceTags)
	}

	// Target tags.
	if len(spec.TargetTags) > 0 {
		args.TargetTags = toPulumiStringArray(spec.TargetTags)
	}

	// Source service accounts.
	if len(spec.SourceServiceAccounts) > 0 {
		args.SourceServiceAccounts = toPulumiStringArray(spec.SourceServiceAccounts)
	}

	// Target service accounts.
	if len(spec.TargetServiceAccounts) > 0 {
		args.TargetServiceAccounts = toPulumiStringArray(spec.TargetServiceAccounts)
	}

	// Log config.
	if spec.LogConfig != nil {
		args.LogConfig = &compute.FirewallLogConfigArgs{
			Metadata: pulumi.String(spec.LogConfig.Metadata),
		}
	}

	createdFirewall, err := compute.NewFirewall(ctx, "firewall-rule", args, pulumi.Provider(gcpProvider))
	if err != nil {
		return errors.Wrap(err, "failed to create firewall rule")
	}

	ctx.Export(OpFirewallSelfLink, createdFirewall.SelfLink)
	ctx.Export(OpFirewallName, createdFirewall.Name)
	ctx.Export(OpCreationTimestamp, createdFirewall.CreationTimestamp)

	return nil
}

// mapToAllowRules converts the spec's GcpFirewallProtocolPort list to Pulumi FirewallAllowArray.
func mapToAllowRules(rules []*gcpfirewallrulev1.GcpFirewallProtocolPort) compute.FirewallAllowArray {
	var result compute.FirewallAllowArray
	for _, rule := range rules {
		result = append(result, &compute.FirewallAllowArgs{
			Protocol: pulumi.String(rule.Protocol),
			Ports:    toPulumiStringArray(rule.Ports),
		})
	}
	return result
}

// mapToDenyRules converts the spec's GcpFirewallProtocolPort list to Pulumi FirewallDenyArray.
func mapToDenyRules(rules []*gcpfirewallrulev1.GcpFirewallProtocolPort) compute.FirewallDenyArray {
	var result compute.FirewallDenyArray
	for _, rule := range rules {
		result = append(result, &compute.FirewallDenyArgs{
			Protocol: pulumi.String(rule.Protocol),
			Ports:    toPulumiStringArray(rule.Ports),
		})
	}
	return result
}

// toPulumiStringArray converts a Go string slice to a Pulumi StringArray.
func toPulumiStringArray(values []string) pulumi.StringArray {
	arr := make(pulumi.StringArray, len(values))
	for i, v := range values {
		arr[i] = pulumi.String(v)
	}
	return arr
}
