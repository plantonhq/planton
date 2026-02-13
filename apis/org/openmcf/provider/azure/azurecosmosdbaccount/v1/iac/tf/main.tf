# -----------------------------------------------------------------------------
# Cosmos DB Account
# -----------------------------------------------------------------------------
resource "azurerm_cosmosdb_account" "main" {
  name                = var.spec.name
  location            = var.spec.region
  resource_group_name = var.spec.resource_group
  offer_type          = "Standard"
  kind                = var.spec.kind

  consistency_policy {
    consistency_level       = coalesce(var.spec.consistency_policy.consistency_level, "Session")
    max_interval_in_seconds = var.spec.consistency_policy.consistency_level == "BoundedStaleness" ? coalesce(var.spec.consistency_policy.max_interval_in_seconds, 5) : null
    max_staleness_prefix    = var.spec.consistency_policy.consistency_level == "BoundedStaleness" ? coalesce(var.spec.consistency_policy.max_staleness_prefix, 100) : null
  }

  dynamic "geo_location" {
    for_each = var.spec.geo_locations
    content {
      location          = geo_location.value.location
      failover_priority  = geo_location.value.failover_priority
      zone_redundant     = coalesce(geo_location.value.zone_redundant, false)
    }
  }

  dynamic "capabilities" {
    for_each = local.effective_capabilities
    content {
      name = capabilities.value
    }
  }

  free_tier_enabled                  = var.spec.free_tier_enabled
  automatic_failover_enabled         = var.spec.automatic_failover_enabled
  multiple_write_locations_enabled   = var.spec.multiple_write_locations_enabled
  public_network_access_enabled      = var.spec.public_network_access_enabled
  is_virtual_network_filter_enabled  = var.spec.is_virtual_network_filter_enabled

  dynamic "virtual_network_rule" {
    for_each = var.spec.virtual_network_rules
    content {
      id = virtual_network_rule.value.subnet_id
    }
  }

  ip_range_filter = length(var.spec.ip_range_filter) > 0 ? toset(var.spec.ip_range_filter) : null

  dynamic "backup" {
    for_each = var.spec.backup != null ? [1] : []
    content {
      type                = var.spec.backup.type
      interval_in_minutes  = var.spec.backup.type == "Periodic" ? coalesce(var.spec.backup.interval_in_minutes, 240) : null
      retention_in_hours   = var.spec.backup.type == "Periodic" ? coalesce(var.spec.backup.retention_in_hours, 8) : null
      storage_redundancy   = var.spec.backup.type == "Periodic" ? coalesce(var.spec.backup.storage_redundancy, "Geo") : null
      tier                 = var.spec.backup.type == "Continuous" ? coalesce(var.spec.backup.tier, "Continuous7Days") : null
    }
  }

  mongo_server_version = var.spec.kind == "MongoDB" ? var.spec.mongo_server_version : null

  tags = local.final_tags
}

# -----------------------------------------------------------------------------
# SQL API Databases (when kind = GlobalDocumentDB)
# -----------------------------------------------------------------------------
resource "azurerm_cosmosdb_sql_database" "sql_databases" {
  for_each = var.spec.kind == "GlobalDocumentDB" ? { for db in var.spec.sql_databases : db.name => db } : {}

  name                = each.value.name
  resource_group_name = var.spec.resource_group
  account_name        = azurerm_cosmosdb_account.main.name

  throughput = each.value.throughput != null ? each.value.throughput : null

  dynamic "autoscale_settings" {
    for_each = each.value.autoscale_max_throughput != null ? [1] : []
    content {
      max_throughput = each.value.autoscale_max_throughput
    }
  }
}

# -----------------------------------------------------------------------------
# SQL API Containers (when kind = GlobalDocumentDB)
# -----------------------------------------------------------------------------
resource "azurerm_cosmosdb_sql_container" "sql_containers" {
  for_each = var.spec.kind == "GlobalDocumentDB" ? local.sql_containers_map : {}

  name                = each.value.name
  resource_group_name = var.spec.resource_group
  account_name        = azurerm_cosmosdb_account.main.name
  database_name       = each.value.db_name

  partition_key_paths = each.value.partition_key_paths
  partition_key_kind = each.value.partition_key_kind

  throughput = each.value.throughput != null ? each.value.throughput : null

  dynamic "autoscale_settings" {
    for_each = each.value.autoscale_max_throughput != null ? [1] : []
    content {
      max_throughput = each.value.autoscale_max_throughput
    }
  }

  default_ttl = each.value.default_ttl
}

# -----------------------------------------------------------------------------
# MongoDB API Databases (when kind = MongoDB)
# -----------------------------------------------------------------------------
resource "azurerm_cosmosdb_mongo_database" "mongo_databases" {
  for_each = var.spec.kind == "MongoDB" ? { for db in var.spec.mongo_databases : db.name => db } : {}

  name                = each.value.name
  resource_group_name = var.spec.resource_group
  account_name        = azurerm_cosmosdb_account.main.name

  throughput = each.value.throughput != null ? each.value.throughput : null

  dynamic "autoscale_settings" {
    for_each = each.value.autoscale_max_throughput != null ? [1] : []
    content {
      max_throughput = each.value.autoscale_max_throughput
    }
  }
}

# -----------------------------------------------------------------------------
# MongoDB API Collections (when kind = MongoDB)
# -----------------------------------------------------------------------------
resource "azurerm_cosmosdb_mongo_collection" "mongo_collections" {
  for_each = var.spec.kind == "MongoDB" ? local.mongo_collections_map : {}

  name                = each.value.name
  resource_group_name = var.spec.resource_group
  account_name        = azurerm_cosmosdb_account.main.name
  database_name       = each.value.db_name

  shard_key           = each.value.shard_key
  default_ttl_seconds = each.value.default_ttl_seconds

  throughput = each.value.throughput != null ? each.value.throughput : null

  dynamic "autoscale_settings" {
    for_each = each.value.autoscale_max_throughput != null ? [1] : []
    content {
      max_throughput = each.value.autoscale_max_throughput
    }
  }

  dynamic "index" {
    for_each = each.value.indexes
    content {
      keys   = index.value.keys
      unique = coalesce(index.value.unique, false)
    }
  }
}
