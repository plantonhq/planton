# Create the Azure SQL logical server.
# The server is an administrative container with no compute or storage.
# Compute and storage live on each database independently.
resource "azurerm_mssql_server" "main" {
  name                = var.spec.name
  location            = var.spec.region
  resource_group_name = var.spec.resource_group

  administrator_login          = var.spec.administrator_login
  administrator_login_password = var.spec.administrator_password

  version                      = var.spec.version
  minimum_tls_version          = var.spec.minimum_tls_version
  public_network_access_enabled = var.spec.public_network_access_enabled
  connection_policy            = var.spec.connection_policy

  tags = local.final_tags
}

# Create databases.
# Each MSSQL database carries its own compute SKU and max storage size.
# This is fundamentally different from PostgreSQL/MySQL where the server
# defines compute and databases are lightweight objects.
resource "azurerm_mssql_database" "databases" {
  for_each = { for db in var.spec.databases : db.name => db }

  name                = each.value.name
  server_id           = azurerm_mssql_server.main.id
  sku_name            = each.value.sku_name
  max_size_gb         = each.value.max_size_gb
  collation           = each.value.collation
  zone_redundant      = each.value.zone_redundant
  license_type        = each.value.license_type
  storage_account_type = each.value.storage_account_type
}

# Create firewall rules.
# Only effective when public_network_access_enabled is true.
resource "azurerm_mssql_firewall_rule" "rules" {
  for_each = { for rule in var.spec.firewall_rules : rule.name => rule }

  name             = each.value.name
  server_id        = azurerm_mssql_server.main.id
  start_ip_address = each.value.start_ip_address
  end_ip_address   = each.value.end_ip_address
}
