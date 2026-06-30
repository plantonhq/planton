output "route_name" {
  description = "Name of the created TLSRoute (equals metadata.name)."
  value       = var.metadata.name
}

output "namespace" {
  description = "Namespace of the created TLSRoute."
  value       = var.spec.namespace
}
