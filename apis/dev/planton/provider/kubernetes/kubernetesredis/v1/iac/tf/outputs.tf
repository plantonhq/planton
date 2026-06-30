output "namespace" {
  description = "The namespace where Redis and related resources are created."
  value       = local.namespace
}

output "service" {
  description = "The Kubernetes service name for Redis."
  value       = local.kube_service_name
}

output "port_forward_command" {
  description = "A handy command to port-forward local 8080 to the Redis service 8080."
  value       = local.kube_port_forward_command
}

output "kube_endpoint" {
  description = "The internal service FQDN for Redis."
  value       = local.kube_service_fqdn
}

output "external_hostname" {
  description = "The external LB hostname, if ingress is enabled."
  value       = local.ingress_external_hostname
}
