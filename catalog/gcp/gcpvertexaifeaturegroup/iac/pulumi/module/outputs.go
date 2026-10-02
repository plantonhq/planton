package module

// Output keys must match the field names in outputs.proto -- the outputs
// transformer maps raw engine outputs onto the proto by name.
const (
	OpName           = "name"
	OpFeatureGroupId = "feature_group_id"
	OpLocation       = "location"
	OpFeatureNames   = "feature_names"
)
