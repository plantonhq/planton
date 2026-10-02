package module

// Output keys must match the field names in outputs.proto — the outputs
// transformer maps raw engine outputs onto the proto by name.
const (
	OpSelfLink           = "self_link"
	OpBackendServiceName = "backend_service_name"
	OpGeneratedId        = "generated_id"
	OpFingerprint        = "fingerprint"
	OpRegion             = "region"
)
