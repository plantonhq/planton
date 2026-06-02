output "service_entry_name" {
  description = "Name of the created ServiceEntry (equals metadata.name)."
  value       = var.metadata.name
}

output "namespace" {
  description = "Namespace of the created ServiceEntry."
  value       = var.spec.namespace
}
