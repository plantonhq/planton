# Outputs — exactly the DigitalOceanKubernetesClusterOutputs
# contract, identical across both provisioners.

output "cluster_id" {
  description = "The unique identifier (UUID) of the Kubernetes cluster"
  value       = digitalocean_kubernetes_cluster.cluster.id
}

output "kubeconfig" {
  description = "The raw kubeconfig YAML for accessing the cluster (contains admin credentials)"
  value       = digitalocean_kubernetes_cluster.cluster.kube_config[0].raw_config
  sensitive   = true
}

output "api_server_endpoint" {
  description = "The endpoint URL of the Kubernetes API server"
  value       = digitalocean_kubernetes_cluster.cluster.endpoint
}

output "urn" {
  description = "The uniform resource name of the cluster (do:kubernetes:<cluster_id>)"
  value       = digitalocean_kubernetes_cluster.cluster.urn
}

output "ipv4_address" {
  description = "The public IPv4 address of the control plane (empty on HA clusters)"
  value       = digitalocean_kubernetes_cluster.cluster.ipv4_address
}

output "default_node_pool_id" {
  description = "The unique identifier (UUID) of the inline default node pool"
  value       = digitalocean_kubernetes_cluster.cluster.node_pool[0].id
}

output "cluster_subnet" {
  description = "The CIDR block from which pod IPs are assigned"
  value       = digitalocean_kubernetes_cluster.cluster.cluster_subnet
}

output "service_subnet" {
  description = "The CIDR block from which service ClusterIPs are assigned"
  value       = digitalocean_kubernetes_cluster.cluster.service_subnet
}
