package module

const (
	// OpClusterId is the exported stack output name for the Redis
	// cluster's zonal ID. Referenced by downstream resources.
	OpClusterId = "cluster_id"

	// OpPublicNetworkPort is the exported stack output name for the
	// cluster's public endpoint port. Zero when using Private Network.
	OpPublicNetworkPort = "public_network_port"

	// OpPublicNetworkIps is the exported stack output name for the
	// cluster's public endpoint IP addresses. Empty when using PN.
	OpPublicNetworkIps = "public_network_ips"

	// OpPrivateNetworkPort is the exported stack output name for the
	// cluster's Private Network endpoint port. Zero when not using PN.
	OpPrivateNetworkPort = "private_network_port"

	// OpPrivateNetworkIps is the exported stack output name for the
	// cluster's Private Network endpoint IPs. Empty when not using PN.
	OpPrivateNetworkIps = "private_network_ips"

	// OpCertificate is the exported stack output name for the TLS
	// certificate used to verify the Redis server. Empty when TLS
	// is disabled.
	OpCertificate = "certificate"
)
