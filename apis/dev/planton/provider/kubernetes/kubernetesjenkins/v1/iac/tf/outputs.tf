output "namespace" {
  description = "The Kubernetes namespace where Jenkins is deployed."
  value       = local.namespace
}

output "helm_release_name" {
  description = "The Helm release name for Jenkins."
  value       = local.resource_id
}

output "ingress_external_hostname" {
  description = "The external hostname for Jenkins, if ingress is enabled."
  value       = local.ingress_external_hostname
}

output "ingress_internal_hostname" {
  description = "The internal hostname for Jenkins, if ingress is enabled."
  value       = local.ingress_internal_hostname
}
