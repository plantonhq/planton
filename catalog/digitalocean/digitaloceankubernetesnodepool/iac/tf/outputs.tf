output "node_pool_id" {
  description = "The unique identifier (UUID) of the created node pool"
  value       = digitalocean_kubernetes_node_pool.node_pool.id
}

output "cluster_id" {
  description = "The UUID of the cluster that owns this pool"
  value       = digitalocean_kubernetes_node_pool.node_pool.cluster_id
}

# The pool's nodes (nodes[*].id, nodes[*].droplet_id) are deliberately not
# exported: DOKS replaces them by design (autoscaling, upgrades, auto-repair),
# so an apply-time list is stale the next time the pool changes shape.
# Droplet-scoped resources target the pool's tags instead.
