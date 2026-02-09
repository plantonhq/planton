output "cluster_id" {
  description = "The unique identifier (UUID) of the cluster"
  value       = openstack_containerinfra_cluster_v1.main.id
}

output "name" {
  description = "The name of the cluster"
  value       = openstack_containerinfra_cluster_v1.main.name
}

output "api_address" {
  description = "The Kubernetes API server endpoint"
  value       = openstack_containerinfra_cluster_v1.main.api_address
}

output "coe_version" {
  description = "The version of the container orchestration engine"
  value       = openstack_containerinfra_cluster_v1.main.coe_version
}

output "master_addresses" {
  description = "IP addresses of the master nodes"
  value       = openstack_containerinfra_cluster_v1.main.master_addresses
}

output "node_addresses" {
  description = "IP addresses of the worker nodes"
  value       = openstack_containerinfra_cluster_v1.main.node_addresses
}

output "kubeconfig_raw" {
  description = "Full kubeconfig YAML"
  value       = try(openstack_containerinfra_cluster_v1.main.kubeconfig["raw_config"], "")
  sensitive   = true
}

output "kubeconfig_host" {
  description = "Kubernetes API server URL from kubeconfig"
  value       = try(openstack_containerinfra_cluster_v1.main.kubeconfig["host"], "")
}

output "kubeconfig_cluster_ca_cert" {
  description = "Cluster CA certificate from kubeconfig"
  value       = try(openstack_containerinfra_cluster_v1.main.kubeconfig["cluster_ca_certificate"], "")
  sensitive   = true
}

output "kubeconfig_client_cert" {
  description = "Client certificate from kubeconfig"
  value       = try(openstack_containerinfra_cluster_v1.main.kubeconfig["client_certificate"], "")
  sensitive   = true
}

output "kubeconfig_client_key" {
  description = "Client private key from kubeconfig"
  value       = try(openstack_containerinfra_cluster_v1.main.kubeconfig["client_key"], "")
  sensitive   = true
}

output "region" {
  description = "The OpenStack region"
  value       = openstack_containerinfra_cluster_v1.main.region
}
