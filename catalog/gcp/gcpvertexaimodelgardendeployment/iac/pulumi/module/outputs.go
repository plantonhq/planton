package module

// Output keys must match the field names in outputs.proto -- the outputs
// transformer maps raw engine outputs onto the proto by name.
const (
	OpEndpointId               = "endpoint_id"
	OpEndpointName             = "endpoint_name"
	OpDeployedModelId          = "deployed_model_id"
	OpDeployedModelDisplayName = "deployed_model_display_name"
	OpLocation                 = "location"
)
