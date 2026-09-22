package module

// Output keys must match the field names in outputs.proto -- the outputs
// transformer maps raw engine outputs onto the proto by name.
const (
	OpName              = "name"
	OpReasoningEngineId = "reasoning_engine_id"
	OpLocation          = "location"
	OpCreateTime        = "create_time"
	OpUpdateTime        = "update_time"
)
