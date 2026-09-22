package module

// Output keys must match the field names in outputs.proto -- the outputs
// transformer maps raw engine outputs onto the proto by name.
const (
	OpClusterName     = "cluster_name"
	OpEndpointCount   = "endpoint_count"
	OpConnectionCount = "connection_count"
	OpRegion          = "region"
)
