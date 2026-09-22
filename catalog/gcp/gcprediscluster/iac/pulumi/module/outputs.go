package module

// Output keys must match the field names in outputs.proto -- the outputs
// transformer maps raw engine outputs onto the proto by name.
const (
	OpName                       = "name"
	OpUid                        = "uid"
	OpState                      = "state"
	OpDiscoveryEndpointAddress   = "discovery_endpoint_address"
	OpDiscoveryEndpointPort      = "discovery_endpoint_port"
	OpDiscoveryServiceAttachment = "discovery_service_attachment"
	OpPrimaryServiceAttachment   = "primary_service_attachment"
	OpReaderServiceAttachment    = "reader_service_attachment"
	OpSizeGb                     = "size_gb"
	OpShardCount                 = "shard_count"
	OpReplicaCount               = "replica_count"
	OpBackupCollection           = "backup_collection"
)
