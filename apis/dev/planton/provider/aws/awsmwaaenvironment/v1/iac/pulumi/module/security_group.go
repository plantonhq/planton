package module

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/ec2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func securityGroup(ctx *pulumi.Context, locals *Locals, provider *aws.Provider) (*ec2.SecurityGroup, error) {
	spec := locals.AwsMwaaEnvironment.Spec
	if spec == nil {
		return nil, nil
	}

	hasIngressRefs := len(spec.SecurityGroupIds) > 0 || len(spec.AllowedCidrBlocks) > 0
	if !hasIngressRefs {
		return nil, nil
	}

	vpcId := ""
	if spec.VpcId != nil {
		vpcId = spec.VpcId.GetValue()
	}

	sg, err := ec2.NewSecurityGroup(ctx, "environment-sg", &ec2.SecurityGroupArgs{
		Name:        pulumi.String(locals.AwsMwaaEnvironment.Metadata.Id),
		Description: pulumi.String("Managed security group for MWAA environment"),
		VpcId:       pulumi.String(vpcId),
		Tags:        pulumi.ToStringMap(locals.Labels),
	}, pulumi.Provider(provider))
	if err != nil {
		return nil, errors.Wrap(err, "create security group")
	}

	// Self-referencing inbound rule: MWAA VPC endpoints must communicate with each other.
	// This is the AWS-recommended configuration for MWAA security groups.
	_, err = ec2.NewSecurityGroupRule(ctx, "ingress-self", &ec2.SecurityGroupRuleArgs{
		Type:                  pulumi.String("ingress"),
		FromPort:              pulumi.Int(0),
		ToPort:                pulumi.Int(0),
		Protocol:              pulumi.String("-1"),
		SourceSecurityGroupId: sg.ID(),
		SecurityGroupId:       sg.ID(),
	}, pulumi.Provider(provider), pulumi.Parent(sg))
	if err != nil {
		return nil, errors.Wrap(err, "create self-referencing ingress rule")
	}

	// Ingress on port 443 (HTTPS) from source security groups for Airflow UI access
	for i, sgOrRef := range spec.SecurityGroupIds {
		_, err := ec2.NewSecurityGroupRule(ctx, fmt.Sprintf("ingress-https-sg-%d", i), &ec2.SecurityGroupRuleArgs{
			Type:                  pulumi.String("ingress"),
			FromPort:              pulumi.Int(443),
			ToPort:                pulumi.Int(443),
			Protocol:              pulumi.String("tcp"),
			SourceSecurityGroupId: pulumi.String(sgOrRef.GetValue()),
			SecurityGroupId:       sg.ID(),
		}, pulumi.Provider(provider), pulumi.Parent(sg))
		if err != nil {
			return nil, errors.Wrapf(err, "create HTTPS ingress rule from sg %d", i)
		}
	}

	// Ingress on port 443 from CIDR blocks
	if len(spec.AllowedCidrBlocks) > 0 {
		_, err := ec2.NewSecurityGroupRule(ctx, "ingress-https-cidr", &ec2.SecurityGroupRuleArgs{
			Type:            pulumi.String("ingress"),
			FromPort:        pulumi.Int(443),
			ToPort:          pulumi.Int(443),
			Protocol:        pulumi.String("tcp"),
			CidrBlocks:      pulumi.ToStringArray(spec.AllowedCidrBlocks),
			SecurityGroupId: sg.ID(),
		}, pulumi.Provider(provider), pulumi.Parent(sg))
		if err != nil {
			return nil, errors.Wrap(err, "create HTTPS ingress rule from cidr")
		}
	}

	// Allow all outbound traffic
	_, err = ec2.NewSecurityGroupRule(ctx, "egress-all", &ec2.SecurityGroupRuleArgs{
		Type:            pulumi.String("egress"),
		FromPort:        pulumi.Int(0),
		ToPort:          pulumi.Int(0),
		Protocol:        pulumi.String("-1"),
		CidrBlocks:      pulumi.StringArray{pulumi.String("0.0.0.0/0")},
		SecurityGroupId: sg.ID(),
	}, pulumi.Provider(provider), pulumi.Parent(sg))
	if err != nil {
		return nil, errors.Wrap(err, "create egress rule")
	}

	return sg, nil
}
