output "authorization_policy_name" {
  description = "Name of the created AuthorizationPolicy (equals metadata.name)."
  value       = var.metadata.name
}

output "namespace" {
  description = "Namespace of the created AuthorizationPolicy."
  value       = var.spec.namespace
}
