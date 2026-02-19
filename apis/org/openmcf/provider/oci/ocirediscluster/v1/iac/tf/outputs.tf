output "cluster_id" {
  description = "OCID of the Redis cluster"
  value       = oci_redis_redis_cluster.this.id
}

output "primary_fqdn" {
  description = "FQDN of the primary (read-write) endpoint"
  value       = oci_redis_redis_cluster.this.primary_fqdn
}

output "primary_endpoint_ip_address" {
  description = "Private IP address of the primary endpoint"
  value       = oci_redis_redis_cluster.this.primary_endpoint_ip_address
}

output "replicas_fqdn" {
  description = "FQDN of the replica (read-only) endpoint"
  value       = oci_redis_redis_cluster.this.replicas_fqdn
}

output "discovery_fqdn" {
  description = "FQDN of the discovery endpoint for sharded clusters"
  value       = oci_redis_redis_cluster.this.discovery_fqdn
}
