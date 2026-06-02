output "reference_grant_name" {
  description = "Name of the created ReferenceGrant (equals metadata.name)."
  value       = var.metadata.name
}

output "namespace" {
  description = "Namespace of the created ReferenceGrant."
  value       = var.spec.namespace
}
