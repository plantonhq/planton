output "destination_rule_name" {
  description = "Name of the created DestinationRule (equals metadata.name)"
  value       = local.destination_rule_name
}

output "namespace" {
  description = "Namespace the DestinationRule was created in"
  value       = local.namespace
}
