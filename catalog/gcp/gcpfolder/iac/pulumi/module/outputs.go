package module

// Output keys must match the field names in outputs.proto -- the outputs
// transformer maps raw engine outputs onto the proto by name.
const (
	// OpFolderId is the folder's numeric id -- what every child references.
	OpFolderId = "folder_id"
	// OpName is the folder's resource name, folders/{folder_id}.
	OpName = "name"
	// OpLifecycleState is ACTIVE or DELETE_REQUESTED.
	OpLifecycleState = "lifecycle_state"
	// OpCreateTime is the RFC 3339 creation timestamp.
	OpCreateTime = "create_time"
)
