# The deploying credential's context -- the tenant fallback for Entra
# (AAD) authentication and administrator grants when the spec does not
# pin one.
data "azurerm_client_config" "current" {}

# The PostgreSQL Flexible Server. Databases, firewall rules, server
# parameters, and Entra administrators are separate Azure sub-resources
# declared below -- the server carries compute, storage, networking,
# authentication, encryption, and lifecycle (replica/restore) settings.
resource "azurerm_postgresql_flexible_server" "main" {
  name                = var.spec.server_name
  location            = var.spec.region
  resource_group_name = var.spec.resource_group

  # Lifecycle: how the server comes into existence. The replica/restore
  # modes consume the source server ID (and, for restores, the
  # timestamp); all three are fixed at creation.
  create_mode                       = local.create_mode
  source_server_id                  = var.spec.source_server_id
  point_in_time_restore_time_in_utc = var.spec.point_in_time_restore_time_in_utc

  # Replica promotion (day-2 only): Azure rejects replication_role at
  # creation, and the only legal update value is "None" -- which breaks
  # replication and promotes the replica to a standalone primary.
  replication_role = local.replication_role

  # Password-auth credentials -- omitted (null) on Entra-only servers and
  # on replicas/restores, which inherit from the source. The login is
  # fixed once set; the password rotates in place.
  administrator_login    = local.administrator_login
  administrator_password = local.administrator_password

  # Version is only sent for a fresh server: replicas and restores
  # inherit the source's version. In-place major upgrades are supported
  # (higher versions only); a downgrade replaces the server.
  version = local.version

  # Compute and storage. A replica left unset inherits the source's SKU
  # and size. storage_mb only grows; storage_tier buys IOPS within the
  # size's valid tier range (spec validation mirrors Azure's matrix).
  sku_name          = local.sku_name
  storage_mb        = var.spec.storage_mb
  storage_tier      = local.storage_tier
  auto_grow_enabled = var.spec.auto_grow_enabled

  # Networking: the public endpoint dial and the VNet-injection pair.
  # Azure requires public access OFF on an injected server -- enforced by
  # spec validation before the plan ever runs.
  public_network_access_enabled = var.spec.public_network_access_enabled
  delegated_subnet_id           = var.spec.delegated_subnet_id
  private_dns_zone_id           = var.spec.private_dns_zone_id

  zone = var.spec.zone

  backup_retention_days        = var.spec.backup_retention_days
  geo_redundant_backup_enabled = var.spec.geo_redundant_backup_enabled

  # Authentication mechanisms. Omitting the block applies Azure's default
  # (password on, Entra off). The tenant is only sent when Entra auth is
  # on, falling back to the deploying credential's tenant.
  dynamic "authentication" {
    for_each = var.spec.authentication != null ? [var.spec.authentication] : []
    content {
      active_directory_auth_enabled = authentication.value.active_directory_auth_enabled
      password_auth_enabled         = authentication.value.password_auth_enabled
      tenant_id                     = local.aad_auth_enabled ? local.aad_tenant_id : null
    }
  }

  # High availability: a standby with synchronous replication and
  # automatic failover. The standby zone is fixed at creation -- after
  # that, zones only change via planned failover.
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

  # The server's managed identity. A user-assigned identity is required
  # for customer-managed-key encryption (it unwraps the key).
  dynamic "identity" {
    for_each = var.spec.identity != null ? [var.spec.identity] : []
    content {
      type         = local.identity_type_map[identity.value.type]
      identity_ids = identity.value.identity_ids
    }
  }

  # Customer-managed-key encryption (fixed at creation). The geo-backup
  # pair encrypts the paired-region backup data and is only meaningful
  # with geo-redundant backups.
  dynamic "customer_managed_key" {
    for_each = var.spec.customer_managed_key != null ? [var.spec.customer_managed_key] : []
    content {
      key_vault_key_id                     = customer_managed_key.value.key_vault_key_id
      primary_user_assigned_identity_id    = customer_managed_key.value.primary_user_assigned_identity_id
      geo_backup_key_vault_key_id          = customer_managed_key.value.geo_backup_key_vault_key_id
      geo_backup_user_assigned_identity_id = customer_managed_key.value.geo_backup_user_assigned_identity_id
    }
  }

  # Elastic cluster (PG 17+): a sharded, citus-based cluster of the
  # declared node count. Fixed at creation; the size grows in place but
  # never shrinks.
  dynamic "cluster" {
    for_each = var.spec.cluster != null ? [var.spec.cluster] : []
    content {
      size                  = cluster.value.size
      default_database_name = cluster.value.default_database_name
    }
  }

  tags = local.final_tags
}

# Databases, one Azure sub-resource each. All fields are fixed at
# creation -- changing any replaces just that database, never the server.
resource "azurerm_postgresql_flexible_server_database" "main" {
  for_each = { for db in var.spec.databases : db.name => db }

  name      = each.value.name
  server_id = azurerm_postgresql_flexible_server.main.id
  charset   = each.value.charset
  collation = each.value.collation
}

# Public-endpoint firewall rules. Only meaningful while the public
# endpoint is enabled; Azure ignores them on a VNet-injected server.
resource "azurerm_postgresql_flexible_server_firewall_rule" "main" {
  for_each = { for rule in var.spec.firewall_rules : rule.name => rule }

  name             = each.value.name
  server_id        = azurerm_postgresql_flexible_server.main.id
  start_ip_address = each.value.start_ip_address
  end_ip_address   = each.value.end_ip_address
}

# Server-parameter overrides. Azure applies each as a user override on
# the per-SKU default; destroying the resource resets the parameter to
# its default rather than deleting anything. Static (non-dynamic)
# parameters report "pending restart" until the server restarts.
resource "azurerm_postgresql_flexible_server_configuration" "main" {
  for_each = var.spec.server_parameters

  name      = each.key
  server_id = azurerm_postgresql_flexible_server.main.id
  value     = each.value
}

# Microsoft Entra administrator grants, keyed by the principal's object
# ID. Azure validates principal_type against the directory object, and
# the grant rides the same tenant as the server's Entra auth
# configuration.
resource "azurerm_postgresql_flexible_server_active_directory_administrator" "main" {
  for_each = { for admin in var.spec.aad_administrators : admin.object_id => admin }

  server_name         = azurerm_postgresql_flexible_server.main.name
  resource_group_name = var.spec.resource_group
  tenant_id           = local.aad_tenant_id
  object_id           = each.value.object_id
  principal_name      = each.value.principal_name
  principal_type      = local.principal_type_map[each.value.principal_type]
}
