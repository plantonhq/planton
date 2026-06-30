package module

const (
	// OpLoadBalancerId is the exported stack output name that contains the
	// UUID of the created DigitalOcean Load Balancer.
	OpLoadBalancerId = "load_balancer_id"

	// OpIp is the exported stack output name that contains the public IP address.
	OpIp = "ip"

	// OpDnsName is the exported stack output name for the LB DNS/FQDN.
	// DigitalOcean does not expose an explicit DNS name; we export the Name as a best‑effort placeholder.
	OpDnsName = "dns_name"
)
