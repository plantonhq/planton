output "policy_name" {
  description = "Name of the created BackendTLSPolicy (equals metadata.name)."
  value       = var.metadata.name
}

output "namespace" {
  description = "Namespace of the created BackendTLSPolicy."
  value       = var.spec.namespace
}
