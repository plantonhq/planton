package module

// Output keys must match the field names in outputs.proto -- the outputs
// transformer maps raw engine outputs onto the proto by name.
const (
	OpName       = "name"
	OpSelfLink   = "self_link"
	OpFamily     = "family"
	OpDiskSizeGb = "disk_size_gb"
)
