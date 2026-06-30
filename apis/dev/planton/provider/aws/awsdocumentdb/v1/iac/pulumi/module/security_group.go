package module

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/ec2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func securityGroup(ctx *pulumi.Context, locals *Locals, provider *aws.Provider) (*ec2.SecurityGroup, error) {
	spec := locals.AwsDocumentDb.Spec

	if spec == nil {
		return nil, nil
	}

	// If neither CIDRs nor SG attachments are provided, skip creating SG
	hasIngressRefs := len(spec.SecurityGroups) > 0 || len(spec.AllowedCidrs) > 0
	if !hasIngressRefs {
		return nil, nil
	}

	vpcId := ""
	if spec.Vpc != nil {
		vpcId = spec.Vpc.GetValue()
	}

	sg, err := ec2.NewSecurityGroup(ctx, "cluster-sg", &ec2.SecurityGroupArgs{
		Name:        pulumi.String(locals.AwsDocumentDb.Metadata.Id),
		Description: pulumi.String("Security group for DocumentDB cluster"),
		VpcId:       pulumi.String(vpcId),
		Tags:        pulumi.ToStringMap(locals.Labels),
	}, pulumi.Provider(provider))
	if err != nil {
		return nil, errors.Wrap(err, "create security group")
	}

	port := getEffectivePort(spec)

	// Ingress from security groups
	for i, sgOrRef := range spec.SecurityGroups {
		if sgOrRef.GetValue() == "" {
			continue
		}
		_, err := ec2.NewSecurityGroupRule(ctx, fmt.Sprintf("ingress-from-sg-%d", i), &ec2.SecurityGroupRuleArgs{
			Type:                  pulumi.String("ingress"),
			FromPort:              pulumi.Int(port),
			ToPort:                pulumi.Int(port),
			Protocol:              pulumi.String("tcp"),
			SourceSecurityGroupId: pulumi.String(sgOrRef.GetValue()),
			SecurityGroupId:       sg.ID(),
		}, pulumi.Provider(provider), pulumi.Parent(sg))
		if err != nil {
			return nil, errors.Wrap(err, "create sg ingress rule from sg")
		}
	}

	// Ingress from CIDRs
	if len(spec.AllowedCidrs) > 0 {
		_, err := ec2.NewSecurityGroupRule(ctx, "ingress-from-cidr", &ec2.SecurityGroupRuleArgs{
			Type:            pulumi.String("ingress"),
			FromPort:        pulumi.Int(port),
			ToPort:          pulumi.Int(port),
			Protocol:        pulumi.String("tcp"),
			CidrBlocks:      pulumi.ToStringArray(spec.AllowedCidrs),
			SecurityGroupId: sg.ID(),
		}, pulumi.Provider(provider), pulumi.Parent(sg))
		if err != nil {
			return nil, errors.Wrap(err, "create sg ingress rule from cidr")
		}
	}

	// Egress all
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
