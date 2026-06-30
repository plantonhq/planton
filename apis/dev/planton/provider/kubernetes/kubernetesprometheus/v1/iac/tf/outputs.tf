output "namespace" {
  description = "The namespace in which the Prometheus resources are deployed."
  value       = local.namespace
}

output "service" {
  description = "Name of the Prometheus service."
  value       = local.kube_service_name
}

output "port_forward_command" {
  description = "Convenient command to port-forward to the Prometheus service."
  value       = local.kube_port_forward_command
}

output "kube_endpoint" {
  description = "FQDN of the Prometheus service within the cluster."
  value       = local.kube_service_fqdn
}

output "external_hostname" {
  description = "The external hostname for Prometheus if ingress is enabled."
  value       = local.external_hostname
}

output "internal_hostname" {
  description = "The internal hostname for Prometheus if ingress is enabled."
  value       = local.internal_hostname
}

