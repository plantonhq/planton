output "namespace" {
  description = "Kubernetes namespace where Temporal is deployed"
  value       = local.namespace
}

output "frontend_service_name" {
  description = "Service name for the Temporal frontend"
  value       = local.frontend_service_name
}

output "ui_service_name" {
  description = "Service name for the Temporal web UI"
  value       = local.ui_service_name
}

output "port_forward_frontend_command" {
  description = "Command to port-forward the frontend service"
  value       = local.port_forward_frontend_command
}

output "port_forward_ui_command" {
  description = "Command to port-forward the UI service"
  value       = local.port_forward_ui_command
}

output "frontend_endpoint" {
  description = "Internal cluster endpoint for the frontend (e.g. temporal-frontend.namespace.svc.cluster.local:7233)"
  value       = local.frontend_endpoint
}

output "web_ui_endpoint" {
  description = "Internal cluster endpoint for the UI (e.g. temporal-ui.namespace.svc.cluster.local:8080)"
  value       = local.web_ui_endpoint
}

output "external_frontend_hostname" {
  description = "External hostname if load balancer is enabled for the frontend"
  value       = local.external_frontend_hostname
}

output "external_ui_hostname" {
  description = "External hostname for the UI if ingress is configured"
  value       = local.external_ui_hostname
}

