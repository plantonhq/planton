package module

import (
	"strings"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-openstack/sdk/v5/go/openstack"
	"github.com/pulumi/pulumi-openstack/sdk/v5/go/openstack/networking"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// router provisions the OpenStack Neutron router and exports outputs.
func router(
	ctx *pulumi.Context,
	locals *Locals,
	openstackProvider *openstack.Provider,
) error {
	spec := locals.OpenStackRouter.Spec
	routerName := locals.OpenStackRouter.Metadata.Name

	routerArgs := &networking.RouterArgs{
		Name: pulumi.String(routerName),
	}

	// Set external_network_id if provided (optional FK).
	if locals.ExternalNetworkId != "" {
		routerArgs.ExternalNetworkId = pulumi.StringPtr(locals.ExternalNetworkId)
	}

	// Set admin_state_up if explicitly provided (default is handled by middleware).
	if spec.AdminStateUp != nil {
		routerArgs.AdminStateUp = pulumi.BoolPtr(spec.GetAdminStateUp())
	}

	// Set enable_snat if explicitly provided.
	// Only valid when external_network_id is configured (enforced by CEL validation).
	if spec.EnableSnat != nil {
		routerArgs.EnableSnat = pulumi.BoolPtr(spec.GetEnableSnat())
	}

	// Set distributed if explicitly provided.
	// This is a create-time setting -- cannot be changed after creation.
	if spec.Distributed != nil {
		routerArgs.Distributed = pulumi.BoolPtr(spec.GetDistributed())
	}

	// Set external_fixed_ips if provided.
	// Only valid when external_network_id is configured (enforced by CEL validation).
	if len(spec.ExternalFixedIps) > 0 {
		fixedIps := make(networking.RouterExternalFixedIpArray, len(spec.ExternalFixedIps))
		for i, fip := range spec.ExternalFixedIps {
			args := &networking.RouterExternalFixedIpArgs{}
			if fip.SubnetId != "" {
				args.SubnetId = pulumi.StringPtr(fip.SubnetId)
			}
			if fip.IpAddress != "" {
				args.IpAddress = pulumi.StringPtr(fip.IpAddress)
			}
			fixedIps[i] = args
		}
		routerArgs.ExternalFixedIps = fixedIps
	}

	// Set description if provided.
	if spec.Description != "" {
		routerArgs.Description = pulumi.StringPtr(spec.Description)
	}

	// Set tags if provided.
	if len(spec.Tags) > 0 {
		tags := make(pulumi.StringArray, len(spec.Tags))
		for i, tag := range spec.Tags {
			tags[i] = pulumi.String(tag)
		}
		routerArgs.Tags = tags
	}

	// Set region override if provided.
	if spec.Region != "" {
		routerArgs.Region = pulumi.StringPtr(spec.Region)
	}

	createdRouter, err := networking.NewRouter(
		ctx,
		strings.ToLower(routerName),
		routerArgs,
		pulumi.Provider(openstackProvider),
	)
	if err != nil {
		return errors.Wrap(err, "failed to create openstack router")
	}

	// Export required outputs (matching stack_outputs.proto fields).
	ctx.Export(OpRouterId, createdRouter.ID())
	ctx.Export(OpName, createdRouter.Name)
	ctx.Export(OpExternalNetworkId, createdRouter.ExternalNetworkId)
	ctx.Export(OpRegion, createdRouter.Region)

	// Extract the primary external gateway IP from the computed external_fixed_ips.
	// This is a convenience output -- the first IP from the external gateway allocation.
	// Empty string when no external gateway is configured.
	externalGatewayIp := createdRouter.ExternalFixedIps.ApplyT(
		func(fips []networking.RouterExternalFixedIp) string {
			if len(fips) > 0 && fips[0].IpAddress != nil {
				return *fips[0].IpAddress
			}
			return ""
		},
	).(pulumi.StringOutput)
	ctx.Export(OpExternalGatewayIp, externalGatewayIp)

	return nil
}
