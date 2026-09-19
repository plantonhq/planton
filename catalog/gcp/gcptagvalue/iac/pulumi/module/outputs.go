package module

// Output keys must match the field names in outputs.proto -- the outputs
// transformer maps raw engine outputs onto the proto by name.
const (
	// OpName is the value's resource name, tagValues/{numeric_id}.
	OpName = "name"
	// OpNamespacedName is {parent}/{key_short_name}/{short_name}.
	OpNamespacedName = "namespaced_name"
	// OpTagValueId is the bare numeric id.
	OpTagValueId = "tag_value_id"
	// OpCreateTime is the RFC 3339 creation timestamp.
	OpCreateTime = "create_time"
)
