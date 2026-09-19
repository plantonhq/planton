package module

// Output keys must match the field names in outputs.proto -- the outputs
// transformer maps raw engine outputs onto the proto by name.
const (
	// OpName is the key's resource name, tagKeys/{numeric_id}.
	OpName = "name"
	// OpNamespacedName is {org_or_project}/{short_name}.
	OpNamespacedName = "namespaced_name"
	// OpTagKeyId is the bare numeric id.
	OpTagKeyId = "tag_key_id"
	// OpCreateTime is the RFC 3339 creation timestamp.
	OpCreateTime = "create_time"
)
