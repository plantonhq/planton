# Create the Azure Storage Account
resource "azurerm_storage_account" "main" {
  name                     = local.storage_account_name
  location                 = var.spec.region
  resource_group_name      = var.spec.resource_group
  account_kind             = var.spec.account_kind
  account_tier             = var.spec.account_tier
  account_replication_type = var.spec.replication_type
  access_tier              = var.spec.access_tier

  # Security settings
  enable_https_traffic_only       = var.spec.enable_https_traffic_only
  min_tls_version                 = var.spec.min_tls_version
  allow_nested_items_to_be_public = false

  # Network rules configuration
  network_rules {
    default_action             = local.network_rules.default_action
    bypass                     = local.network_bypass
    ip_rules                   = local.network_rules.ip_rules
    virtual_network_subnet_ids = local.network_rules.virtual_network_subnet_ids
  }

  # Blob properties configuration
  blob_properties {
    versioning_enabled = local.blob_properties.enable_versioning

    dynamic "delete_retention_policy" {
      for_each = local.blob_properties.soft_delete_retention_days > 0 ? [1] : []
      content {
        days = local.blob_properties.soft_delete_retention_days
      }
    }

    dynamic "container_delete_retention_policy" {
      for_each = local.blob_properties.container_soft_delete_retention_days > 0 ? [1] : []
      content {
        days = local.blob_properties.container_soft_delete_retention_days
      }
    }
  }

  # Tags
  tags = local.final_tags
}

# Create blob containers
resource "azurerm_storage_container" "containers" {
  for_each = local.containers_map

  name                  = each.value.name
  storage_account_name  = azurerm_storage_account.main.name
  container_access_type = each.value.access_type
}
