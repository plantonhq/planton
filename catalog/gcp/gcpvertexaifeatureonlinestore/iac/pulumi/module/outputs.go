package module

// Output keys must match the field names in outputs.proto -- the outputs
// transformer maps raw engine outputs onto the proto by name.
const (
	OpName                     = "name"
	OpFeatureOnlineStoreId     = "feature_online_store_id"
	OpLocation                 = "location"
	OpPublicEndpointDomainName = "public_endpoint_domain_name"
	OpServiceAttachment        = "service_attachment"
	OpFeatureViewNames         = "feature_view_names"
)
