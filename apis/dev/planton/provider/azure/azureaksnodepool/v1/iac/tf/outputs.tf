# Outputs aligned to AzureAksNodePoolStackOutputs proto

output "node_pool_name" {
  description = "Name of the node pool in AKS"
  value       = azurerm_kubernetes_cluster_node_pool.node_pool.name
}

output "agent_pool_resource_id" {
  description = "Azure Resource Manager ID of the created node pool"
  value       = azurerm_kubernetes_cluster_node_pool.node_pool.id
}

output "max_pods_per_node" {
  description = "Maximum number of pods that can run on each node"
  value       = azurerm_kubernetes_cluster_node_pool.node_pool.max_pods
}

