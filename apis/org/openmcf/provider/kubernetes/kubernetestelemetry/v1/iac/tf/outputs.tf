output "telemetry_name" {
  description = "Name of the created Telemetry resource (equals metadata.name)"
  value       = local.telemetry_name
}

output "namespace" {
  description = "Namespace the Telemetry resource was created in"
  value       = local.namespace
}
