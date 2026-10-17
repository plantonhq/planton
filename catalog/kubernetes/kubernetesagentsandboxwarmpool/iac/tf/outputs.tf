output "sandbox_warm_pool_name" {
  description = "Name of the created SandboxWarmPool (equals metadata.name)."
  value       = var.metadata.name
}

output "namespace" {
  description = "Namespace of the created SandboxWarmPool."
  value       = var.spec.namespace
}
