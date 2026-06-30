# Cluster ID
output "cluster_id" {
  description = "The unique identifier (UUID) of the created Kubernetes cluster"
  value       = digitalocean_kubernetes_cluster.cluster.id
}

# Kubeconfig
output "kubeconfig" {
  description = "Base64-encoded Kubernetes config for accessing the cluster"
  value       = digitalocean_kubernetes_cluster.cluster.kube_config[0].raw_config
  sensitive   = true
}

# API Server Endpoint
output "api_server_endpoint" {
  description = "The endpoint URL of the Kubernetes API server"
  value       = digitalocean_kubernetes_cluster.cluster.endpoint
}

# Cluster URN
output "cluster_urn" {
  description = "The uniform resource name (URN) of the cluster"
  value       = digitalocean_kubernetes_cluster.cluster.urn
}

# Cluster Status
output "cluster_status" {
  description = "The current status of the cluster"
  value       = digitalocean_kubernetes_cluster.cluster.status
}

# Cluster Created At
output "created_at" {
  description = "The timestamp when the cluster was created"
  value       = digitalocean_kubernetes_cluster.cluster.created_at
}

# Cluster Updated At
output "updated_at" {
  description = "The timestamp when the cluster was last updated"
  value       = digitalocean_kubernetes_cluster.cluster.updated_at
}

# Node Pool ID
output "node_pool_id" {
  description = "The ID of the default node pool"
  value       = digitalocean_kubernetes_cluster.cluster.node_pool[0].id
}

# Cluster IPv4 Address
output "ipv4_address" {
  description = "The public IPv4 address of the cluster's control plane"
  value       = digitalocean_kubernetes_cluster.cluster.ipv4_address
}

