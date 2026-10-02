package module

// Output keys must match the field names in outputs.proto -- the outputs
// transformer maps raw engine outputs onto the proto by name.
const (
	// OpGatewaySelfLink is what a GcpHaVpnConnection's gateway references.
	OpGatewaySelfLink = "gateway_self_link"
	// OpGatewayName is the gateway's name in GCP.
	OpGatewayName = "gateway_name"
	// OpRegion is the region the gateway and router live in.
	OpRegion = "region"
	// OpInterface0IpAddress is the public IP of gateway interface 0.
	OpInterface0IpAddress = "interface_0_ip_address"
	// OpInterface1IpAddress is the public IP of gateway interface 1.
	OpInterface1IpAddress = "interface_1_ip_address"
	// OpRouterName is what a GcpHaVpnConnection's router references.
	OpRouterName = "router_name"
	// OpRouterSelfLink is the Cloud Router's self link.
	OpRouterSelfLink = "router_self_link"
	// OpRouterAsn is the ASN the router speaks as.
	OpRouterAsn = "router_asn"
)
