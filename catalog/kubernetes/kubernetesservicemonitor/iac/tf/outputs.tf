output "service_monitor_name" {
  description = "Name of the created ServiceMonitor (equals metadata.name)."
  value       = var.metadata.name
}

output "namespace" {
  description = "Namespace of the created ServiceMonitor."
  value       = var.spec.namespace
}
