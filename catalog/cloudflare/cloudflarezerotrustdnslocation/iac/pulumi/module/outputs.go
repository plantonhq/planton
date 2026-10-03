package module

const (
	// OpLocationId is the exported output containing the
	// Cloudflare-assigned UUID of the location.
	OpLocationId = "location_id"

	// OpDohSubdomain is the exported output containing the location's
	// unique DNS-over-HTTPS subdomain.
	OpDohSubdomain = "doh_subdomain"

	// OpIp is the exported output containing the IPv4 destination
	// assigned to the location's plain-DNS endpoint.
	OpIp = "ip"
)
