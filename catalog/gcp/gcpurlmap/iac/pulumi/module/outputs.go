package module

// Output keys must match the field names in outputs.proto — the outputs
// transformer maps raw engine outputs onto the proto by name.
const (
	OpSelfLink    = "self_link"
	OpUrlMapName  = "url_map_name"
	OpMapId       = "map_id"
	OpFingerprint = "fingerprint"
	OpRegion      = "region"
)
