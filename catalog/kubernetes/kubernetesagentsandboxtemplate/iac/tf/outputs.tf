output "sandbox_template_name" {
  description = "Name of the created SandboxTemplate (equals metadata.name)."
  value       = var.metadata.name
}

output "namespace" {
  description = "Namespace of the created SandboxTemplate."
  value       = var.spec.namespace
}
