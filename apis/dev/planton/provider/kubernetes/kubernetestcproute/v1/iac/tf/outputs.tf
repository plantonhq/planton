output "route_name" {
  description = "Name of the created TCPRoute (equals metadata.name)."
  value       = var.metadata.name
}

output "namespace" {
  description = "Namespace of the created TCPRoute."
  value       = var.spec.namespace
}
