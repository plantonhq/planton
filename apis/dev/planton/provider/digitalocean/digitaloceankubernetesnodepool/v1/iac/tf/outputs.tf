# Node Pool ID
output "node_pool_id" {
  description = "The unique identifier (UUID) of the created node pool"
  value       = digitalocean_kubernetes_node_pool.node_pool.id
}

# Node IDs
output "node_ids" {
  description = "List of node IDs in this pool"
  value       = digitalocean_kubernetes_node_pool.node_pool.nodes[*].id
}

# Node Names
output "node_names" {
  description = "List of node names in this pool"
  value       = digitalocean_kubernetes_node_pool.node_pool.nodes[*].name
}

# Actual Node Count
output "actual_node_count" {
  description = "The actual number of nodes in the pool"
  value       = digitalocean_kubernetes_node_pool.node_pool.actual_node_count
}

# Pool Name
output "pool_name" {
  description = "The name of the node pool"
  value       = digitalocean_kubernetes_node_pool.node_pool.name
}

# Pool Size
output "pool_size" {
  description = "The Droplet size used for nodes in this pool"
  value       = digitalocean_kubernetes_node_pool.node_pool.size
}

