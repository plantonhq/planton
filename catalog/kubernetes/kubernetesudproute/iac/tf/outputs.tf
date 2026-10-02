output "route_name" {
  description = "Name of the created UDPRoute (equals metadata.name)."
  value       = var.metadata.name
}

output "namespace" {
  description = "Namespace of the created UDPRoute."
  value       = var.spec.namespace
}
