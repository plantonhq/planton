output "namespace" {
  description = "Kubernetes namespace where Harbor is deployed"
  value       = local.namespace
}

output "core_service" {
  description = "Harbor Core service name"
  value       = local.core_service_name
}

output "portal_service" {
  description = "Harbor Portal service name"
  value       = local.portal_service_name
}

output "registry_service" {
  description = "Harbor Registry service name"
  value       = local.registry_service_name
}

output "port_forward_command" {
  description = "kubectl port-forward command for local access"
  value       = "kubectl port-forward -n ${local.namespace} service/${local.portal_service_name} 8080:80"
}

