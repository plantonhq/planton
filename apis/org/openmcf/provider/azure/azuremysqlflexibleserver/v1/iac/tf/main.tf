# Create the Azure MySQL Flexible Server.
# Network mode is derived from the presence of delegated_subnet_id:
# - Subnet set --> private VNet access, public access disabled
# - Subnet not set --> public access, firewall rules apply
resource "azurerm_mysql_flexible_server" "main" {
  name                = var.spec.name
  location            = var.spec.region
  resource_group_name = var.spec.resource_group

  administrator_login    = var.spec.administrator_login
  administrator_password = var.spec.administrator_password

  version  = var.spec.version
  sku_name = var.spec.sku_name

  # MySQL uses a storage block (not flat storage_mb like PostgreSQL).
  storage {
    size_gb           = var.spec.storage_size_gb
    auto_grow_enabled = var.spec.auto_grow_enabled
  }

  delegated_subnet_id = local.is_vnet_integrated ? var.spec.delegated_subnet_id : null
  private_dns_zone_id = local.is_vnet_integrated && var.spec.private_dns_zone_id != null ? var.spec.private_dns_zone_id : null

  zone = var.spec.zone

  backup_retention_days        = var.spec.backup_retention_days
  geo_redundant_backup_enabled = var.spec.geo_redundant_backup_enabled

  dynamic "high_availability" {
    for_each = var.spec.high_availability != null ? [var.spec.high_availability] : []
    content {
      mode                      = high_availability.value.mode
      standby_availability_zone = high_availability.value.standby_availability_zone
    }
  }

  tags = local.final_tags
}

# Create databases.
# MySQL uses server_name + resource_group_name (NOT server_id like PostgreSQL).
resource "azurerm_mysql_flexible_database" "databases" {
  for_each = { for db in var.spec.databases : db.name => db }

  name                = each.value.name
  server_name         = azurerm_mysql_flexible_server.main.name
  resource_group_name = var.spec.resource_group
  charset             = each.value.charset
  collation           = each.value.collation
}

# Create firewall rules.
# MySQL uses server_name + resource_group_name (NOT server_id like PostgreSQL).
# Only effective in public access mode (when delegated_subnet_id is not set).
resource "azurerm_mysql_flexible_server_firewall_rule" "rules" {
  for_each = { for rule in var.spec.firewall_rules : rule.name => rule }

  name                = each.value.name
  server_name         = azurerm_mysql_flexible_server.main.name
  resource_group_name = var.spec.resource_group
  start_ip_address    = each.value.start_ip_address
  end_ip_address      = each.value.end_ip_address
}
