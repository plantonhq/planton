output "sandbox_claim_name" {
  description = "Name of the created SandboxClaim (equals metadata.name)."
  value       = var.metadata.name
}

output "namespace" {
  description = "Namespace of the created SandboxClaim."
  value       = var.spec.namespace
}
