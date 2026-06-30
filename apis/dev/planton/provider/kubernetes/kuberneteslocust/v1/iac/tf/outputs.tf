output "namespace" {
  description = "The Kubernetes namespace."
  value       = local.namespace
}

output "service" {
  description = "The Kubernetes service name."
  value       = local.kube_service_name
}

output "kube_endpoint" {
  description = "The service's internal DNS name."
  value       = local.kube_service_fqdn
}

output "port_forward_command" {
  description = "A handy port-forward command for local debugging."
  value       = local.kube_port_forward_command
}

output "external_hostname" {
  description = "The external ingress hostname (if ingress is enabled)."
  value       = local.ingress_external_hostname
}

output "internal_hostname" {
  description = "The internal ingress hostname (if ingress is enabled)."
  value       = local.ingress_internal_hostname
}
