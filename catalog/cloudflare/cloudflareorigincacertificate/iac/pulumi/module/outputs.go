package module

const (
	// OpCertificateId is the exported output containing the certificate id.
	OpCertificateId = "certificate_id"
	// OpCertificate is the exported output containing the issued certificate (PEM).
	OpCertificate = "certificate"
	// OpPrivateKey is the exported output containing the (sensitive) generated
	// private key; empty when a user-supplied CSR was used.
	OpPrivateKey = "private_key"
	// OpExpiresOn is the exported output containing the expiry timestamp.
	OpExpiresOn = "expires_on"
)
