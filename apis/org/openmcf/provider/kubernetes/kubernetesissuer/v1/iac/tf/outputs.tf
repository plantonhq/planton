output "namespace" {
  description = "Namespace where the Issuer was created"
  value       = local.namespace
}

output "issuer_name" {
  description = "Name of the created Issuer (equals metadata.name)"
  value       = local.issuer_name
}
