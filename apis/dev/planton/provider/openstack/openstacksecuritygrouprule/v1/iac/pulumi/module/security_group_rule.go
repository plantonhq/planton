package module

import (
	"strings"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-openstack/sdk/v5/go/openstack"
	"github.com/pulumi/pulumi-openstack/sdk/v5/go/openstack/networking"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// securityGroupRule provisions a standalone OpenStack Neutron security group rule
// and exports outputs.
func securityGroupRule(
	ctx *pulumi.Context,
	locals *Locals,
	openstackProvider *openstack.Provider,
) error {
	spec := locals.OpenStackSecurityGroupRule.Spec
	resourceName := locals.OpenStackSecurityGroupRule.Metadata.Name

	ruleArgs := &networking.SecGroupRuleArgs{
		SecurityGroupId: pulumi.String(locals.SecurityGroupId),
		Direction:       pulumi.String(spec.Direction),
		Ethertype:       pulumi.String(spec.Ethertype),
	}

	// Set protocol if provided (empty = all protocols).
	if spec.Protocol != "" {
		ruleArgs.Protocol = pulumi.StringPtr(spec.Protocol)
	}

	// Set port range if provided.
	if spec.PortRangeMin != nil {
		ruleArgs.PortRangeMin = pulumi.IntPtr(int(spec.GetPortRangeMin()))
	}
	if spec.PortRangeMax != nil {
		ruleArgs.PortRangeMax = pulumi.IntPtr(int(spec.GetPortRangeMax()))
	}

	// Set remote source (mutually exclusive, enforced by proto CEL validation).
	if spec.RemoteIpPrefix != "" {
		ruleArgs.RemoteIpPrefix = pulumi.StringPtr(spec.RemoteIpPrefix)
	}
	if locals.RemoteGroupId != "" {
		ruleArgs.RemoteGroupId = pulumi.StringPtr(locals.RemoteGroupId)
	}

	// Set description if provided.
	if spec.Description != "" {
		ruleArgs.Description = pulumi.StringPtr(spec.Description)
	}

	// Set region override if provided.
	if spec.Region != "" {
		ruleArgs.Region = pulumi.StringPtr(spec.Region)
	}

	createdRule, err := networking.NewSecGroupRule(
		ctx,
		strings.ToLower(resourceName),
		ruleArgs,
		pulumi.Provider(openstackProvider),
	)
	if err != nil {
		return errors.Wrap(err, "failed to create openstack security group rule")
	}

	// Export required outputs (matching stack_outputs.proto fields).
	ctx.Export(OpRuleId, createdRule.ID())
	ctx.Export(OpSecurityGroupId, createdRule.SecurityGroupId)
	ctx.Export(OpDirection, createdRule.Direction)
	ctx.Export(OpProtocol, createdRule.Protocol)
	ctx.Export(OpPortRangeMin, createdRule.PortRangeMin)
	ctx.Export(OpPortRangeMax, createdRule.PortRangeMax)
	ctx.Export(OpRegion, createdRule.Region)

	return nil
}
