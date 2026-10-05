# Outputs — identical names and derivations in the Pulumi module
# (AzurePlantonRunnerOutputs).

output "container_app_id" {
  description = "The Azure resource ID of the Container App keeping the runner running."
  value       = azurerm_container_app.runner.id
}

output "container_app_name" {
  description = "The Container App's name (metadata.name)."
  value       = azurerm_container_app.runner.name
}

output "token_secret_name" {
  description = "The Container App secret holding the runner token -- the token authorizes joining and is never the runner's identity."
  value       = local.token_secret_name
}

output "runner_name" {
  description = "The name the runner registers itself under with the control plane -- shown by `planton runner list` the moment it joins."
  value       = local.registration_name
}

output "resource_group_name" {
  description = "The resource group the runner was deployed in."
  value       = var.spec.resource_group
}
