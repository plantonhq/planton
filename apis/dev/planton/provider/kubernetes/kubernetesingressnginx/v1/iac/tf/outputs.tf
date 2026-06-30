output "namespace" {
  description = "The namespace where the Ingress Nginx is deployed."
  value       = local.namespace
}

output "release_name" {
  description = "The name of the Helm release."
  value       = local.release_name
}

output "service_name" {
  description = "The name of the service created by the Ingress Nginx controller."
  value       = local.service_name
}

output "service_type" {
  description = "The type of service (LoadBalancer)."
  value       = local.service_type
}

