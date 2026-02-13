# Pool ID
output "pool_id" {
  description = "The unique identifier of the created node pool"
  value       = scaleway_k8s_pool.pool.id
}

# Pool Version
output "pool_version" {
  description = "The actual Kubernetes version running on pool nodes"
  value       = scaleway_k8s_pool.pool.version
}

# Current Size
output "current_size" {
  description = "The actual number of nodes currently in the pool"
  value       = scaleway_k8s_pool.pool.current_size
}

# Pool Status
output "pool_status" {
  description = "The current status of the node pool"
  value       = scaleway_k8s_pool.pool.status
}
