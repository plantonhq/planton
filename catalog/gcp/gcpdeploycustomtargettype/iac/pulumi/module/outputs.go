package module

// Output keys must match the field names in outputs.proto -- the outputs
// transformer maps raw engine outputs onto the proto by name.
const (
	OpName               = "name"
	OpCustomTargetTypeId = "custom_target_type_id"
	OpUid                = "uid"
)
