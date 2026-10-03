package module

import (
	"github.com/pkg/errors"
	azurenetworksecuritygroupv1alpha1 "github.com/plantonhq/planton/catalog/azure/azurenetworksecuritygroup/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/pulumiazureprovider"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/network"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *azurenetworksecuritygroupv1alpha1.AzureNetworkSecurityGroupIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the Azure provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static client secret, keyless web identity, or ambient chain).
	azureProvider, err := pulumiazureprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzureNetworkSecurityGroup.Spec

	// Rules are managed inline on the group. This module deliberately uses
	// the inline form (not the standalone rule resource) because the pinned
	// SDK's standalone rule flattens the application-security-group ID
	// lists to a single value -- the inline form carries the full lists.
	// The deployed ARM state is identical either way, and inline gives the
	// same empty-list semantics as the Terraform module: removing the last
	// rule really removes it.
	inlineRules := network.NetworkSecurityGroupSecurityRuleArray{}
	for _, rule := range spec.SecurityRules {
		ruleArgs := network.NetworkSecurityGroupSecurityRuleArgs{
			Name:      pulumi.String(rule.Name),
			Priority:  pulumi.Int(int(rule.Priority)),
			Direction: pulumi.String(directionToArm(rule.Direction)),
			Access:    pulumi.String(accessToArm(rule.Access)),
			Protocol:  pulumi.String(protocolToArm(rule.Protocol)),
		}

		if rule.Description != "" {
			ruleArgs.Description = pulumi.String(rule.Description)
		}

		// Ports: the spec guarantees at most one form is set (exactly one
		// for destination). An unset source means any -- "*" is sent so
		// both engines deploy the identical rule.
		if len(rule.SourcePortRanges) > 0 {
			ruleArgs.SourcePortRanges = pulumi.ToStringArray(rule.SourcePortRanges)
		} else if rule.SourcePortRange != nil {
			ruleArgs.SourcePortRange = pulumi.String(rule.GetSourcePortRange())
		} else {
			ruleArgs.SourcePortRange = pulumi.String("*")
		}
		if len(rule.DestinationPortRanges) > 0 {
			ruleArgs.DestinationPortRanges = pulumi.ToStringArray(rule.DestinationPortRanges)
		} else {
			ruleArgs.DestinationPortRange = pulumi.String(rule.GetDestinationPortRange())
		}

		// Addressing: the spec guarantees at most one style per side
		// (single prefix, prefix list, or application security groups).
		// All unset means any -- "*" is sent so both engines deploy the
		// identical rule.
		// Application security group references are resolved to literal ARM
		// IDs by the platform middleware before the module runs, so
		// GetValue() returns the resolved id for both a literal and a
		// valueFrom reference.
		switch {
		case len(rule.SourceApplicationSecurityGroupIds) > 0:
			sourceAsgIds := make([]string, 0, len(rule.SourceApplicationSecurityGroupIds))
			for _, asg := range rule.SourceApplicationSecurityGroupIds {
				sourceAsgIds = append(sourceAsgIds, asg.GetValue())
			}
			ruleArgs.SourceApplicationSecurityGroupIds = pulumi.ToStringArray(sourceAsgIds)
		case len(rule.SourceAddressPrefixes) > 0:
			ruleArgs.SourceAddressPrefixes = pulumi.ToStringArray(rule.SourceAddressPrefixes)
		case rule.SourceAddressPrefix != nil:
			ruleArgs.SourceAddressPrefix = pulumi.String(rule.GetSourceAddressPrefix())
		default:
			ruleArgs.SourceAddressPrefix = pulumi.String("*")
		}
		switch {
		case len(rule.DestinationApplicationSecurityGroupIds) > 0:
			destAsgIds := make([]string, 0, len(rule.DestinationApplicationSecurityGroupIds))
			for _, asg := range rule.DestinationApplicationSecurityGroupIds {
				destAsgIds = append(destAsgIds, asg.GetValue())
			}
			ruleArgs.DestinationApplicationSecurityGroupIds = pulumi.ToStringArray(destAsgIds)
		case len(rule.DestinationAddressPrefixes) > 0:
			ruleArgs.DestinationAddressPrefixes = pulumi.ToStringArray(rule.DestinationAddressPrefixes)
		case rule.DestinationAddressPrefix != nil:
			ruleArgs.DestinationAddressPrefix = pulumi.String(rule.GetDestinationAddressPrefix())
		default:
			ruleArgs.DestinationAddressPrefix = pulumi.String("*")
		}

		inlineRules = append(inlineRules, ruleArgs)
	}

	// Lifecycle notes worth knowing before operating this resource:
	// - Rules and tags update IN PLACE and take effect immediately for
	//   every subnet and NIC the group guards. Name, region, and resource
	//   group are the group's ARM identity; changing any of them replaces
	//   it, detaching it from every subnet until re-attached.
	// - The group itself is a shell: with no rules, Azure's implicit
	//   defaults govern (allow VNet-internal and load-balancer traffic,
	//   deny other inbound, allow all outbound).
	createdNsg, err := network.NewNetworkSecurityGroup(ctx,
		spec.Name,
		&network.NetworkSecurityGroupArgs{
			Name:              pulumi.String(spec.Name),
			Location:          pulumi.String(spec.Region),
			ResourceGroupName: pulumi.String(locals.ResourceGroupName),
			SecurityRules:     inlineRules,
			Tags:              pulumi.ToStringMap(locals.AzureTags),
		},
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create network security group %s", spec.Name)
	}

	// Export outputs from the created resource.
	ctx.Export(OpNetworkSecurityGroupId, createdNsg.ID())
	ctx.Export(OpNetworkSecurityGroupName, createdNsg.Name)

	return nil
}
