package module

const (
	// OpFloatingIpId is the exported stack output name that contains the
	// Hetzner Cloud numeric ID of the created Floating IP.
	OpFloatingIpId = "floating_ip_id"

	// OpIpAddress is the exported stack output name that contains the
	// allocated IP address (single address for IPv4, first address in /64 for IPv6).
	OpIpAddress = "ip_address"

	// OpIpNetwork is the exported stack output name that contains the
	// allocated IPv6 /64 CIDR. Empty for IPv4 Floating IPs.
	OpIpNetwork = "ip_network"
)
