package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/ec2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"

	// Adjust import path to match your actual proto package path
	awssecuritygroupv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/aws/awssecuritygroup/v1"
)

// securityGroup creates an AWS EC2 Security Group within the specified VPC.
// Ingress and egress rules are mapped from the repeated fields in AwsSecurityGroupSpec.
func securityGroup(ctx *pulumi.Context, locals *Locals, provider *aws.Provider) error {
	spec := locals.AwsSecurityGroup.Spec

	// If your proto field names differ (e.g., "ingress" vs. "ingressRules"), update accordingly.
	ingressRules := buildIngress(spec.Ingress)
	egressRules := buildEgress(spec.Egress)

	sg, err := ec2.NewSecurityGroup(ctx, locals.AwsSecurityGroup.Metadata.Name, &ec2.SecurityGroupArgs{
		VpcId:       pulumi.String(spec.VpcId.GetValue()),
		Name:        pulumi.String(locals.AwsSecurityGroup.Metadata.Name),
		Description: pulumi.String(spec.Description),
		Ingress:     ingressRules,
		Egress:      egressRules,
		Tags:        pulumi.ToStringMap(locals.AwsTags),
	}, pulumi.Provider(provider))
	if err != nil {
		return errors.Wrap(err, "unable to create AWS Security Group")
	}

	// Export stack outputs

	// If you want to expose the Security Group ID, do so here:
	ctx.Export(OpSecurityGroupId, sg.ID())

	return nil
}

// buildIngress converts proto-based SecurityGroupRules into Pulumi's SecurityGroupIngressArgs array.
func buildIngress(rules []*awssecuritygroupv1.SecurityGroupRule) ec2.SecurityGroupIngressArray {
	var ingress ec2.SecurityGroupIngressArray
	for _, r := range rules {
		ingress = append(ingress, ruleToIngress(r)...)
	}
	return ingress
}

// buildEgress converts proto-based SecurityGroupRules into Pulumi's SecurityGroupEgressArgs array.
func buildEgress(rules []*awssecuritygroupv1.SecurityGroupRule) ec2.SecurityGroupEgressArray {
	var egress ec2.SecurityGroupEgressArray
	for _, r := range rules {
		egress = append(egress, ruleToEgress(r)...)
	}
	return egress
}

// ruleToIngress maps a single SecurityGroupRule to one or more SecurityGroupIngressArgs entries.
func ruleToIngress(r *awssecuritygroupv1.SecurityGroupRule) ec2.SecurityGroupIngressArray {
	var sourceSGs pulumi.StringArray
	for _, sg := range r.SourceSecurityGroupIds {
		sourceSGs = append(sourceSGs, pulumi.String(sg.GetValue()))
	}
	return ec2.SecurityGroupIngressArray{
		&ec2.SecurityGroupIngressArgs{
			Protocol:       pulumi.String(r.Protocol),
			FromPort:       pulumi.Int(int(r.FromPort)),
			ToPort:         pulumi.Int(int(r.ToPort)),
			CidrBlocks:     pulumi.ToStringArray(r.Ipv4Cidrs),
			Ipv6CidrBlocks: pulumi.ToStringArray(r.Ipv6Cidrs),
			SecurityGroups: sourceSGs,
			Self:           pulumi.Bool(r.SelfReference),
			Description:    pulumi.String(r.Description),
		},
	}
}

// ruleToEgress maps a single SecurityGroupRule to one or more SecurityGroupEgressArgs entries.
func ruleToEgress(r *awssecuritygroupv1.SecurityGroupRule) ec2.SecurityGroupEgressArray {
	var destSGs pulumi.StringArray
	for _, sg := range r.DestinationSecurityGroupIds {
		destSGs = append(destSGs, pulumi.String(sg.GetValue()))
	}
	return ec2.SecurityGroupEgressArray{
		&ec2.SecurityGroupEgressArgs{
			Protocol:       pulumi.String(r.Protocol),
			FromPort:       pulumi.Int(int(r.FromPort)),
			ToPort:         pulumi.Int(int(r.ToPort)),
			CidrBlocks:     pulumi.ToStringArray(r.Ipv4Cidrs),
			Ipv6CidrBlocks: pulumi.ToStringArray(r.Ipv6Cidrs),
			SecurityGroups: destSGs,
			Self:           pulumi.Bool(r.SelfReference),
			Description:    pulumi.String(r.Description),
		},
	}
}
