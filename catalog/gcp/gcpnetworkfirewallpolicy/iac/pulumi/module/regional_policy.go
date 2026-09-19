package module

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/compute"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// regionalPolicy provisions the REGIONAL resource family -- the same
// shape as the global one plus `region` on every resource. See
// globalPolicy for the send posture; nothing else differs.
func regionalPolicy(ctx *pulumi.Context, locals *Locals, gcpProvider *gcp.Provider, computeApiService pulumi.Resource) error {
	spec := locals.GcpNetworkFirewallPolicy.Spec
	region := pulumi.String(spec.Region)

	policyArgs := &compute.RegionNetworkFirewallPolicyArgs{
		Name:   pulumi.String(locals.PolicyName),
		Region: region,
	}
	if locals.ProjectId != "" {
		policyArgs.Project = pulumi.String(locals.ProjectId)
	}
	if spec.Description != "" {
		policyArgs.Description = pulumi.StringPtr(spec.Description)
	}
	if spec.PolicyType != nil && spec.GetPolicyType() != "" {
		policyArgs.PolicyType = pulumi.StringPtr(spec.GetPolicyType())
	}
	if spec.DeletionPolicy != "" {
		policyArgs.DeletionPolicy = pulumi.StringPtr(spec.DeletionPolicy)
	}

	createdPolicy, err := compute.NewRegionNetworkFirewallPolicy(ctx, "policy", policyArgs,
		pulumi.Provider(gcpProvider), pulumi.DependsOn([]pulumi.Resource{computeApiService}))
	if err != nil {
		return errors.Wrap(err, "failed to create region network firewall policy")
	}

	for _, rule := range spec.Rules {
		in := renderRule(rule)
		ruleArgs := &compute.RegionNetworkFirewallPolicyRuleArgs{
			FirewallPolicy:        createdPolicy.Name,
			Region:                region,
			Priority:              pulumi.Int(in.Priority),
			Action:                in.Action,
			Direction:             in.Direction,
			Description:           in.Description,
			RuleName:              in.RuleName,
			Disabled:              in.Disabled,
			EnableLogging:         in.EnableLogging,
			TargetType:            in.TargetType,
			SecurityProfileGroup:  in.SecurityProfileGroup,
			TlsInspect:            in.TlsInspect,
			TargetServiceAccounts: in.TargetServiceAccounts,
			TargetForwardingRules: in.TargetForwardingRules,
			Match:                 regionalMatchArgs(in.Match),
		}
		if len(in.TargetSecureTags) > 0 {
			tags := compute.RegionNetworkFirewallPolicyRuleTargetSecureTagArray{}
			for _, name := range in.TargetSecureTags {
				tags = append(tags, &compute.RegionNetworkFirewallPolicyRuleTargetSecureTagArgs{Name: name})
			}
			ruleArgs.TargetSecureTags = tags
		}
		if locals.ProjectId != "" {
			ruleArgs.Project = pulumi.String(locals.ProjectId)
		}
		if spec.DeletionPolicy != "" {
			ruleArgs.DeletionPolicy = pulumi.StringPtr(spec.DeletionPolicy)
		}
		if _, err := compute.NewRegionNetworkFirewallPolicyRule(ctx, fmt.Sprintf("rule-%d", in.Priority), ruleArgs,
			pulumi.Provider(gcpProvider), pulumi.Parent(createdPolicy)); err != nil {
			return errors.Wrapf(err, "failed to create rule at priority %d", in.Priority)
		}
	}

	for i, association := range spec.Associations {
		name := locals.AssociationNames[i]
		associationArgs := &compute.RegionNetworkFirewallPolicyAssociationArgs{
			FirewallPolicy:   createdPolicy.Name,
			Region:           region,
			Name:             pulumi.String(name),
			AttachmentTarget: pulumi.String(association.Network.GetValue()),
		}
		if locals.ProjectId != "" {
			associationArgs.Project = pulumi.String(locals.ProjectId)
		}
		if spec.DeletionPolicy != "" {
			associationArgs.DeletionPolicy = pulumi.StringPtr(spec.DeletionPolicy)
		}
		if _, err := compute.NewRegionNetworkFirewallPolicyAssociation(ctx, "association-"+name, associationArgs,
			pulumi.Provider(gcpProvider), pulumi.Parent(createdPolicy)); err != nil {
			return errors.Wrapf(err, "failed to create association %s", name)
		}
	}

	ctx.Export(OpPolicyName, createdPolicy.Name)
	ctx.Export(OpPolicyId, createdPolicy.RegionNetworkFirewallPolicyId)
	ctx.Export(OpSelfLink, createdPolicy.SelfLink)
	ctx.Export(OpRegion, createdPolicy.Region)
	ctx.Export(OpRuleTupleCount, createdPolicy.RuleTupleCount)
	ctx.Export(OpAssociationNames, pulumi.ToStringArray(locals.AssociationNames))

	return nil
}

func regionalMatchArgs(in matchInputs) *compute.RegionNetworkFirewallPolicyRuleMatchArgs {
	layer4 := compute.RegionNetworkFirewallPolicyRuleMatchLayer4ConfigArray{}
	for _, config := range in.Layer4Configs {
		layer4 = append(layer4, &compute.RegionNetworkFirewallPolicyRuleMatchLayer4ConfigArgs{
			IpProtocol: config.IpProtocol,
			Ports:      config.Ports,
		})
	}
	args := &compute.RegionNetworkFirewallPolicyRuleMatchArgs{
		Layer4Configs:           layer4,
		SrcIpRanges:             in.SrcIpRanges,
		DestIpRanges:            in.DestIpRanges,
		SrcAddressGroups:        in.SrcAddressGroups,
		DestAddressGroups:       in.DestAddressGroups,
		SrcFqdns:                in.SrcFqdns,
		DestFqdns:               in.DestFqdns,
		SrcRegionCodes:          in.SrcRegionCodes,
		DestRegionCodes:         in.DestRegionCodes,
		SrcThreatIntelligences:  in.SrcThreatIntelligences,
		DestThreatIntelligences: in.DestThreatIntelligences,
		SrcNetworks:             in.SrcNetworks,
		SrcNetworkContext:       in.SrcNetworkContext,
		DestNetworkContext:      in.DestNetworkContext,
	}
	if len(in.SrcSecureTags) > 0 {
		tags := compute.RegionNetworkFirewallPolicyRuleMatchSrcSecureTagArray{}
		for _, name := range in.SrcSecureTags {
			tags = append(tags, &compute.RegionNetworkFirewallPolicyRuleMatchSrcSecureTagArgs{Name: name})
		}
		args.SrcSecureTags = tags
	}
	return args
}
