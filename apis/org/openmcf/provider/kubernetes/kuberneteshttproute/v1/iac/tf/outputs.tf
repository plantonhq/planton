output "route_name" {
  description = "Name of the created HTTPRoute (equals metadata.name)"
  value       = local.route_name
}

output "namespace" {
  description = "Namespace the HTTPRoute was created in"
  value       = local.namespace
}
