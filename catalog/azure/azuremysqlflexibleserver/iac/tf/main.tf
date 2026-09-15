# The deploying credential's context -- the tenant fallback for the
# Microsoft Entra administrator grant when the spec does not pin one.
data "azurerm_client_config" "current" {}

# The MySQL Flexible Server. Databases, firewall rules, server
# parameters, and the Entra administrator are separate Azure
# sub-resources declared below -- the server carries compute, storage,
# networking, encryption, and lifecycle (replica/restore) settings.
resource "azurerm_mysql_flexible_server" "main" {
  name                = var.spec.server_name
  location            = var.spec.region
  resource_group_name = var.spec.resource_group

  # Lifecycle: how the server comes into existence. The replica/restore
  # modes consume the source server ID (and, for point-in-time restore,
  # the timestamp); all three are fixed at creation. GEO_RESTORE takes
  # no timestamp -- it restores the latest geo-replicated backup.
  create_mode                       = local.create_mode
  source_server_id                  = var.spec.source_server_id
  point_in_time_restore_time_in_utc = var.spec.point_in_time_restore_time_in_utc

  # Replica promotion (day-2 only): Azure rejects replication_role at
  # creation, and the only legal update value is "None" -- which breaks
  # replication and promotes the replica to a standalone primary.
  replication_role = local.replication_role

  # Password-auth credentials -- omitted (null) on replicas/restores,
  # which inherit from the source. MySQL always keeps password auth on
  # (unlike PostgreSQL Flexible Server, it cannot be disabled). The
  # login is fixed once set; the password rotates in place.
  administrator_login    = local.administrator_login
  administrator_password = local.administrator_password

  # Version is only sent for a fresh server: replicas and restores
  # inherit the source's version. 5.7 -> 8.0.21 upgrades in place
  # (irreversible); a downgrade replaces the server.
  version = local.version

  # Compute. A replica left unset inherits the source's SKU.
  sku_name = local.sku_name

  # Storage: capacity only grows (shrinking replaces the server), and
  # iops is mutually exclusive with elastic IO scaling -- enforced by
  # spec validation before the plan ever runs.
  dynamic "storage" {
    for_each = var.spec.storage != null ? [var.spec.storage] : []
    content {
      size_gb             = storage.value.size_gb
      iops                = storage.value.iops
      auto_grow_enabled   = storage.value.auto_grow_enabled
      io_scaling_enabled  = storage.value.io_scaling_enabled
      log_on_disk_enabled = storage.value.log_on_disk_enabled
    }
  }

  # Networking: unset public_network_access lets Azure derive the value
  # (Enabled publicly, Disabled when VNet-injected); the injection pair
  # is fixed at creation.
  public_network_access = local.public_network_access
  delegated_subnet_id   = var.spec.delegated_subnet_id
  private_dns_zone_id   = var.spec.private_dns_zone_id

  # The primary's zone can only change via a planned failover that swaps
  # zone and standby_availability_zone -- Azure rejects an independent
  # zone change.
  zone = var.spec.zone

  backup_retention_days        = var.spec.backup_retention_days
  geo_redundant_backup_enabled = var.spec.geo_redundant_backup_enabled

  # High availability: a standby with synchronous replication and
  # automatic failover. Not supported on burstable SKUs or replicas.
  dynamic "high_availability" {
    for_each = var.spec.high_availability != null ? [var.spec.high_availability] : []
    content {
      mode                      = local.ha_mode_map[high_availability.value.mode]
      standby_availability_zone = high_availability.value.standby_availability_zone
    }
  }

  # The weekly patching window. Azure applies it via a secondary update
  # right after creation; omitting the block leaves the window
  # system-managed.
  dynamic "maintenance_window" {
    for_each = var.spec.maintenance_window != null ? [var.spec.maintenance_window] : []
    content {
      day_of_week  = coalesce(maintenance_window.value.day_of_week, 0)
      start_hour   = coalesce(maintenance_window.value.start_hour, 0)
      start_minute = coalesce(maintenance_window.value.start_minute, 0)
    }
  }

  # The server's identities. MySQL Flexible Server supports
  # user-assigned identities only -- they unwrap the CMK and back the
  # Entra administrator grant.
  dynamic "identity" {
    for_each = length(var.spec.user_assigned_identity_ids) > 0 ? [1] : []
    content {
      type         = "UserAssigned"
      identity_ids = var.spec.user_assigned_identity_ids
    }
  }

  # Customer-managed-key encryption. The geo-backup pair encrypts the
  # paired-region backup data and is only meaningful with geo-redundant
  # backups.
  dynamic "customer_managed_key" {
    for_each = var.spec.customer_managed_key != null ? [var.spec.customer_managed_key] : []
    content {
      key_vault_key_id                     = customer_managed_key.value.key_vault_key_id
      primary_user_assigned_identity_id    = customer_managed_key.value.primary_user_assigned_identity_id
      geo_backup_key_vault_key_id          = customer_managed_key.value.geo_backup_key_vault_key_id
      geo_backup_user_assigned_identity_id = customer_managed_key.value.geo_backup_user_assigned_identity_id
    }
  }

  tags = local.final_tags
}

# Databases, one Azure sub-resource each. All fields are fixed at
# creation -- changing any replaces just that database, never the server.
resource "azurerm_mysql_flexible_database" "main" {
  for_each = { for db in var.spec.databases : db.name => db }

  name                = each.value.name
  resource_group_name = var.spec.resource_group
  server_name         = azurerm_mysql_flexible_server.main.name
  charset             = each.value.charset
  collation           = each.value.collation
}

# Public-endpoint firewall rules. Only meaningful while the public
# endpoint is enabled; Azure ignores them on a VNet-injected server.
resource "azurerm_mysql_flexible_server_firewall_rule" "main" {
  for_each = { for rule in var.spec.firewall_rules : rule.name => rule }

  name                = each.value.name
  resource_group_name = var.spec.resource_group
  server_name         = azurerm_mysql_flexible_server.main.name
  start_ip_address    = each.value.start_ip_address
  end_ip_address      = each.value.end_ip_address
}

# Server-parameter overrides. Azure applies each as a user override on
# the per-SKU default; destroying the resource resets the parameter to
# its default rather than deleting anything. Static (non-dynamic)
# parameters report "pending restart" until the server restarts.
resource "azurerm_mysql_flexible_server_configuration" "main" {
  for_each = var.spec.server_parameters

  name                = each.key
  resource_group_name = var.spec.resource_group
  server_name         = azurerm_mysql_flexible_server.main.name
  value               = each.value
}

# The single Microsoft Entra administrator (MySQL supports exactly one
# per server). The grant is backed by a user-assigned identity attached
# to the server, which Azure uses to read directory objects when
# validating Entra logins; the tenant falls back to the deploying
# credential's.
resource "azurerm_mysql_flexible_server_active_directory_administrator" "main" {
  count = var.spec.aad_administrator != null ? 1 : 0

  server_id   = azurerm_mysql_flexible_server.main.id
  identity_id = var.spec.aad_administrator.identity_id
  login       = var.spec.aad_administrator.login
  object_id   = var.spec.aad_administrator.object_id
  tenant_id   = local.aad_tenant_id
}
