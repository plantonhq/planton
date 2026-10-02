package module

import (
	"github.com/pkg/errors"
	gcphavpngatewayv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcphavpngateway/v1alpha1"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/compute"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/projects"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// haVpnGateway provisions the HA VPN gateway and the Cloud Router its
// tunnels terminate BGP on, as one node: Google allows many routers per
// network and region, so a VPN router beside a NAT router is the normal
// topology and nothing is shared between them.
//
// Nearly everything on both resources is immutable: the gateway's network,
// region, IP version, stack type, and interface pinning recreate it (and a
// recreated gateway has NEW public IPs); the router's ASN and
// encrypted-interconnect flag recreate the router. Labels, the router's
// description, and the BGP advertisement change in place. Optional inputs
// are sent only when set so the provider's defaults stay the provider's.
func haVpnGateway(ctx *pulumi.Context, locals *Locals, gcpProvider *gcp.Provider) error {
	spec := locals.GcpHaVpnGateway.Spec

	// Enable the Compute Engine API first so a fresh project works on the
	// first deploy. disable_on_destroy stays false: tearing down one gateway
	// must never disable the API for everything else in the project.
	serviceArgs := &projects.ServiceArgs{
		Service:                  pulumi.String("compute.googleapis.com"),
		DisableDependentServices: pulumi.BoolPtr(true),
		DisableOnDestroy:         pulumi.BoolPtr(false),
	}
	if spec.ProjectId.GetValue() != "" {
		serviceArgs.Project = pulumi.String(spec.ProjectId.GetValue())
	}
	createdProjectService, err := projects.NewService(ctx,
		"havpngateway-compute.googleapis.com", serviceArgs, pulumi.Provider(gcpProvider))
	if err != nil {
		return errors.Wrap(err, "failed to enable compute.googleapis.com api")
	}

	// --- The gateway -------------------------------------------------------

	gatewayArgs := &compute.HaVpnGatewayArgs{
		Name:    pulumi.String(locals.GatewayName),
		Region:  pulumi.String(spec.Region),
		Network: pulumi.String(spec.Network.GetValue()),
		Labels:  pulumi.ToStringMap(locals.GcpLabels),
	}
	// An empty project falls back to the provider's default project -- the
	// ambient-project contract every GCP kind honors.
	if spec.ProjectId.GetValue() != "" {
		gatewayArgs.Project = pulumi.String(spec.ProjectId.GetValue())
	}
	if spec.Description != "" {
		gatewayArgs.Description = pulumi.StringPtr(spec.Description)
	}
	if spec.GatewayIpVersion != "" {
		gatewayArgs.GatewayIpVersion = pulumi.StringPtr(spec.GatewayIpVersion)
	}
	if spec.StackType != "" {
		gatewayArgs.StackType = pulumi.StringPtr(spec.StackType)
	}
	// HA VPN over Interconnect: pin each interface to its VLAN attachment.
	// Empty means the ordinary internet-facing gateway, whose interfaces the
	// API fills in with public IPs (the block is Optional+Computed, so it
	// stays out of the payload when unset).
	if len(spec.VpnInterfaces) > 0 {
		interfaces := compute.HaVpnGatewayVpnInterfaceArray{}
		for _, vpnInterface := range spec.VpnInterfaces {
			interfaces = append(interfaces, &compute.HaVpnGatewayVpnInterfaceArgs{
				Id:                     pulumi.IntPtr(int(vpnInterface.Id)),
				InterconnectAttachment: pulumi.StringPtr(vpnInterface.InterconnectAttachment),
			})
		}
		gatewayArgs.VpnInterfaces = interfaces
	}
	// Create-time resource-manager tags (org policy / IAM conditions).
	if len(spec.ResourceManagerTags) > 0 {
		gatewayArgs.Params = &compute.HaVpnGatewayParamsArgs{
			ResourceManagerTags: pulumi.ToStringMap(spec.ResourceManagerTags),
		}
	}
	// Empty defers to the provider default (DELETE); applied to the gateway
	// and the router alike.
	if spec.DeletionPolicy != "" {
		gatewayArgs.DeletionPolicy = pulumi.StringPtr(spec.DeletionPolicy)
	}

	createdGateway, err := compute.NewHaVpnGateway(ctx, "gateway", gatewayArgs,
		pulumi.Provider(gcpProvider), pulumi.DependsOn([]pulumi.Resource{createdProjectService}))
	if err != nil {
		return errors.Wrap(err, "failed to create ha vpn gateway")
	}

	// --- The router --------------------------------------------------------

	router := spec.Router
	routerArgs := &compute.RouterArgs{
		Name:    pulumi.String(locals.RouterName),
		Region:  pulumi.String(spec.Region),
		Network: pulumi.String(spec.Network.GetValue()),
		// Dedicates the router to encrypted VLAN attachments (HA VPN over
		// Interconnect). Immutable; an encrypted router cannot be converted.
		EncryptedInterconnectRouter: pulumi.BoolPtr(router.EncryptedInterconnectRouter),
		// The BGP block is required by the spec: an HA VPN router exists to
		// run BGP and every tunnel session needs the ASN.
		Bgp: routerBgpArgs(router.Bgp),
	}
	if spec.ProjectId.GetValue() != "" {
		routerArgs.Project = pulumi.String(spec.ProjectId.GetValue())
	}
	if router.Description != "" {
		routerArgs.Description = pulumi.StringPtr(router.Description)
	}
	if len(router.ResourceManagerTags) > 0 {
		routerArgs.Params = &compute.RouterParamsArgs{
			ResourceManagerTags: pulumi.ToStringMap(router.ResourceManagerTags),
		}
	}
	if spec.DeletionPolicy != "" {
		routerArgs.DeletionPolicy = pulumi.StringPtr(spec.DeletionPolicy)
	}

	createdRouter, err := compute.NewRouter(ctx, "router", routerArgs,
		pulumi.Provider(gcpProvider), pulumi.DependsOn([]pulumi.Resource{createdProjectService}))
	if err != nil {
		return errors.Wrap(err, "failed to create router")
	}

	// --- Outputs -----------------------------------------------------------

	ctx.Export(OpGatewaySelfLink, createdGateway.SelfLink)
	ctx.Export(OpGatewayName, createdGateway.Name)
	ctx.Export(OpRegion, createdGateway.Region)
	// The two interface addresses are API-assigned for an internet-facing
	// gateway; the list is read back with the interface ids Google
	// assigns (0 and 1), so each address is picked by id, not position.
	ctx.Export(OpInterface0IpAddress, interfaceIpAddress(createdGateway, 0))
	ctx.Export(OpInterface1IpAddress, interfaceIpAddress(createdGateway, 1))
	ctx.Export(OpRouterName, createdRouter.Name)
	ctx.Export(OpRouterSelfLink, createdRouter.SelfLink)
	ctx.Export(OpRouterAsn, pulumi.Int(int(router.Bgp.Asn)))

	return nil
}

// routerBgpArgs maps the spec's BGP message onto the provider's block.
// advertise_mode empty means DEFAULT; custom groups and ranges are legal
// only in CUSTOM mode (spec CELs mirror the provider). identifier_range is
// Optional+Computed and stays out of the payload when empty so it does not
// fight the API's assigned value.
func routerBgpArgs(bgp *gcphavpngatewayv1alpha1.GcpHaVpnGatewayRouterBgp) *compute.RouterBgpArgs {
	args := &compute.RouterBgpArgs{
		Asn: pulumi.Int(int(bgp.Asn)),
	}
	if bgp.AdvertiseMode != "" {
		args.AdvertiseMode = pulumi.StringPtr(bgp.AdvertiseMode)
	}
	if len(bgp.AdvertisedGroups) > 0 {
		args.AdvertisedGroups = pulumi.ToStringArray(bgp.AdvertisedGroups)
	}
	if len(bgp.AdvertisedIpRanges) > 0 {
		ranges := compute.RouterBgpAdvertisedIpRangeArray{}
		for _, advertised := range bgp.AdvertisedIpRanges {
			rangeArgs := &compute.RouterBgpAdvertisedIpRangeArgs{
				Range: pulumi.String(advertised.Range),
			}
			if advertised.Description != "" {
				rangeArgs.Description = pulumi.StringPtr(advertised.Description)
			}
			ranges = append(ranges, rangeArgs)
		}
		args.AdvertisedIpRanges = ranges
	}
	if bgp.KeepaliveInterval > 0 {
		args.KeepaliveInterval = pulumi.IntPtr(int(bgp.KeepaliveInterval))
	}
	if bgp.IdentifierRange != "" {
		args.IdentifierRange = pulumi.StringPtr(bgp.IdentifierRange)
	}
	return args
}

// interfaceIpAddress picks the public address of the interface with the
// given id from the gateway's read-back interface list; empty when the
// interface is Interconnect-backed (no public address) or not yet assigned.
func interfaceIpAddress(gateway *compute.HaVpnGateway, id int) pulumi.StringOutput {
	return gateway.VpnInterfaces.ApplyT(func(interfaces []compute.HaVpnGatewayVpnInterface) string {
		for _, vpnInterface := range interfaces {
			if vpnInterface.Id != nil && *vpnInterface.Id == id && vpnInterface.IpAddress != nil {
				return *vpnInterface.IpAddress
			}
		}
		return ""
	}).(pulumi.StringOutput)
}
