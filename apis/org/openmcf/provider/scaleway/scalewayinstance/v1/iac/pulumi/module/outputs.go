package module

const (
	// OpServerId is the exported stack output name that contains the
	// zoned ID of the created Scaleway instance server.
	OpServerId = "server_id"

	// OpPublicIpAddress is the exported stack output name that contains
	// the public IPv4 address of the instance's Flexible IP. Empty string
	// if no public IP was configured (spec.public_ip is nil).
	OpPublicIpAddress = "public_ip_address"

	// OpPublicIpId is the exported stack output name that contains the
	// zoned ID of the Flexible IP resource. Empty string if no public IP
	// was configured.
	OpPublicIpId = "public_ip_id"

	// OpPrivateIpAddress is the exported stack output name that contains
	// the private IP address assigned to the instance on the attached
	// Private Network. Empty string if no Private Network was configured.
	OpPrivateIpAddress = "private_ip_address"
)
