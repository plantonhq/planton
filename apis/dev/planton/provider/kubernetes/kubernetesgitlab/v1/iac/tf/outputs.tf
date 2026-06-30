###########################
# outputs.tf
###########################

output "namespace" {
  description = "Namespace where GitLab is deployed"
  value       = local.namespace_name
}

output "service_name" {
  description = "Kubernetes service name for GitLab"
  value       = kubernetes_service.gitlab.metadata[0].name
}

output "service_fqdn" {
  description = "Fully qualified domain name of the GitLab service"
  value       = local.gitlab_service_fqdn
}

output "port_forward_command" {
  description = "Command to port-forward to GitLab service"
  value       = local.port_forward_command
}

output "ingress_hostname" {
  description = "External hostname for GitLab (if ingress is enabled)"
  value       = local.ingress_is_enabled ? local.ingress_external_hostname : null
}

output "internal_endpoint" {
  description = "Internal Kubernetes endpoint for GitLab"
  value       = "http://${local.gitlab_service_fqdn}:${local.gitlab_port}"
}

output "external_endpoint" {
  description = "External HTTPS endpoint for GitLab (if ingress is enabled)"
  value       = local.ingress_is_enabled ? "https://${local.ingress_external_hostname}" : null
}

