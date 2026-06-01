output "reference_grant_name" {
  description = "Name of the created ReferenceGrant (equals metadata.name)"
  value       = local.reference_grant_name
}

output "namespace" {
  description = "Namespace the ReferenceGrant was created in"
  value       = local.namespace
}
