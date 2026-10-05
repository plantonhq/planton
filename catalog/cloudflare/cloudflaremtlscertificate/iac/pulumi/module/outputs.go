package module

const (
	// OpCertificateId is the exported output containing the uploaded
	// certificate's ID -- what consumers reference.
	OpCertificateId = "certificate_id"
	// OpExpiresOn is the exported output containing the certificate's
	// expiry timestamp (RFC3339).
	OpExpiresOn = "expires_on"
	// OpSerialNumber is the exported output containing the certificate's
	// serial number.
	OpSerialNumber = "serial_number"
)
