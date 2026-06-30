package module

const (
	// OpLoadBalancerId is the exported ID of the Cloudflare Load Balancer.
	OpLoadBalancerId = "load_balancer_id"
	// OpLoadBalancerDnsRecordName is the exported hostname (DNS record name).
	OpLoadBalancerDnsRecordName = "load_balancer_dns_record_name"
	// OpLoadBalancerCnameTarget is the canonical CNAME target returned by Cloudflare.
	OpLoadBalancerCnameTarget = "load_balancer_cname_target"
)
