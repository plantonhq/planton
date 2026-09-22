package module

// Output keys must match the field names in outputs.proto -- the outputs
// transformer maps raw engine outputs onto the proto by name.
const (
	OpName         = "name"
	OpCollectionId = "collection_id"
	OpLocation     = "location"
	OpIndexNames   = "index_names"
	OpIndexCount   = "index_count"
)
