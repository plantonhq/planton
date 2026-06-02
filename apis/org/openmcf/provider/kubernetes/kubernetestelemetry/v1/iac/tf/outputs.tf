output "telemetry_name" {
  description = "Name of the created Telemetry (equals metadata.name)."
  value       = var.metadata.name
}

output "namespace" {
  description = "Namespace of the created Telemetry."
  value       = var.spec.namespace
}
