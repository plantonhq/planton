package module

const (
	// OpInstanceId is the exported stack output name for the RDB
	// instance's regional ID. Referenced by downstream resources
	// (read replicas, monitoring, backup tools).
	OpInstanceId = "instance_id"

	// OpEndpointIp is the exported stack output name for the instance's
	// public endpoint IP address.
	OpEndpointIp = "endpoint_ip"

	// OpEndpointPort is the exported stack output name for the instance's
	// public endpoint port number.
	OpEndpointPort = "endpoint_port"

	// OpPrivateEndpointIp is the exported stack output name for the
	// instance's Private Network endpoint IP. Empty if no PN is attached.
	OpPrivateEndpointIp = "private_endpoint_ip"

	// OpPrivateEndpointPort is the exported stack output name for the
	// instance's Private Network endpoint port. Zero if no PN is attached.
	OpPrivateEndpointPort = "private_endpoint_port"

	// OpCertificate is the exported stack output name for the TLS
	// certificate used to verify the database server's identity.
	OpCertificate = "certificate"
)
