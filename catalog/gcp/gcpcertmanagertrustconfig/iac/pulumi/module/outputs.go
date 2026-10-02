package module

// Output keys must match the field names in outputs.proto -- the outputs
// transformer maps raw engine outputs onto the proto by name.
const (
	OpTrustConfigId   = "trust_config_id"
	OpTrustConfigName = "trust_config_name"
	OpLocation        = "location"
)
