output "destination_rule_name" {
  description = "Name of the created DestinationRule (equals metadata.name)."
  value       = var.metadata.name
}

output "namespace" {
  description = "Namespace the DestinationRule was created in."
  value       = var.spec.namespace
}
