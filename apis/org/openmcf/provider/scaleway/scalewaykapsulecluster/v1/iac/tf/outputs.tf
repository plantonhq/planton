# Cluster ID
output "cluster_id" {
  description = "The unique identifier of the created Kapsule cluster"
  value       = scaleway_k8s_cluster.cluster.id
}

# Kubeconfig (sensitive)
output "kubeconfig" {
  description = "Raw kubeconfig file content for connecting to the cluster"
  value       = scaleway_k8s_cluster.cluster.kubeconfig[0].config_file
  sensitive   = true
}

# API Server URL
output "apiserver_url" {
  description = "The URL of the Kubernetes API server"
  value       = scaleway_k8s_cluster.cluster.apiserver_url
}

# Cluster CA Certificate
output "cluster_ca_certificate" {
  description = "The CA certificate of the Kubernetes API server (base64-encoded)"
  value       = scaleway_k8s_cluster.cluster.kubeconfig[0].cluster_ca_certificate
  sensitive   = true
}

# Wildcard DNS
output "wildcard_dns" {
  description = "DNS wildcard for ready nodes in the cluster"
  value       = scaleway_k8s_cluster.cluster.wildcard_dns
}

# Default Pool ID
output "default_pool_id" {
  description = "The unique identifier of the default node pool"
  value       = scaleway_k8s_pool.default.id
}
