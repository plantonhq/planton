package module

// Output keys must match the field names in outputs.proto -- the outputs
// transformer maps raw engine outputs onto the proto by name.
const (
	OpName                  = "name"
	OpTensorboardId         = "tensorboard_id"
	OpLocation              = "location"
	OpBlobStoragePathPrefix = "blob_storage_path_prefix"
	OpExperimentNames       = "experiment_names"
	OpRunNames              = "run_names"
)
