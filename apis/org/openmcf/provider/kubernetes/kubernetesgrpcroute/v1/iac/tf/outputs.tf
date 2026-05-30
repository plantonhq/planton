output "route_name" {
  description = "Name of the created GRPCRoute (equals metadata.name)"
  value       = local.route_name
}

output "namespace" {
  description = "Namespace the GRPCRoute was created in"
  value       = local.namespace
}
