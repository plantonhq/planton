output "envoy_filter_name" {
  description = "Name of the created EnvoyFilter (equals metadata.name)"
  value       = local.envoy_filter_name
}

output "namespace" {
  description = "Namespace the EnvoyFilter was created in"
  value       = local.namespace
}
