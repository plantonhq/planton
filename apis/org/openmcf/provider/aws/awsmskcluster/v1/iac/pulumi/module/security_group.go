package module

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/ec2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func securityGroup(ctx *pulumi.Context, locals *Locals, provider *aws.Provider) (*ec2.SecurityGroup, error) {
	spec := locals.AwsMskCluster.Spec
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

	sg, err := ec2.NewSecurityGroup(ctx, "cluster-sg", &ec2.SecurityGroupArgs{
		Name:        pulumi.String(locals.AwsMskCluster.Metadata.Id),
		Description: pulumi.String("Ingress for MSK cluster"),
		VpcId:       pulumi.String(vpcId),
		Tags:        pulumi.ToStringMap(locals.Labels),
	}, pulumi.Provider(provider))
	if err != nil {
		return nil, errors.Wrap(err, "create security group")
	}

	// Kafka broker ports: 9092 (plaintext), 9094 (TLS), 9096 (SASL/SCRAM), 9098 (SASL/IAM)
	kafkaFromPort := 9092
	kafkaToPort := 9098

	// ZooKeeper ports: 2181 (plaintext), 2182 (TLS)
	zkFromPort := 2181
	zkToPort := 2182

	// Ingress rules from source security groups
	for i, sgOrRef := range spec.SecurityGroupIds {
		_, err := ec2.NewSecurityGroupRule(ctx, fmt.Sprintf("ingress-kafka-sg-%d", i), &ec2.SecurityGroupRuleArgs{
			Type:                  pulumi.String("ingress"),
			FromPort:              pulumi.Int(kafkaFromPort),
			ToPort:                pulumi.Int(kafkaToPort),
			Protocol:              pulumi.String("tcp"),
			SourceSecurityGroupId: pulumi.String(sgOrRef.GetValue()),
			SecurityGroupId:       sg.ID(),
		}, pulumi.Provider(provider), pulumi.Parent(sg))
		if err != nil {
			return nil, errors.Wrapf(err, "create kafka ingress rule from sg %d", i)
		}

		_, err = ec2.NewSecurityGroupRule(ctx, fmt.Sprintf("ingress-zk-sg-%d", i), &ec2.SecurityGroupRuleArgs{
			Type:                  pulumi.String("ingress"),
			FromPort:              pulumi.Int(zkFromPort),
			ToPort:                pulumi.Int(zkToPort),
			Protocol:              pulumi.String("tcp"),
			SourceSecurityGroupId: pulumi.String(sgOrRef.GetValue()),
			SecurityGroupId:       sg.ID(),
		}, pulumi.Provider(provider), pulumi.Parent(sg))
		if err != nil {
			return nil, errors.Wrapf(err, "create zk ingress rule from sg %d", i)
		}
	}

	// Ingress rules from CIDR blocks
	if len(spec.AllowedCidrBlocks) > 0 {
		_, err := ec2.NewSecurityGroupRule(ctx, "ingress-kafka-cidr", &ec2.SecurityGroupRuleArgs{
			Type:            pulumi.String("ingress"),
			FromPort:        pulumi.Int(kafkaFromPort),
			ToPort:          pulumi.Int(kafkaToPort),
			Protocol:        pulumi.String("tcp"),
			CidrBlocks:      pulumi.ToStringArray(spec.AllowedCidrBlocks),
			SecurityGroupId: sg.ID(),
		}, pulumi.Provider(provider), pulumi.Parent(sg))
		if err != nil {
			return nil, errors.Wrap(err, "create kafka ingress rule from cidr")
		}

		_, err = ec2.NewSecurityGroupRule(ctx, "ingress-zk-cidr", &ec2.SecurityGroupRuleArgs{
			Type:            pulumi.String("ingress"),
			FromPort:        pulumi.Int(zkFromPort),
			ToPort:          pulumi.Int(zkToPort),
			Protocol:        pulumi.String("tcp"),
			CidrBlocks:      pulumi.ToStringArray(spec.AllowedCidrBlocks),
			SecurityGroupId: sg.ID(),
		}, pulumi.Provider(provider), pulumi.Parent(sg))
		if err != nil {
			return nil, errors.Wrap(err, "create zk ingress rule from cidr")
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
