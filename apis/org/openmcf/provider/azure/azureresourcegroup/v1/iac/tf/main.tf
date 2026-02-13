# Create the Azure Resource Group
resource "azurerm_resource_group" "main" {
  name     = var.spec.name
  location = var.spec.region
  tags     = local.final_tags
}
