package module

import (
	"strings"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-openstack/sdk/v5/go/openstack"
	"github.com/pulumi/pulumi-openstack/sdk/v5/go/openstack/loadbalancer"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// loadBalancer provisions the OpenStack Octavia load balancer and exports outputs.
func loadBalancer(
	ctx *pulumi.Context,
	locals *Locals,
	openstackProvider *openstack.Provider,
) error {
	spec := locals.OpenStackLoadBalancer.Spec
	lbName := locals.OpenStackLoadBalancer.Metadata.Name

	lbArgs := &loadbalancer.LoadBalancerArgs{
		Name:        pulumi.String(lbName),
		VipSubnetId: pulumi.StringPtr(locals.VipSubnetId),
	}

	// Set vip_address if provided.
	if spec.VipAddress != "" {
		lbArgs.VipAddress = pulumi.StringPtr(spec.VipAddress)
	}

	// Set description if provided.
	if spec.Description != "" {
		lbArgs.Description = pulumi.StringPtr(spec.Description)
	}

	// Set admin_state_up if explicitly provided.
	if spec.AdminStateUp != nil {
		lbArgs.AdminStateUp = pulumi.BoolPtr(spec.GetAdminStateUp())
	}

	// Set flavor_id if provided.
	if spec.FlavorId != "" {
		lbArgs.FlavorId = pulumi.StringPtr(spec.FlavorId)
	}

	// Set tags if provided.
	if len(spec.Tags) > 0 {
		tags := make(pulumi.StringArray, len(spec.Tags))
		for i, tag := range spec.Tags {
			tags[i] = pulumi.String(tag)
		}
		lbArgs.Tags = tags
	}

	// Set region override if provided.
	if spec.Region != "" {
		lbArgs.Region = pulumi.StringPtr(spec.Region)
	}

	createdLb, err := loadbalancer.NewLoadBalancer(
		ctx,
		strings.ToLower(lbName),
		lbArgs,
		pulumi.Provider(openstackProvider),
	)
	if err != nil {
		return errors.Wrap(err, "failed to create openstack load balancer")
	}

	// Export required outputs (matching stack_outputs.proto fields).
	ctx.Export(OpLoadBalancerId, createdLb.ID())
	ctx.Export(OpName, createdLb.Name)
	ctx.Export(OpVipAddress, createdLb.VipAddress)
	ctx.Export(OpVipPortId, createdLb.VipPortId)
	ctx.Export(OpVipSubnetId, createdLb.VipSubnetId)
	ctx.Export(OpRegion, createdLb.Region)

	return nil
}
