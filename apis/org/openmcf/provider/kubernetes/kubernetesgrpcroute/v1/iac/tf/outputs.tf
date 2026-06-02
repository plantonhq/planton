output "route_name" {
  description = "Name of the created GRPCRoute (equals metadata.name)."
  value       = var.metadata.name
}

output "namespace" {
  description = "Namespace of the created GRPCRoute."
  value       = var.spec.namespace
}
