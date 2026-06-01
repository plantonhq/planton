output "route_name" {
  description = "Name of the created TLSRoute (equals metadata.name)"
  value       = local.route_name
}

output "namespace" {
  description = "Namespace the TLSRoute was created in"
  value       = local.namespace
}
