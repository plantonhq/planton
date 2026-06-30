package module

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-openstack/sdk/v5/go/openstack"
	"github.com/pulumi/pulumi-openstack/sdk/v5/go/openstack/networking"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// securityGroup provisions the OpenStack Neutron security group, creates inline
// rules, and exports outputs.
func securityGroup(
	ctx *pulumi.Context,
	locals *Locals,
	openstackProvider *openstack.Provider,
) error {
	spec := locals.OpenStackSecurityGroup.Spec
	sgName := locals.OpenStackSecurityGroup.Metadata.Name

	sgArgs := &networking.SecGroupArgs{
		Name: pulumi.String(sgName),
	}

	// Set description if provided.
	if spec.Description != "" {
		sgArgs.Description = pulumi.StringPtr(spec.Description)
	}

	// Set delete_default_rules if explicitly provided.
	if spec.DeleteDefaultRules != nil {
		sgArgs.DeleteDefaultRules = pulumi.BoolPtr(spec.GetDeleteDefaultRules())
	}

	// Set stateful if explicitly provided.
	if spec.Stateful != nil {
		sgArgs.Stateful = pulumi.BoolPtr(spec.GetStateful())
	}

	// Set tags if provided.
	if len(spec.Tags) > 0 {
		tags := make(pulumi.StringArray, len(spec.Tags))
		for i, tag := range spec.Tags {
			tags[i] = pulumi.String(tag)
		}
		sgArgs.Tags = tags
	}

	// Set region override if provided.
	if spec.Region != "" {
		sgArgs.Region = pulumi.StringPtr(spec.Region)
	}

	createdSG, err := networking.NewSecGroup(
		ctx,
		strings.ToLower(sgName),
		sgArgs,
		pulumi.Provider(openstackProvider),
	)
	if err != nil {
		return errors.Wrap(err, "failed to create openstack security group")
	}

	// Create inline rules. Each rule is a separate SecGroupRule resource,
	// keyed by the rule's `key` field for stable naming.
	for _, rule := range spec.Rules {
		ruleArgs := &networking.SecGroupRuleArgs{
			SecurityGroupId: createdSG.ID(),
			Direction:       pulumi.String(rule.Direction),
			Ethertype:       pulumi.String(rule.Ethertype),
		}

		if rule.Protocol != "" {
			ruleArgs.Protocol = pulumi.StringPtr(rule.Protocol)
		}

		if rule.PortRangeMin != nil {
			ruleArgs.PortRangeMin = pulumi.IntPtr(int(rule.GetPortRangeMin()))
		}

		if rule.PortRangeMax != nil {
			ruleArgs.PortRangeMax = pulumi.IntPtr(int(rule.GetPortRangeMax()))
		}

		if rule.RemoteIpPrefix != "" {
			ruleArgs.RemoteIpPrefix = pulumi.StringPtr(rule.RemoteIpPrefix)
		}

		if rule.RemoteGroupId != "" {
			ruleArgs.RemoteGroupId = pulumi.StringPtr(rule.RemoteGroupId)
		}

		if rule.Description != "" {
			ruleArgs.Description = pulumi.StringPtr(rule.Description)
		}

		// Set region on the rule to match the security group.
		if spec.Region != "" {
			ruleArgs.Region = pulumi.StringPtr(spec.Region)
		}

		ruleName := fmt.Sprintf("%s-rule-%s", strings.ToLower(sgName), rule.Key)
		_, err := networking.NewSecGroupRule(
			ctx,
			ruleName,
			ruleArgs,
			pulumi.Provider(openstackProvider),
			pulumi.DependsOn([]pulumi.Resource{createdSG}),
		)
		if err != nil {
			return errors.Wrapf(err, "failed to create security group rule %q", rule.Key)
		}
	}

	// Export required outputs (matching stack_outputs.proto fields).
	ctx.Export(OpSecurityGroupId, createdSG.ID())
	ctx.Export(OpName, createdSG.Name)
	ctx.Export(OpRegion, createdSG.Region)

	return nil
}
