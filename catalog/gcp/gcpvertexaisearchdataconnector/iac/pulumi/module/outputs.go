package module

// Output keys must match the field names in outputs.proto -- the outputs
// transformer maps raw engine outputs onto the proto by name.
const (
	OpName                         = "name"
	OpCollectionId                 = "collection_id"
	OpLocation                     = "location"
	OpState                        = "state"
	OpEntityDataStores             = "entity_data_stores"
	OpStaticIpAddresses            = "static_ip_addresses"
	OpPrivateConnectivityProjectId = "private_connectivity_project_id"
)
