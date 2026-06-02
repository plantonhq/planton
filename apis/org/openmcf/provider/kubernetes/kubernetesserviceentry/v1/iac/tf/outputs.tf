output "service_entry_name" {
  description = "Name of the created ServiceEntry (equals metadata.name)"
  value       = local.service_entry_name
}

output "namespace" {
  description = "Namespace the ServiceEntry was created in"
  value       = local.namespace
}
