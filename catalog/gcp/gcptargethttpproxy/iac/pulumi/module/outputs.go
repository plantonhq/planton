package module

// Output keys must match the field names in outputs.proto — the outputs
// transformer maps raw engine outputs onto the proto by name.
const (
	OpSelfLink    = "self_link"
	OpProxyName   = "proxy_name"
	OpProxyId     = "proxy_id"
	OpFingerprint = "fingerprint"
	OpRegion      = "region"
)
