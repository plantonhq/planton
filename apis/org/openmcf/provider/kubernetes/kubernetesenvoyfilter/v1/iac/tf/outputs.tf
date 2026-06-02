output "envoy_filter_name" {
  description = "Name of the created EnvoyFilter (equals metadata.name)."
  value       = var.metadata.name
}

output "namespace" {
  description = "Namespace of the created EnvoyFilter."
  value       = var.spec.namespace
}
