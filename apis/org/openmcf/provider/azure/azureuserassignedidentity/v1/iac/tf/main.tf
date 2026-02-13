# Create the Azure User-Assigned Managed Identity
resource "azurerm_user_assigned_identity" "main" {
  name                = var.spec.name
  location            = var.spec.region
  resource_group_name = var.spec.resource_group
  tags                = local.final_tags
}

# Create role assignments for the managed identity.
# Each assignment binds a specific Azure RBAC role at a specific scope.
resource "azurerm_role_assignment" "main" {
  for_each = local.role_assignments

  scope                            = each.value.scope
  role_definition_name             = each.value.role_definition_name
  principal_id                     = azurerm_user_assigned_identity.main.principal_id
  skip_service_principal_aad_check = true
}
