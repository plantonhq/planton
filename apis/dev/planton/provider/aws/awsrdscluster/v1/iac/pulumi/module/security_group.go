package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/ec2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func securityGroup(ctx *pulumi.Context, locals *Locals, provider *aws.Provider) (*ec2.SecurityGroup, error) {
	spec := locals.AwsRdsCluster.Spec

	if spec == nil {
		return nil, nil
	}

	// If neither CIDRs nor SG attachments are provided, skip creating SG
	hasIngressRefs := len(spec.SecurityGroupIds) > 0 || len(spec.AllowedCidrBlocks) > 0
	if !hasIngressRefs {
		return nil, nil
	}

	vpcId := ""
	if spec.VpcId != nil {
		vpcId = spec.VpcId.GetValue()
	}

	sg, err := ec2.NewSecurityGroup(ctx, "cluster-sg", &ec2.SecurityGroupArgs{
		Name:        pulumi.String(locals.AwsRdsCluster.Metadata.Id),
		Description: pulumi.String("Ingress for RDS cluster"),
		VpcId:       pulumi.String(vpcId),
		Tags:        pulumi.ToStringMap(locals.Labels),
	}, pulumi.Provider(provider))
	if err != nil {
		return nil, errors.Wrap(err, "create security group")
	}

	// from SGs
	for _, sgOrRef := range spec.SecurityGroupIds {
		_, err := ec2.NewSecurityGroupRule(ctx, "ingress-from-sg", &ec2.SecurityGroupRuleArgs{
			Type:                  pulumi.String("ingress"),
			FromPort:              pulumi.Int(spec.Port),
			ToPort:                pulumi.Int(spec.Port),
			Protocol:              pulumi.String("tcp"),
			SourceSecurityGroupId: pulumi.String(sgOrRef.GetValue()),
			SecurityGroupId:       sg.ID(),
		}, pulumi.Provider(provider), pulumi.Parent(sg))
		if err != nil {
			return nil, errors.Wrap(err, "create sg ingress rule from sg")
		}
	}

	// from CIDRs
	if len(spec.AllowedCidrBlocks) > 0 {
		_, err := ec2.NewSecurityGroupRule(ctx, "ingress-from-cidr", &ec2.SecurityGroupRuleArgs{
			Type:            pulumi.String("ingress"),
			FromPort:        pulumi.Int(spec.Port),
			ToPort:          pulumi.Int(spec.Port),
			Protocol:        pulumi.String("tcp"),
			CidrBlocks:      pulumi.ToStringArray(spec.AllowedCidrBlocks),
			SecurityGroupId: sg.ID(),
		}, pulumi.Provider(provider), pulumi.Parent(sg))
		if err != nil {
			return nil, errors.Wrap(err, "create sg ingress rule from cidr")
		}
	}

	// all egress
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
