output "request_authentication_name" {
  description = "Name of the created RequestAuthentication (equals metadata.name)."
  value       = var.metadata.name
}

output "namespace" {
  description = "Namespace of the created RequestAuthentication."
  value       = var.spec.namespace
}
