package module

const (
	// OpVpcId is the exported output name that contains the
	// UUID of the created DigitalOcean VPC.
	OpVpcId = "vpc_id"

	// OpIpRange is the exported output name that contains the VPC's IP
	// range in CIDR notation (covers the DigitalOcean-auto-assigned case).
	OpIpRange = "ip_range"

	// OpUrn is the exported output name that contains the VPC's
	// uniform resource name (URN).
	OpUrn = "urn"
)
