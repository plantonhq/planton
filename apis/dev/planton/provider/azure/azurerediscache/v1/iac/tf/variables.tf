variable "metadata" {
  description = "Metadata for the resource, including name and labels"
  type = object({
    name    = string,
    id      = optional(string),
    org     = optional(string),
    env     = optional(string),
    labels  = optional(map(string)),
    tags    = optional(list(string)),
    version = optional(object({ id = string, message = string }))
  })
}

variable "spec" {
  description = "Azure Cache for Redis specification"
  type = object({
    # The Azure region
    region = string

    # The Azure Resource Group name
    resource_group = string

    # The cache name (globally unique)
    name = string

    # SKU tier: "Basic", "Standard", "Premium"
    sku_name = optional(string, "Standard")

    # Cache capacity (0-6 for Basic/Standard, 1-5 for Premium)
    capacity = number

    # Redis engine version ("4", "6")
    redis_version = optional(string, "6")

    # Subnet ID for VNet injection (Premium only)
    subnet_id = optional(string)

    # Availability zones
    zones = optional(list(string), [])

    # Number of shards for Redis Cluster (Premium only, 1-10)
    shard_count = optional(number)

    # Enable non-SSL port 6379
    non_ssl_port_enabled = optional(bool, false)

    # Minimum TLS version
    minimum_tls_version = optional(string, "1.2")

    # Whether the cache is publicly accessible
    public_network_access_enabled = optional(bool, true)

    # Memory eviction policy
    maxmemory_policy = optional(string, "volatile-lru")

    # Patch (maintenance) schedules
    patch_schedules = optional(list(object({
      day_of_week        = string
      start_hour_utc     = optional(number, 0)
      maintenance_window = optional(string, "PT5H")
    })), [])

    # Firewall rules for IP-based access control
    firewall_rules = optional(list(object({
      name     = string
      start_ip = string
      end_ip   = string
    })), [])
  })
}
