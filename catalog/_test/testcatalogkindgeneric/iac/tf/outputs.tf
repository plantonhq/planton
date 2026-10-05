# Outputs -- names and meanings mirror outputs.proto exactly.

output "id" {
  description = "Synthetic resource id (id_prefix + metadata.name)"
  value       = local.resource_id
}

output "name" {
  description = "The resource name"
  value       = var.metadata.name
}

output "url" {
  description = "Deterministic URL derived from inputs"
  value       = local.url
}

output "tags" {
  description = "Echo of spec.steps commands -- proves repeated message fields round-trip"
  value       = [for s in var.spec.steps : s.command]
}
