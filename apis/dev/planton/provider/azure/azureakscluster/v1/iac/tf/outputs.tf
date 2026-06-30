# Outputs aligned to AzureAksClusterStackOutputs proto

output "api_server_endpoint" {
  description = "The URL of the Kubernetes API server endpoint for the AKS cluster"
  value       = azurerm_kubernetes_cluster.aks.fqdn
}

output "cluster_resource_id" {
  description = "The Azure Resource ID of the AKS cluster"
  value       = azurerm_kubernetes_cluster.aks.id
}

output "cluster_kubeconfig" {
  description = "Kubeconfig file contents for the cluster (base64-encoded)"
  value       = azurerm_kubernetes_cluster.aks.kube_config_raw
  sensitive   = true
}

output "managed_identity_principal_id" {
  description = "The Azure AD principal ID of the cluster's managed identity"
  value       = azurerm_kubernetes_cluster.aks.identity[0].principal_id
}

# Additional helpful outputs

output "resource_group_name" {
  description = "Name of the resource group containing the AKS cluster"
  value       = local.resource_group_name
}

output "cluster_name" {
  description = "Name of the AKS cluster"
  value       = azurerm_kubernetes_cluster.aks.name
}

output "node_resource_group" {
  description = "The auto-generated Resource Group which contains the resources for this Managed Kubernetes Cluster"
  value       = azurerm_kubernetes_cluster.aks.node_resource_group
}

