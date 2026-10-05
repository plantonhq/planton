package module

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/plantonhq/planton/internal/valuefrom"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/ec2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// securityGroup creates the runner's own outbound-only group. The runner
// initiates every connection it uses (control plane, work queue, image
// pulls), so the group carries the permissive egress rule and NO inbound
// rules at all -- running the appliance adds zero inbound surface to the
// VPC. Private targets that admit traffic by source security group
// reference this group's id (published as an output) to trust the
// runner.
//
// The VPC is derived from the first referenced subnet rather than asked
// for on the spec: a separate vpc field could only ever agree with the
// subnets or contradict them.
func securityGroup(ctx *pulumi.Context, locals *Locals, provider *aws.Provider) (*ec2.SecurityGroup, error) {
	runnerName := locals.AwsPlantonRunner.Metadata.Name
	subnetIds := valuefrom.ToStringArray(locals.AwsPlantonRunner.Spec.Subnets)
	if len(subnetIds) == 0 {
		return nil, errors.New("at least one subnet is required")
	}

	firstSubnet, err := ec2.LookupSubnet(ctx,
		&ec2.LookupSubnetArgs{Id: &subnetIds[0]},
		pulumi.Provider(provider))
	if err != nil {
		return nil, errors.Wrapf(err, "failed to look up subnet %s", subnetIds[0])
	}

	createdSecurityGroup, err := ec2.NewSecurityGroup(ctx,
		"security-group",
		&ec2.SecurityGroupArgs{
			Name: pulumi.String(runnerName),
			// SG descriptions reject quote characters (the API's allowed
			// set is a-zA-Z0-9. _-:/()#,@[]+=&;{}!$*), so the name is
			// embedded bare.
			Description: pulumi.String(fmt.Sprintf("Planton runner %s -- outbound only, no inbound", runnerName)),
			VpcId:       pulumi.String(firstSubnet.VpcId),
			Egress: ec2.SecurityGroupEgressArray{
				&ec2.SecurityGroupEgressArgs{
					Protocol:   pulumi.String("-1"),
					FromPort:   pulumi.Int(0),
					ToPort:     pulumi.Int(0),
					CidrBlocks: pulumi.StringArray{pulumi.String("0.0.0.0/0")},
				},
			},
			Tags: pulumi.ToStringMap(locals.AwsTags),
		},
		pulumi.Provider(provider))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create security group")
	}

	return createdSecurityGroup, nil
}
