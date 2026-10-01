package module

// Output keys must match the field names in outputs.proto -- the outputs
// transformer maps raw engine outputs onto the proto by name.
const (
	OpProjectId = "project_id"
	OpName      = "name"
	OpUid       = "uid"
)
