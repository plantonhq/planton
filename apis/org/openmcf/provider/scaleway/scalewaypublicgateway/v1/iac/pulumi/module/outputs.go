package module

const (
	// OpGatewayId is the exported stack output name that contains the
	// zoned ID of the created Scaleway Public Gateway.
	OpGatewayId = "gateway_id"

	// OpPublicIpAddress is the exported stack output name that contains
	// the public IPv4 address assigned to the gateway. This is the address
	// that external traffic sees and that NAT masquerade uses as the source.
	OpPublicIpAddress = "public_ip_address"

	// OpPublicIpId is the exported stack output name that contains the
	// zoned ID of the Flexible IP resource. Useful for IP reassignment
	// or independent management.
	OpPublicIpId = "public_ip_id"

	// OpGatewayNetworkId is the exported stack output name that contains
	// the zoned ID of the GatewayNetwork attachment (the binding between
	// the gateway and the Private Network).
	OpGatewayNetworkId = "gateway_network_id"
)
