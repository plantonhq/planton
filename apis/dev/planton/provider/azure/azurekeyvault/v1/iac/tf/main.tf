# Data source to get current Azure client configuration
data "azurerm_client_config" "current" {}

# Create the Azure Key Vault
resource "azurerm_key_vault" "main" {
  name                = local.vault_name
  location            = var.spec.region
  resource_group_name = var.spec.resource_group
  tenant_id           = data.azurerm_client_config.current.tenant_id

  # SKU configuration
  sku_name = var.spec.sku

  # Security settings
  enable_rbac_authorization   = var.spec.enable_rbac_authorization
  purge_protection_enabled    = var.spec.enable_purge_protection
  soft_delete_retention_days  = var.spec.soft_delete_retention_days
  
  # Disable legacy features (not needed with RBAC)
  enabled_for_deployment          = false
  enabled_for_disk_encryption     = false
  enabled_for_template_deployment = false

  # Network ACLs configuration
  network_acls {
    default_action             = local.network_acls.default_action
    bypass                     = local.network_bypass
    ip_rules                   = local.network_acls.ip_rules
    virtual_network_subnet_ids = local.network_acls.virtual_network_subnet_ids
  }

  # Tags
  tags = local.final_tags
}

# Create placeholder secrets (actual values must be set separately)
# These secrets are created with empty values as placeholders
# The actual secret values should be set using Azure CLI, SDK, or Portal after creation
resource "azurerm_key_vault_secret" "secrets" {
  for_each = toset(var.spec.secret_names)

  name         = each.value
  value        = "" # Empty placeholder - must be set separately
  key_vault_id = azurerm_key_vault.main.id

  tags = local.final_tags

  # Lifecycle to prevent destroying secrets with actual values
  lifecycle {
    ignore_changes = [value]
  }
}

