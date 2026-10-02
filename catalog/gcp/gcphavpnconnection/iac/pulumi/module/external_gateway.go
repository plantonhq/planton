package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/compute"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// externalGateway creates the external VPN gateway resource that holds the
// peer device's public addresses -- only when the peer is an external
// device. Returns nil for a Google-to-Google connection, where the tunnels
// reference the other side's HA VPN gateway directly.
//
// The resource is free metadata and immutable except labels: a device's
// address change is a new external gateway, and the tunnels that reference
// it are recreated with it.
func externalGateway(ctx *pulumi.Context, locals *Locals, gcpProvider *gcp.Provider) (*compute.ExternalVpnGateway, error) {
	if !locals.IsExternalPeer {
		ctx.Export(OpExternalGatewaySelfLink, pulumi.String(""))
		return nil, nil
	}
	spec := locals.GcpHaVpnConnection.Spec
	external := spec.Peer.ExternalGateway

	interfaces := compute.ExternalVpnGatewayInterfaceArray{}
	for _, iface := range external.Interfaces {
		ifaceArgs := &compute.ExternalVpnGatewayInterfaceArgs{
			Id: pulumi.IntPtr(int(iface.Id)),
		}
		// Exactly one of the two is set (spec CEL); the other stays out of
		// the payload.
		if iface.IpAddress != "" {
			ifaceArgs.IpAddress = pulumi.StringPtr(iface.IpAddress)
		}
		if iface.Ipv6Address != "" {
			ifaceArgs.Ipv6Address = pulumi.StringPtr(iface.Ipv6Address)
		}
		interfaces = append(interfaces, ifaceArgs)
	}

	args := &compute.ExternalVpnGatewayArgs{
		Name:           pulumi.String(locals.ExternalGatewayName),
		RedundancyType: pulumi.StringPtr(external.RedundancyType),
		Interfaces:     interfaces,
		Labels:         pulumi.ToStringMap(locals.mergeLabels(external.Labels)),
	}
	if locals.Project != "" {
		args.Project = pulumi.String(locals.Project)
	}
	if external.Description != "" {
		args.Description = pulumi.StringPtr(external.Description)
	}
	if len(spec.ResourceManagerTags) > 0 {
		args.Params = &compute.ExternalVpnGatewayParamsArgs{
			ResourceManagerTags: pulumi.ToStringMap(spec.ResourceManagerTags),
		}
	}
	if spec.DeletionPolicy != "" {
		args.DeletionPolicy = pulumi.StringPtr(spec.DeletionPolicy)
	}

	created, err := compute.NewExternalVpnGateway(ctx, "external-gateway", args, pulumi.Provider(gcpProvider))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create external vpn gateway")
	}

	ctx.Export(OpExternalGatewaySelfLink, created.SelfLink)
	return created, nil
}
