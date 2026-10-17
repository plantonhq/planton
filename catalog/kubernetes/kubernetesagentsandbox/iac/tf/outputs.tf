output "sandbox_name" {
  description = "Name of the created Sandbox (equals metadata.name)."
  value       = var.metadata.name
}

output "namespace" {
  description = "Namespace of the created Sandbox."
  value       = var.spec.namespace
}
