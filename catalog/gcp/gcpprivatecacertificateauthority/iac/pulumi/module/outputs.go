package module

// Output keys must match the field names in outputs.proto -- the outputs
// transformer maps raw engine outputs onto the proto by name.
const (
	OpName                   = "name"
	OpCertificateAuthorityId = "certificate_authority_id"
	OpState                  = "state"
	OpPemCaCertificate       = "pem_ca_certificate"
	OpPemCaCertificates      = "pem_ca_certificates"
	OpCaCertificateAccessUrl = "ca_certificate_access_url"
	OpCrlAccessUrls          = "crl_access_urls"
)
