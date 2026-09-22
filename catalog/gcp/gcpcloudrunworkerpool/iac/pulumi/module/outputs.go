package module

// Output keys must match the field names in outputs.proto -- the outputs
// transformer maps raw engine outputs onto the proto by name.
const (
	OpName                  = "name"
	OpWorkerPoolName        = "worker_pool_name"
	OpUid                   = "uid"
	OpLocation              = "location"
	OpProjectId             = "project_id"
	OpLatestCreatedRevision = "latest_created_revision"
	OpLatestReadyRevision   = "latest_ready_revision"
	OpObservedGeneration    = "observed_generation"
	OpEtag                  = "etag"
)
