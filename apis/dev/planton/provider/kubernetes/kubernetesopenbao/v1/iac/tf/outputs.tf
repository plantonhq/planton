output "namespace" {
  description = "Kubernetes namespace where OpenBao is deployed"
  value       = local.namespace
}

output "service" {
  description = "Kubernetes service name for OpenBao"
  value       = local.kube_service_name
}

output "port_forward_command" {
  description = "Command to set up port-forwarding for local access"
  value       = local.kube_port_forward_command
}

output "kube_endpoint" {
  description = "Internal Kubernetes endpoint for OpenBao"
  value       = local.kube_service_fqdn
}

output "external_hostname" {
  description = "External hostname for OpenBao (when ingress is enabled)"
  value       = local.ingress_is_enabled ? local.ingress_external_hostname : null
}

output "api_address" {
  description = "Full API address for OpenBao"
  value       = local.api_address
}

output "cluster_address" {
  description = "Cluster communication address (HA mode)"
  value       = local.cluster_address
}

output "ha_enabled" {
  description = "Boolean indicating if HA mode is enabled"
  value       = local.ha_enabled
}
