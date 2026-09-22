# Full resource path -- the composition key: a SECONDARY's primary_cluster
# reference and a GcpRedisClusterEndpointSet's cluster reference resolve to
# it.
output "name" {
  description = "Full resource path of the cluster"
  value       = google_redis_cluster.this.name
}

output "uid" {
  description = "Server-generated unique identifier for the cluster"
  value       = google_redis_cluster.this.uid
}

output "state" {
  description = "Lifecycle state of the cluster after provisioning"
  value       = google_redis_cluster.this.state
}

# The Google-placed discovery endpoint: present only when psc_configs asked
# for automatic connectivity; empty otherwise.
output "discovery_endpoint_address" {
  description = "IP address of the discovery endpoint (empty without psc_configs)"
  value       = try(google_redis_cluster.this.discovery_endpoints[0].address, "")
}

output "discovery_endpoint_port" {
  description = "Port of the discovery endpoint (0 without psc_configs)"
  value       = try(google_redis_cluster.this.discovery_endpoints[0].port, 0)
}

# One scalar handle per connection type (see locals.service_attachments_by_type)
# so a consumer's forwarding rule can reference the attachment it needs.
output "discovery_service_attachment" {
  description = "Service attachment a consumer forwarding rule targets for the discovery endpoint"
  value       = try(local.service_attachments_by_type["CONNECTION_TYPE_DISCOVERY"][0], "")
}

output "primary_service_attachment" {
  description = "Service attachment for the primary endpoint (empty when Google publishes none)"
  value       = try(local.service_attachments_by_type["CONNECTION_TYPE_PRIMARY"][0], "")
}

output "reader_service_attachment" {
  description = "Service attachment for the reader endpoint (empty without replicas)"
  value       = try(local.service_attachments_by_type["CONNECTION_TYPE_READER"][0], "")
}

output "size_gb" {
  description = "Redis memory across the whole cluster in GB"
  value       = google_redis_cluster.this.size_gb
}

output "shard_count" {
  description = "Shard count in effect after provisioning"
  value       = google_redis_cluster.this.shard_count
}

output "replica_count" {
  description = "Replica count per shard in effect after provisioning"
  value       = google_redis_cluster.this.replica_count
}

# Where managed backups for this cluster live once automated backups are
# configured -- the source of managed_backup_source paths for seeding new
# clusters.
output "backup_collection" {
  description = "Full resource path of the cluster's managed backup collection"
  value       = google_redis_cluster.this.backup_collection
}
