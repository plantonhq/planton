package module

import (
	gcpnetworkfirewallpolicyv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpnetworkfirewallpolicy/v1alpha1"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// ruleInputs is a rule rendered once into provider-neutral Pulumi inputs.
// The global and regional resource families type every block separately
// (NetworkFirewallPolicyRuleMatchArgs versus
// RegionNetworkFirewallPolicyRuleMatchArgs, and so on), so each arm maps
// these inputs onto its own structs; the rendering rules -- lists only when
// non-empty, Optional+Computed enums only when set, the security profile
// group and TLS flag together -- live here exactly once.
type ruleInputs struct {
	Priority              int
	Action                pulumi.StringInput
	Direction             pulumi.StringInput
	Description           pulumi.StringPtrInput
	RuleName              pulumi.StringPtrInput
	Disabled              pulumi.BoolPtrInput
	EnableLogging         pulumi.BoolPtrInput
	TargetType            pulumi.StringPtrInput
	TargetForwardingRules pulumi.StringArrayInput
	TargetSecureTags      pulumi.StringArray
	TargetServiceAccounts pulumi.StringArrayInput
	SecurityProfileGroup  pulumi.StringPtrInput
	TlsInspect            pulumi.BoolPtrInput
	Match                 matchInputs
}

// matchInputs is the match block rendered once; a nil pointer or empty
// array means "not sent".
type matchInputs struct {
	Layer4Configs           []layer4Inputs
	SrcIpRanges             pulumi.StringArrayInput
	DestIpRanges            pulumi.StringArrayInput
	SrcAddressGroups        pulumi.StringArrayInput
	DestAddressGroups       pulumi.StringArrayInput
	SrcFqdns                pulumi.StringArrayInput
	DestFqdns               pulumi.StringArrayInput
	SrcRegionCodes          pulumi.StringArrayInput
	DestRegionCodes         pulumi.StringArrayInput
	SrcThreatIntelligences  pulumi.StringArrayInput
	DestThreatIntelligences pulumi.StringArrayInput
	SrcSecureTags           pulumi.StringArray
	SrcNetworks             pulumi.StringArrayInput
	SrcNetworkContext       pulumi.StringPtrInput
	DestNetworkContext      pulumi.StringPtrInput
}

type layer4Inputs struct {
	IpProtocol pulumi.StringInput
	Ports      pulumi.StringArrayInput
}

func renderRule(rule *gcpnetworkfirewallpolicyv1alpha1.GcpNetworkFirewallPolicyRule) ruleInputs {
	inputs := ruleInputs{
		Priority:              int(rule.GetPriority()),
		Action:                pulumi.String(rule.Action),
		Direction:             pulumi.String(rule.Direction),
		Disabled:              pulumi.BoolPtr(rule.Disabled),
		EnableLogging:         pulumi.BoolPtr(rule.EnableLogging),
		TargetForwardingRules: arrayInput(refListValues(rule.TargetForwardingRules)),
		TargetSecureTags:      refListValues(rule.TargetSecureTags),
		TargetServiceAccounts: arrayInput(refListValues(rule.TargetServiceAccounts)),
		Match:                 renderMatch(rule.Match),
	}
	if rule.Description != "" {
		inputs.Description = pulumi.StringPtr(rule.Description)
	}
	if rule.RuleName != "" {
		inputs.RuleName = pulumi.StringPtr(rule.RuleName)
	}
	// Optional+Computed: unset stays out of the payload so Google's default
	// (INSTANCES) never shows as a diff on re-plan.
	if rule.TargetType != nil && rule.GetTargetType() != "" {
		inputs.TargetType = pulumi.StringPtr(rule.GetTargetType())
	}
	// Present exactly when action is apply_security_profile_group (spec
	// CEL); tls_inspect rides along with that action only.
	if rule.SecurityProfileGroup != "" {
		inputs.SecurityProfileGroup = pulumi.StringPtr(rule.SecurityProfileGroup)
		inputs.TlsInspect = pulumi.BoolPtr(rule.TlsInspect)
	}
	return inputs
}

func renderMatch(match *gcpnetworkfirewallpolicyv1alpha1.GcpNetworkFirewallPolicyRuleMatch) matchInputs {
	inputs := matchInputs{
		SrcIpRanges:             stringArrayOrNil(match.SrcIpRanges),
		DestIpRanges:            stringArrayOrNil(match.DestIpRanges),
		SrcAddressGroups:        stringArrayOrNil(match.SrcAddressGroups),
		DestAddressGroups:       stringArrayOrNil(match.DestAddressGroups),
		SrcFqdns:                stringArrayOrNil(match.SrcFqdns),
		DestFqdns:               stringArrayOrNil(match.DestFqdns),
		SrcRegionCodes:          stringArrayOrNil(match.SrcRegionCodes),
		DestRegionCodes:         stringArrayOrNil(match.DestRegionCodes),
		SrcThreatIntelligences:  stringArrayOrNil(match.SrcThreatIntelligences),
		DestThreatIntelligences: stringArrayOrNil(match.DestThreatIntelligences),
		SrcSecureTags:           refListValues(match.SrcSecureTags),
		SrcNetworks:             arrayInput(refListValues(match.SrcNetworks)),
	}
	for _, config := range match.Layer4Configs {
		inputs.Layer4Configs = append(inputs.Layer4Configs, layer4Inputs{
			IpProtocol: pulumi.String(config.IpProtocol),
			Ports:      stringArrayOrNil(config.Ports),
		})
	}
	if match.SrcNetworkContext != nil && match.GetSrcNetworkContext() != "" {
		inputs.SrcNetworkContext = pulumi.StringPtr(match.GetSrcNetworkContext())
	}
	if match.DestNetworkContext != nil && match.GetDestNetworkContext() != "" {
		inputs.DestNetworkContext = pulumi.StringPtr(match.GetDestNetworkContext())
	}
	return inputs
}

// stringArrayOrNil sends a list only when it has entries: the returned
// interface is a true nil for an empty list, so the SDK omits the argument
// instead of sending an empty array that would diff against the API's
// unset value on re-plan.
func stringArrayOrNil(values []string) pulumi.StringArrayInput {
	if len(values) == 0 {
		return nil
	}
	return pulumi.ToStringArray(values)
}

// arrayInput is stringArrayOrNil for an already-rendered array.
func arrayInput(values pulumi.StringArray) pulumi.StringArrayInput {
	if len(values) == 0 {
		return nil
	}
	return values
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
