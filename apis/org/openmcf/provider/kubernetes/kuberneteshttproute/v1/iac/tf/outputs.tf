output "route_name" {
  description = "Name of the created HTTPRoute (equals metadata.name)."
  value       = var.metadata.name
}

output "namespace" {
  description = "Namespace of the created HTTPRoute."
  value       = var.spec.namespace
}
