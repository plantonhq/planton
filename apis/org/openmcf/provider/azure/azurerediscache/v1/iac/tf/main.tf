# Create the Azure Cache for Redis.
# The SKU family is auto-derived from sku_name:
# "C" (Basic/Standard) or "P" (Premium).
resource "azurerm_redis_cache" "main" {
  name                = var.spec.name
  location            = var.spec.region
  resource_group_name = var.spec.resource_group

  sku_name = var.spec.sku_name
  family   = local.family
  capacity = var.spec.capacity

  redis_version      = var.spec.redis_version
  minimum_tls_version = var.spec.minimum_tls_version

  non_ssl_port_enabled           = var.spec.non_ssl_port_enabled
  public_network_access_enabled  = var.spec.public_network_access_enabled

  # VNet injection (Premium SKU only)
  subnet_id = var.spec.subnet_id

  # Availability zones
  zones = var.spec.zones

  # Redis Cluster sharding (Premium SKU only)
  shard_count = var.spec.shard_count

  redis_configuration {
    maxmemory_policy = var.spec.maxmemory_policy
  }

  # Patch schedules for maintenance windows
  dynamic "patch_schedule" {
    for_each = var.spec.patch_schedules
    content {
      day_of_week        = patch_schedule.value.day_of_week
      start_hour_utc     = patch_schedule.value.start_hour_utc
      maintenance_window = patch_schedule.value.maintenance_window
    }
  }

  tags = local.final_tags
}

# Create firewall rules.
# Only effective when public access is enabled and cache is not VNet-injected.
resource "azurerm_redis_firewall_rule" "rules" {
  for_each = { for rule in var.spec.firewall_rules : rule.name => rule }

  name                = each.value.name
  redis_cache_name    = azurerm_redis_cache.main.name
  resource_group_name = var.spec.resource_group
  start_ip            = each.value.start_ip
  end_ip              = each.value.end_ip
}
