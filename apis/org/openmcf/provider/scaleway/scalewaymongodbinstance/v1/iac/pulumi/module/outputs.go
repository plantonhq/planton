package module

const (
	// OpInstanceId is the exported stack output name for the MongoDB
	// instance's regional ID. Referenced by downstream resources
	// (snapshots, monitoring, management automation).
	OpInstanceId = "instance_id"

	// OpPublicDnsRecord is the exported stack output name for the
	// instance's public endpoint DNS record. Empty if private-only.
	OpPublicDnsRecord = "public_dns_record"

	// OpPublicPort is the exported stack output name for the instance's
	// public endpoint port number. Zero if private-only.
	OpPublicPort = "public_port"

	// OpPrivateDnsRecords is the exported stack output name for the
	// instance's Private Network endpoint DNS records. Empty if no PN.
	OpPrivateDnsRecords = "private_dns_records"

	// OpPrivateIps is the exported stack output name for the instance's
	// Private Network endpoint IP addresses. Empty if no PN.
	OpPrivateIps = "private_ips"

	// OpPrivatePort is the exported stack output name for the instance's
	// Private Network endpoint port. Zero if no PN.
	OpPrivatePort = "private_port"

	// OpTlsCertificate is the exported stack output name for the TLS
	// certificate used to verify the database server's identity.
	OpTlsCertificate = "tls_certificate"
)
