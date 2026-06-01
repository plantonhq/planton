output "route_name" {
  description = "Name of the created TCPRoute (equals metadata.name)"
  value       = local.route_name
}

output "namespace" {
  description = "Namespace the TCPRoute was created in"
  value       = local.namespace
}
