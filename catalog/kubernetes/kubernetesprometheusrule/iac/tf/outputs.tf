output "prometheus_rule_name" {
  description = "Name of the created PrometheusRule (equals metadata.name)."
  value       = var.metadata.name
}

output "namespace" {
  description = "Namespace of the created PrometheusRule."
  value       = var.spec.namespace
}
