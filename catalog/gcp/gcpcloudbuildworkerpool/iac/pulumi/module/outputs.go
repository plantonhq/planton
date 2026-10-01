package module

// Output keys must match the field names in outputs.proto -- the outputs
// transformer maps raw engine outputs onto the proto by name.
const (
	OpName         = "name"
	OpWorkerPoolId = "worker_pool_id"
	OpState        = "state"
	OpUid          = "uid"
)
