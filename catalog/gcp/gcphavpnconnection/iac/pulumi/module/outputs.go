package module

// Output keys must match the field names in outputs.proto -- the outputs
// transformer maps raw engine outputs onto the proto by name. The four
// lists are index-aligned with spec.tunnels.
const (
	// OpTunnelSelfLinks is the tunnels' self links, in spec order.
	OpTunnelSelfLinks = "tunnel_self_links"
	// OpTunnelNames is the tunnels' names in GCP, in spec order.
	OpTunnelNames = "tunnel_names"
	// OpRouterInterfaceNames is the router interfaces' names, in spec order.
	OpRouterInterfaceNames = "router_interface_names"
	// OpBgpPeerNames is the BGP peers' names, in spec order.
	OpBgpPeerNames = "bgp_peer_names"
	// OpExternalGatewaySelfLink is the external VPN gateway's self link, or
	// empty for a Google-to-Google connection.
	OpExternalGatewaySelfLink = "external_gateway_self_link"
	// OpGatewaySelfLink is the resolved gateway reference.
	OpGatewaySelfLink = "gateway_self_link"
	// OpRouterName is the resolved router reference.
	OpRouterName = "router_name"
)
