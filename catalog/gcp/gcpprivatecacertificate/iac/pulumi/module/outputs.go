package module

// Output keys must match the field names in outputs.proto -- the outputs
// transformer maps raw engine outputs onto the proto by name.
const (
	OpName                       = "name"
	OpCertificateId              = "certificate_id"
	OpPemCertificate             = "pem_certificate"
	OpPemCertificateChain        = "pem_certificate_chain"
	OpIssuerCertificateAuthority = "issuer_certificate_authority"
)
