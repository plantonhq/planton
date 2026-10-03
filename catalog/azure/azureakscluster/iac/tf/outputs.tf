# Semantic outputs, matching AzureAksClusterOutputs field for
# field. cluster_id is the parent seam every standalone AzureAksNodePool
# consumes; oidc_issuer_url is the trust anchor an
# AzureFederatedIdentityCredential binds to for workload identity.

output "cluster_id" {
  description = "The Azure Resource Manager ID of the managed cluster."
  value       = azurerm_kubernetes_cluster.main.id
}

output "cluster_name" {
  description = "The name of the managed cluster."
  value       = azurerm_kubernetes_cluster.main.name
}

output "fqdn" {
  description = "The public FQDN of the Kubernetes API server (empty for private clusters without a public FQDN)."
  value       = azurerm_kubernetes_cluster.main.fqdn
}

output "private_fqdn" {
  description = "The private FQDN of the API server, populated for private clusters."
  value       = azurerm_kubernetes_cluster.main.private_fqdn
}

output "portal_fqdn" {
  description = "The FQDN used by the Azure Portal to reach the cluster."
  value       = azurerm_kubernetes_cluster.main.portal_fqdn
}

output "oidc_issuer_url" {
  description = "The cluster's OIDC issuer URL -- consumed by AzureFederatedIdentityCredential as its issuer."
  value       = azurerm_kubernetes_cluster.main.oidc_issuer_url
}

output "node_resource_group" {
  description = "The name of the Azure-managed node resource group."
  value       = azurerm_kubernetes_cluster.main.node_resource_group
}

output "node_resource_group_id" {
  description = "The ARM ID of the node resource group."
  value       = azurerm_kubernetes_cluster.main.node_resource_group_id
}

output "cluster_kubeconfig" {
  description = "Base64-encoded kubeconfig for the cluster (user credential). Treat as a secret."
  value       = base64encode(azurerm_kubernetes_cluster.main.kube_config_raw)
  sensitive   = true
}

output "cluster_identity_principal_id" {
  description = "The principal (object) ID of the cluster's managed identity."
  value       = try(azurerm_kubernetes_cluster.main.identity[0].principal_id, "")
}

output "kubelet_identity_object_id" {
  description = "The object ID of the kubelet identity -- grant it AcrPull on registries."
  value       = try(azurerm_kubernetes_cluster.main.kubelet_identity[0].object_id, "")
}

output "kubelet_identity_client_id" {
  description = "The client ID of the kubelet identity."
  value       = try(azurerm_kubernetes_cluster.main.kubelet_identity[0].client_id, "")
}

output "current_kubernetes_version" {
  description = "The Kubernetes version the control plane is actually running."
  value       = azurerm_kubernetes_cluster.main.current_kubernetes_version
}

# The CA certificate is public cluster identity (the TLS trust anchor), not
# credential material. nonsensitive() deliberately unwraps the sensitivity it
# inherits from the enclosing kube_config attribute so the platform's
# cluster-connection materializer can read it as a plain output -- the
# same posture as the EKS/GKE CA outputs.
output "cluster_ca_certificate" {
  description = "Base64-encoded cluster CA certificate (the standard kubeconfig certificate-authority-data format)."
  value       = try(nonsensitive(azurerm_kubernetes_cluster.main.kube_config[0].cluster_ca_certificate), "")
}

# The applicability gate for token-based cluster connections: only an
# Entra-integrated API server honors short-lived Entra bearer tokens, so the
# materializer publishes a connection only when this reads "true".
output "entra_integration_enabled" {
  description = "\"true\" when the cluster is Entra ID (Azure AD) integrated, else \"false\"."
  value       = var.spec.azure_active_directory_role_based_access_control != null ? "true" : "false"
}
