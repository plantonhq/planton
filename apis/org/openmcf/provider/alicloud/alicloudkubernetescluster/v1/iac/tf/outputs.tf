output "cluster_id" {
  description = "ACK cluster ID"
  value       = alicloud_cs_managed_kubernetes.cluster.id
}

output "cluster_name" {
  description = "Cluster name"
  value       = alicloud_cs_managed_kubernetes.cluster.name
}

output "api_server_internet" {
  description = "Public API server endpoint"
  value       = try(alicloud_cs_managed_kubernetes.cluster.connections["api_server_internet"], "")
}

output "api_server_intranet" {
  description = "Private API server endpoint"
  value       = try(alicloud_cs_managed_kubernetes.cluster.connections["api_server_intranet"], "")
}

output "vpc_id" {
  description = "VPC ID where the cluster is deployed"
  value       = alicloud_cs_managed_kubernetes.cluster.vpc_id
}

output "security_group_id" {
  description = "Security group ID used by cluster nodes"
  value       = alicloud_cs_managed_kubernetes.cluster.security_group_id
}

output "nat_gateway_id" {
  description = "NAT gateway ID auto-created by the cluster"
  value       = alicloud_cs_managed_kubernetes.cluster.nat_gateway_id
}

output "worker_ram_role_name" {
  description = "RAM role name attached to worker nodes"
  value       = alicloud_cs_managed_kubernetes.cluster.worker_ram_role_name
}

output "rrsa_oidc_issuer_url" {
  description = "RRSA OIDC issuer URL"
  value       = try(alicloud_cs_managed_kubernetes.cluster.rrsa_metadata[0].rrsa_oidc_issuer_url, "")
}

output "ram_oidc_provider_name" {
  description = "RRSA OIDC provider name"
  value       = try(alicloud_cs_managed_kubernetes.cluster.rrsa_metadata[0].ram_oidc_provider_name, "")
}

output "ram_oidc_provider_arn" {
  description = "RRSA OIDC provider ARN"
  value       = try(alicloud_cs_managed_kubernetes.cluster.rrsa_metadata[0].ram_oidc_provider_arn, "")
}
