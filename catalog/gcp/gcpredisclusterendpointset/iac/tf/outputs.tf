# Bare cluster name the connections were registered on -- the segment
# Google's resource is keyed by.
output "cluster_name" {
  description = "Bare name of the cluster the connections were registered on"
  value       = google_redis_cluster_user_created_connections.this.name
}

output "endpoint_count" {
  description = "Number of consumer-network endpoints registered"
  value       = length(var.spec.endpoints)
}

output "connection_count" {
  description = "Number of PSC connections registered across every endpoint"
  value       = local.connection_count
}

output "region" {
  description = "The cluster's region"
  value       = google_redis_cluster_user_created_connections.this.region
}
