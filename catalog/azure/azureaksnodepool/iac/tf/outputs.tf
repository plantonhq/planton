# Semantic outputs, matching AzureAksNodePoolOutputs field for
# field. Nothing downstream deploys INTO a pool (workloads target pools
# via Kubernetes labels/taints), so the outputs are the pool's own
# identifiers plus the node image actually rolled out.

output "node_pool_id" {
  description = "The Azure Resource Manager ID of the agent pool."
  value       = azurerm_kubernetes_cluster_node_pool.main.id
}

output "node_pool_name" {
  description = "The name of the node pool."
  value       = azurerm_kubernetes_cluster_node_pool.main.name
}

output "node_image_version" {
  description = "The node image version the pool is running."
  value       = azurerm_kubernetes_cluster_node_pool.main.node_image_version
}
