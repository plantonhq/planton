package module

const (
	// OpPrivateNetworkId is the exported stack output name that contains the
	// UUID of the created Scaleway Private Network. This is the primary
	// cross-resource reference consumed by 8+ downstream resource kinds.
	OpPrivateNetworkId = "private_network_id"

	// OpIpv4SubnetCidr is the exported stack output name that contains the
	// IPv4 CIDR of the subnet associated with this Private Network.
	// If the user specified ipv4_subnet in the spec, this reflects that value.
	// If omitted, it contains the CIDR auto-allocated by Scaleway's IPAM.
	OpIpv4SubnetCidr = "ipv4_subnet_cidr"
)
