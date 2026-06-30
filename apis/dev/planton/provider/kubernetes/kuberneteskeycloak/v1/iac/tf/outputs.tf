output "namespace" {
  description = "The Kubernetes namespace where Keycloak is deployed."
  value       = local.namespace
}

output "service" {
  description = "The name of the Kubernetes service for Keycloak."
  value       = local.service_name
}

output "port_forward_command" {
  description = "Command to setup port-forwarding to access Keycloak from local machine."
  value       = local.port_forward_command
}

output "kube_endpoint" {
  description = "Kubernetes internal endpoint to connect to Keycloak from within the cluster."
  value       = local.kube_endpoint
}

output "external_hostname" {
  description = "Public endpoint to access Keycloak from outside the cluster (if ingress is enabled)."
  value       = local.external_hostname
}

output "internal_hostname" {
  description = "Internal endpoint to access Keycloak from within VPC (if ingress is enabled)."
  value       = local.internal_hostname
}

