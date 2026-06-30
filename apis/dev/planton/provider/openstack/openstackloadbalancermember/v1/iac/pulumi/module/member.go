package module

import (
	"strings"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-openstack/sdk/v5/go/openstack"
	"github.com/pulumi/pulumi-openstack/sdk/v5/go/openstack/loadbalancer"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// member provisions the OpenStack Octavia pool member and exports outputs.
func member(
	ctx *pulumi.Context,
	locals *Locals,
	openstackProvider *openstack.Provider,
) error {
	spec := locals.OpenStackLoadBalancerMember.Spec
	memberName := locals.OpenStackLoadBalancerMember.Metadata.Name

	memberArgs := &loadbalancer.MemberArgs{
		Name:         pulumi.String(memberName),
		PoolId:       pulumi.String(locals.PoolId),
		Address:      pulumi.String(spec.Address),
		ProtocolPort: pulumi.Int(int(spec.ProtocolPort)),
	}

	// Set subnet_id if provided.
	if locals.SubnetId != "" {
		memberArgs.SubnetId = pulumi.StringPtr(locals.SubnetId)
	}

	// Set weight if explicitly provided.
	if spec.Weight != nil {
		memberArgs.Weight = pulumi.IntPtr(int(spec.GetWeight()))
	}

	// Set admin_state_up if explicitly provided.
	if spec.AdminStateUp != nil {
		memberArgs.AdminStateUp = pulumi.BoolPtr(spec.GetAdminStateUp())
	}

	// Set tags if provided.
	if len(spec.Tags) > 0 {
		tags := make(pulumi.StringArray, len(spec.Tags))
		for i, tag := range spec.Tags {
			tags[i] = pulumi.String(tag)
		}
		memberArgs.Tags = tags
	}

	// Set region override if provided.
	if spec.Region != "" {
		memberArgs.Region = pulumi.StringPtr(spec.Region)
	}

	createdMember, err := loadbalancer.NewMember(
		ctx,
		strings.ToLower(memberName),
		memberArgs,
		pulumi.Provider(openstackProvider),
	)
	if err != nil {
		return errors.Wrap(err, "failed to create openstack load balancer member")
	}

	// Export required outputs (matching stack_outputs.proto fields).
	ctx.Export(OpMemberId, createdMember.ID())
	ctx.Export(OpName, createdMember.Name)
	ctx.Export(OpAddress, createdMember.Address)
	ctx.Export(OpProtocolPort, createdMember.ProtocolPort)
	ctx.Export(OpWeight, createdMember.Weight)
	ctx.Export(OpRegion, createdMember.Region)

	return nil
}
