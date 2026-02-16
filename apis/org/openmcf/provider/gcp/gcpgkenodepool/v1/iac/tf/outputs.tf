###############################################################################
# Outputs
###############################################################################

output "node_pool_name" {
  description = "Name of the GKE node pool"
  value       = google_container_node_pool.node_pool.name
}

output "instance_group_urls" {
  description = "URLs of the Compute Instance Group(s) backing this node pool (one per zone for regional clusters)"
  value       = google_container_node_pool.node_pool.instance_group_urls
}

output "min_nodes" {
  description = "Effective minimum size of the node pool"
  value = (
    var.spec.autoscaling != null
    ? var.spec.autoscaling.min_nodes
    : var.spec.node_count
  )
}

output "max_nodes" {
  description = "Effective maximum size of the node pool"
  value = (
    var.spec.autoscaling != null
    ? var.spec.autoscaling.max_nodes
    : var.spec.node_count
  )
}

output "current_node_count" {
  description = "Current number of nodes in this pool (managed by autoscaler if enabled)"
  value       = google_container_node_pool.node_pool.node_count
}

output "node_pool_id" {
  description = "Fully qualified GKE node pool resource ID (projects/{project}/locations/{location}/clusters/{cluster}/nodePools/{nodePool})"
  value       = google_container_node_pool.node_pool.id
}

