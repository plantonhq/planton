package module

const (
	// OpCertificateId is the exported output containing the uploaded
	// certificate's ID -- what per-hostname associations reference.
	OpCertificateId = "certificate_id"
	// OpZoneId is the exported output containing the zone the certificate
	// belongs to.
	OpZoneId = "zone_id"
	// OpExpiresOn is the exported output containing the certificate's
	// expiry timestamp (RFC3339).
	OpExpiresOn = "expires_on"
)
