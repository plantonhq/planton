# Outputs — identical names and derivations in the Pulumi module's
# outputs.go / main.go exports.

output "listener_set_name" {
  description = "Name of the created ListenerSet (equals metadata.name)."
  value       = var.metadata.name
}

output "namespace" {
  description = "Namespace of the created ListenerSet."
  value       = var.spec.namespace
}

output "gateway_name" {
  description = "Name of the parent Gateway the listeners attach to (the resolved spec.parentRef.name)."
  value       = var.spec.parentRef.name
}
