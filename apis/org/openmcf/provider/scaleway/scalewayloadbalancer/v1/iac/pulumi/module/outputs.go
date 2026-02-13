package module

const (
	// OpLbId is the exported stack output name that contains the zoned ID
	// of the created Scaleway Load Balancer.
	OpLbId = "lb_id"

	// OpLbIpAddress is the exported stack output name that contains the
	// public IPv4 address of the Load Balancer's Flexible IP. This is the
	// address that clients connect to and that DNS records should point to.
	OpLbIpAddress = "lb_ip_address"

	// OpLbIpId is the exported stack output name that contains the zoned ID
	// of the Flexible IP resource. The IP has independent lifecycle and
	// survives LB replacement.
	OpLbIpId = "lb_ip_id"
)
