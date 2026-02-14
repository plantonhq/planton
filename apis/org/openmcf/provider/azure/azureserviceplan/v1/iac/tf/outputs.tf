output "plan_id" {
  description = "The Azure Resource Manager ID of the Service Plan"
  value       = azurerm_service_plan.main.id
}

output "plan_name" {
  description = "The name of the Service Plan"
  value       = azurerm_service_plan.main.name
}

output "os_type" {
  description = "The configured operating system type"
  value       = azurerm_service_plan.main.os_type
}

output "sku_name" {
  description = "The configured SKU name"
  value       = azurerm_service_plan.main.sku_name
}
