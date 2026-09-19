package module

import (
	"fmt"

	"github.com/pkg/errors"
	gcphavpnconnectionv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcphavpnconnection/v1alpha1"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/compute"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// tunnels provisions, for every tunnels[] entry, the IPsec tunnel and the
// BGP session inside it: a Cloud Router interface bound to the tunnel and
// a BGP peer on that interface -- Google's own 1:1:1 shape.
//
// Every tunnel argument except labels is immutable (a new secret, cipher,
// or interface pairing recreates the tunnel); the interface is fully
// immutable; the peer's BGP policy changes in place. The external gateway,
// when present, is created first and every tunnel depends on it through
// its self link.
func tunnels(ctx *pulumi.Context, locals *Locals, gcpProvider *gcp.Provider, createdExternalGateway *compute.ExternalVpnGateway) error {
	spec := locals.GcpHaVpnConnection.Spec

	tunnelSelfLinks := pulumi.StringArray{}
	tunnelNames := pulumi.StringArray{}
	interfaceNames := pulumi.StringArray{}
	peerNames := pulumi.StringArray{}

	for index, tunnel := range spec.Tunnels {
		createdTunnel, err := vpnTunnel(ctx, locals, gcpProvider, tunnel, createdExternalGateway)
		if err != nil {
			return errors.Wrapf(err, "failed to create tunnel %s", tunnel.Name)
		}

		createdInterface, err := routerInterface(ctx, locals, gcpProvider, index, tunnel, createdTunnel)
		if err != nil {
			return errors.Wrapf(err, "failed to create router interface for tunnel %s", tunnel.Name)
		}

		createdPeer, err := routerPeer(ctx, locals, gcpProvider, index, tunnel, createdInterface)
		if err != nil {
			return errors.Wrapf(err, "failed to create bgp peer for tunnel %s", tunnel.Name)
		}

		tunnelSelfLinks = append(tunnelSelfLinks, createdTunnel.SelfLink)
		tunnelNames = append(tunnelNames, createdTunnel.Name)
		interfaceNames = append(interfaceNames, createdInterface.Name)
		peerNames = append(peerNames, createdPeer.Name)
	}

	ctx.Export(OpTunnelSelfLinks, tunnelSelfLinks)
	ctx.Export(OpTunnelNames, tunnelNames)
	ctx.Export(OpRouterInterfaceNames, interfaceNames)
	ctx.Export(OpBgpPeerNames, peerNames)
	ctx.Export(OpGatewaySelfLink, pulumi.String(locals.Gateway))
	ctx.Export(OpRouterName, pulumi.String(locals.Router))

	return nil
}

// vpnTunnel creates one IPsec tunnel from a gateway interface to the peer.
// Exactly one of peer_external_gateway / peer_gcp_gateway is set (the
// provider's ConflictsWith, the spec's exactly-one CEL). ike_version is
// always sent (2 when unset -- the provider default, made explicit so the
// spec is the single source of truth); traffic selectors are
// Optional+Computed and sent only when set; the cipher suite only when
// declared.
func vpnTunnel(ctx *pulumi.Context, locals *Locals, gcpProvider *gcp.Provider,
	tunnel *gcphavpnconnectionv1alpha1.GcpHaVpnConnectionTunnel,
	createdExternalGateway *compute.ExternalVpnGateway) (*compute.VPNTunnel, error) {
	spec := locals.GcpHaVpnConnection.Spec

	args := &compute.VPNTunnelArgs{
		Name:                pulumi.String(tunnel.Name),
		Region:              pulumi.String(locals.Region),
		VpnGateway:          pulumi.String(locals.Gateway),
		VpnGatewayInterface: pulumi.IntPtr(int(tunnel.VpnGatewayInterface)),
		Router:              pulumi.String(locals.Router),
		// ToSecret marks it encrypted in Pulumi state.
		SharedSecret: pulumi.ToSecret(pulumi.String(tunnel.SharedSecret)).(pulumi.StringOutput),
		IkeVersion:   pulumi.IntPtr(ikeVersion(tunnel.IkeVersion)),
		Labels:       pulumi.ToStringMap(locals.mergeLabels(tunnel.Labels)),
	}
	if locals.Project != "" {
		args.Project = pulumi.String(locals.Project)
	}
	if locals.IsExternalPeer {
		args.PeerExternalGateway = createdExternalGateway.SelfLink
		args.PeerExternalGatewayInterface = pulumi.IntPtr(int(tunnel.GetPeerExternalGatewayInterface()))
	} else {
		// Google pairs interfaces itself for a Google peer (interface 0 to
		// interface 0), so no peer interface id is sent.
		args.PeerGcpGateway = pulumi.String(spec.Peer.GcpGateway.GetValue())
	}
	if tunnel.Description != "" {
		args.Description = pulumi.StringPtr(tunnel.Description)
	}
	if len(tunnel.LocalTrafficSelector) > 0 {
		args.LocalTrafficSelectors = pulumi.ToStringArray(tunnel.LocalTrafficSelector)
	}
	if len(tunnel.RemoteTrafficSelector) > 0 {
		args.RemoteTrafficSelectors = pulumi.ToStringArray(tunnel.RemoteTrafficSelector)
	}
	if tunnel.CipherSuite != nil {
		args.CipherSuite = cipherSuiteArgs(tunnel.CipherSuite)
	}
	if len(spec.ResourceManagerTags) > 0 {
		args.Params = &compute.VPNTunnelParamsArgs{
			ResourceManagerTags: pulumi.ToStringMap(spec.ResourceManagerTags),
		}
	}
	if spec.DeletionPolicy != "" {
		args.DeletionPolicy = pulumi.StringPtr(spec.DeletionPolicy)
	}

	return compute.NewVPNTunnel(ctx, fmt.Sprintf("tunnel-%s", tunnel.Name), args, pulumi.Provider(gcpProvider))
}

// routerInterface creates the Cloud Router interface bound to the tunnel:
// the Google end of the BGP session's link-local /30. ip_range and
// ip_version are Optional+Computed and sent only when set (Google assigns
// them for IPv6-only sessions).
func routerInterface(ctx *pulumi.Context, locals *Locals, gcpProvider *gcp.Provider, index int,
	tunnel *gcphavpnconnectionv1alpha1.GcpHaVpnConnectionTunnel,
	createdTunnel *compute.VPNTunnel) (*compute.RouterInterface, error) {
	spec := locals.GcpHaVpnConnection.Spec
	session := tunnel.BgpSession

	args := &compute.RouterInterfaceArgs{
		Name:      pulumi.String(locals.SessionNames[index]),
		Region:    pulumi.String(locals.Region),
		Router:    pulumi.String(locals.Router),
		VpnTunnel: createdTunnel.SelfLink,
	}
	if locals.Project != "" {
		args.Project = pulumi.String(locals.Project)
	}
	if session.InterfaceIpRange != "" {
		args.IpRange = pulumi.StringPtr(session.InterfaceIpRange)
	}
	if session.IpVersion != nil && *session.IpVersion != "" {
		args.IpVersion = pulumi.StringPtr(*session.IpVersion)
	}
	if spec.DeletionPolicy != "" {
		args.DeletionPolicy = pulumi.StringPtr(spec.DeletionPolicy)
	}

	return compute.NewRouterInterface(ctx, fmt.Sprintf("interface-%s", tunnel.Name), args,
		pulumi.Provider(gcpProvider), pulumi.Parent(createdTunnel))
}

// routerPeer creates the BGP peer on the tunnel's interface. `enable` is
// always sent (true when unset -- the provider default made explicit);
// the optional numerics and addresses are sent only when set so 0 is a
// real priority and Google keeps assigning the addresses it owns; the
// per-session advertisement overrides ride the same CUSTOM-mode rules as
// the router's.
func routerPeer(ctx *pulumi.Context, locals *Locals, gcpProvider *gcp.Provider, index int,
	tunnel *gcphavpnconnectionv1alpha1.GcpHaVpnConnectionTunnel,
	createdInterface *compute.RouterInterface) (*compute.RouterPeer, error) {
	spec := locals.GcpHaVpnConnection.Spec
	session := tunnel.BgpSession

	args := &compute.RouterPeerArgs{
		Name:       pulumi.String(locals.SessionNames[index]),
		Region:     pulumi.String(locals.Region),
		Router:     pulumi.String(locals.Router),
		Interface:  createdInterface.Name,
		PeerAsn:    pulumi.Int(int(session.PeerAsn)),
		Enable:     pulumi.BoolPtr(enable(session.Enable)),
		EnableIpv6: pulumi.BoolPtr(session.EnableIpv6),
	}
	if locals.Project != "" {
		args.Project = pulumi.String(locals.Project)
	}
	if session.PeerIpAddress != nil && *session.PeerIpAddress != "" {
		args.PeerIpAddress = pulumi.StringPtr(*session.PeerIpAddress)
	}
	if session.AdvertisedRoutePriority != nil {
		args.AdvertisedRoutePriority = pulumi.IntPtr(int(*session.AdvertisedRoutePriority))
	}
	if session.AdvertiseMode != "" {
		args.AdvertiseMode = pulumi.StringPtr(session.AdvertiseMode)
	}
	if len(session.AdvertisedGroups) > 0 {
		args.AdvertisedGroups = pulumi.ToStringArray(session.AdvertisedGroups)
	}
	if len(session.AdvertisedIpRanges) > 0 {
		ranges := compute.RouterPeerAdvertisedIpRangeArray{}
		for _, advertised := range session.AdvertisedIpRanges {
			rangeArgs := &compute.RouterPeerAdvertisedIpRangeArgs{Range: pulumi.String(advertised.Range)}
			if advertised.Description != "" {
				rangeArgs.Description = pulumi.StringPtr(advertised.Description)
			}
			ranges = append(ranges, rangeArgs)
		}
		args.AdvertisedIpRanges = ranges
	}
	if session.EnableIpv4 != nil {
		args.EnableIpv4 = pulumi.BoolPtr(*session.EnableIpv4)
	}
	if session.Ipv6NexthopAddress != nil && *session.Ipv6NexthopAddress != "" {
		args.Ipv6NexthopAddress = pulumi.StringPtr(*session.Ipv6NexthopAddress)
	}
	if session.PeerIpv6NexthopAddress != nil && *session.PeerIpv6NexthopAddress != "" {
		args.PeerIpv6NexthopAddress = pulumi.StringPtr(*session.PeerIpv6NexthopAddress)
	}
	if session.Ipv4NexthopAddress != nil && *session.Ipv4NexthopAddress != "" {
		args.Ipv4NexthopAddress = pulumi.StringPtr(*session.Ipv4NexthopAddress)
	}
	if session.PeerIpv4NexthopAddress != nil && *session.PeerIpv4NexthopAddress != "" {
		args.PeerIpv4NexthopAddress = pulumi.StringPtr(*session.PeerIpv4NexthopAddress)
	}
	if len(session.CustomLearnedIpRanges) > 0 {
		ranges := compute.RouterPeerCustomLearnedIpRangeArray{}
		for _, learned := range session.CustomLearnedIpRanges {
			ranges = append(ranges, &compute.RouterPeerCustomLearnedIpRangeArgs{Range: pulumi.String(learned.Range)})
		}
		args.CustomLearnedIpRanges = ranges
	}
	if session.CustomLearnedRoutePriority != nil {
		args.CustomLearnedRoutePriority = pulumi.IntPtr(int(*session.CustomLearnedRoutePriority))
	}
	if session.Bfd != nil {
		bfdArgs := &compute.RouterPeerBfdArgs{
			SessionInitializationMode: pulumi.String(session.Bfd.SessionInitializationMode),
		}
		if session.Bfd.MinReceiveInterval > 0 {
			bfdArgs.MinReceiveInterval = pulumi.IntPtr(int(session.Bfd.MinReceiveInterval))
		}
		if session.Bfd.MinTransmitInterval > 0 {
			bfdArgs.MinTransmitInterval = pulumi.IntPtr(int(session.Bfd.MinTransmitInterval))
		}
		if session.Bfd.Multiplier > 0 {
			bfdArgs.Multiplier = pulumi.IntPtr(int(session.Bfd.Multiplier))
		}
		args.Bfd = bfdArgs
	}
	// The MD5 key rides the peer: the provider inserts it into the router's
	// key table under this name and attaches it to the session (Google
	// requires each key to be used by exactly one session).
	if session.Md5AuthenticationKey != nil {
		args.Md5AuthenticationKey = &compute.RouterPeerMd5AuthenticationKeyArgs{
			Name: pulumi.String(locals.Md5KeyNames[index]),
			// ToSecret marks it encrypted in Pulumi state.
			Key: pulumi.ToSecret(pulumi.String(session.Md5AuthenticationKey.Key)).(pulumi.StringOutput),
		}
	}
	if len(session.ImportPolicies) > 0 {
		args.ImportPolicies = pulumi.ToStringArray(session.ImportPolicies)
	}
	if len(session.ExportPolicies) > 0 {
		args.ExportPolicies = pulumi.ToStringArray(session.ExportPolicies)
	}
	if spec.DeletionPolicy != "" {
		args.DeletionPolicy = pulumi.StringPtr(spec.DeletionPolicy)
	}

	return compute.NewRouterPeer(ctx, fmt.Sprintf("peer-%s", tunnel.Name), args,
		pulumi.Provider(gcpProvider), pulumi.Parent(createdInterface))
}

// cipherSuiteArgs maps the spec's cipher restriction onto the provider's
// block; empty lists stay out so Google's defaults apply per class.
func cipherSuiteArgs(suite *gcphavpnconnectionv1alpha1.GcpHaVpnConnectionCipherSuite) *compute.VPNTunnelCipherSuiteArgs {
	args := &compute.VPNTunnelCipherSuiteArgs{}
	if p1 := suite.Phase1; p1 != nil {
		phase1 := &compute.VPNTunnelCipherSuitePhase1Args{}
		if len(p1.Encryption) > 0 {
			phase1.Encryptions = pulumi.ToStringArray(p1.Encryption)
		}
		if len(p1.Integrity) > 0 {
			phase1.Integrities = pulumi.ToStringArray(p1.Integrity)
		}
		if len(p1.Prf) > 0 {
			phase1.Prves = pulumi.ToStringArray(p1.Prf)
		}
		if len(p1.Dh) > 0 {
			phase1.Dhs = pulumi.ToStringArray(p1.Dh)
		}
		args.Phase1 = phase1
	}
	if p2 := suite.Phase2; p2 != nil {
		phase2 := &compute.VPNTunnelCipherSuitePhase2Args{}
		if len(p2.Encryption) > 0 {
			phase2.Encryptions = pulumi.ToStringArray(p2.Encryption)
		}
		if len(p2.Integrity) > 0 {
			phase2.Integrities = pulumi.ToStringArray(p2.Integrity)
		}
		if len(p2.Pfs) > 0 {
			phase2.Pfs = pulumi.ToStringArray(p2.Pfs)
		}
		args.Phase2 = phase2
	}
	return args
}

// ikeVersion resolves the tri-state optional to the wire value: the spec's
// value when set, the provider's default (2) when not.
func ikeVersion(value *int32) int {
	if value == nil || *value == 0 {
		return 2
	}
	return int(*value)
}

// enable resolves the tri-state optional to the wire value: the spec's
// value when set, the provider's default (true) when not.
func enable(value *bool) bool {
	if value == nil {
		return true
	}
	return *value
}
