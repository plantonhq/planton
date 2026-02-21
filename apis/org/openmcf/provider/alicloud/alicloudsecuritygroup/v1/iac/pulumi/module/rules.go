package module

import (
	"fmt"

	"github.com/pkg/errors"
	alicloudsecuritygroupv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/alicloud/alicloudsecuritygroup/v1"
	"github.com/pulumi/pulumi-alicloud/sdk/v3/go/alicloud"
	"github.com/pulumi/pulumi-alicloud/sdk/v3/go/alicloud/ecs"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func securityGroupRule(
	ctx *pulumi.Context,
	provider *alicloud.Provider,
	sg *ecs.SecurityGroup,
	sgName string,
	index int,
	rule *alicloudsecuritygroupv1.AliCloudSecurityGroupRule,
) error {
	resourceName := fmt.Sprintf("%s-rule-%d", sgName, index)

	args := &ecs.SecurityGroupRuleArgs{
		SecurityGroupId: sg.ID(),
		Type:            pulumi.String(rule.Type),
		IpProtocol:      pulumi.String(rule.IpProtocol),
		PortRange:       pulumi.String(rulePortRange(rule)),
		NicType:         pulumi.String("intranet"),
		Priority:        pulumi.Int(rulePriority(rule)),
		Policy:          pulumi.String(rulePolicy(rule)),
	}

	if rule.CidrIp != "" {
		args.CidrIp = pulumi.String(rule.CidrIp)
	}

	if rule.SourceSecurityGroupId != "" {
		args.SourceSecurityGroupId = pulumi.String(rule.SourceSecurityGroupId)
	}

	if rule.Description != "" {
		args.Description = pulumi.String(rule.Description)
	}

	_, err := ecs.NewSecurityGroupRule(ctx, resourceName, args,
		pulumi.Provider(provider),
		pulumi.Parent(sg),
	)
	if err != nil {
		return errors.Wrapf(err, "failed to create security group rule %s", resourceName)
	}

	return nil
}
