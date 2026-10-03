output "pod_monitor_name" {
  description = "Name of the created PodMonitor (equals metadata.name)."
  value       = var.metadata.name
}

output "namespace" {
  description = "Namespace of the created PodMonitor."
  value       = var.spec.namespace
}
