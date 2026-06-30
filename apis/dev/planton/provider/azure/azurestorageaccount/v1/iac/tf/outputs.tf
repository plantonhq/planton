output "storage_account_id" {
  description = "The Azure Resource Manager ID of the Storage Account"
  value       = azurerm_storage_account.main.id
}

output "storage_account_name" {
  description = "The name of the Storage Account"
  value       = azurerm_storage_account.main.name
}

output "primary_blob_endpoint" {
  description = "The primary blob service endpoint URL"
  value       = azurerm_storage_account.main.primary_blob_endpoint
}

output "primary_queue_endpoint" {
  description = "The primary queue service endpoint URL"
  value       = azurerm_storage_account.main.primary_queue_endpoint
}

output "primary_table_endpoint" {
  description = "The primary table service endpoint URL"
  value       = azurerm_storage_account.main.primary_table_endpoint
}

output "primary_file_endpoint" {
  description = "The primary file service endpoint URL"
  value       = azurerm_storage_account.main.primary_file_endpoint
}

output "primary_dfs_endpoint" {
  description = "The primary DFS (Data Lake Storage Gen2) endpoint URL"
  value       = azurerm_storage_account.main.primary_dfs_endpoint
}

output "primary_web_endpoint" {
  description = "The primary web (static website) endpoint URL"
  value       = azurerm_storage_account.main.primary_web_endpoint
}

output "container_url_map" {
  description = "Map of container names to their URLs"
  value = {
    for name, container in azurerm_storage_container.containers :
    name => "${azurerm_storage_account.main.primary_blob_endpoint}${container.name}"
  }
}

output "region" {
  description = "The Azure region where the Storage Account was deployed"
  value       = var.spec.region
}

output "resource_group" {
  description = "The resource group name where the Storage Account was created"
  value       = var.spec.resource_group
}
