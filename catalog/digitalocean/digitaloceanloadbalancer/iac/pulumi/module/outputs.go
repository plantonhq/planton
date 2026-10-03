package module

// Output keys — exactly the DigitalOceanLoadBalancerOutputs
// contract, identical across both provisioners.
const (
	// OpLoadBalancerId is the exported output containing the balancer UUID.
	OpLoadBalancerId = "load_balancer_id"
	// OpIp is the exported output containing the public IPv4 address.
	OpIp = "ip"
	// OpUrn is the exported output with the balancer's uniform resource
	// name ("do:loadbalancer:<id>").
	OpUrn = "urn"
	// OpIpv6 is the exported output with the IPv6 address (populated
	// when network_stack is DUALSTACK).
	OpIpv6 = "ipv6"
)
