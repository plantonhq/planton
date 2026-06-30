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
  description = "Azure Cosmos DB account specification"
  type = object({
    # The Azure region (primary write region)
    region = string

    # The Azure Resource Group name
    resource_group = string

    # The Cosmos DB account name (globally unique)
    name = string

    # API kind: "GlobalDocumentDB" (default) or "MongoDB"
    kind = optional(string, "GlobalDocumentDB")

    # Consistency policy
    consistency_policy = optional(object({
      consistency_level       = optional(string, "Session")
      max_interval_in_seconds = optional(number, 5)
      max_staleness_prefix    = optional(number, 100)
    }), { consistency_level = "Session" })

    # Geographic locations (at least one required)
    geo_locations = list(object({
      location         = string
      failover_priority = number
      zone_redundant   = optional(bool, false)
    }))

    # Capabilities to enable (e.g., "EnableServerless", "EnableMongo")
    capabilities = optional(list(string), [])

    # Enable free tier (1000 RU/s, 25 GB per subscription)
    free_tier_enabled = optional(bool, false)

    # Enable automatic failover
    automatic_failover_enabled = optional(bool, false)

    # Enable multi-region writes (active-active)
    multiple_write_locations_enabled = optional(bool, false)

    # Public network access
    public_network_access_enabled = optional(bool, true)

    # Virtual network filtering
    is_virtual_network_filter_enabled = optional(bool, false)

    # Virtual network rules (subnet IDs)
    virtual_network_rules = optional(list(object({
      subnet_id = string
    })), [])

    # IP range filter (CIDR or IP addresses)
    ip_range_filter = optional(list(string), [])

    # Backup policy
    backup = optional(object({
      type                 = string
      interval_in_minutes  = optional(number)
      retention_in_hours   = optional(number)
      storage_redundancy   = optional(string)
      tier                 = optional(string)
    }))

    # MongoDB server version (only when kind = MongoDB)
    mongo_server_version = optional(string)

    # SQL API databases and containers (when kind = GlobalDocumentDB)
    sql_databases = optional(list(object({
      name                   = string
      throughput             = optional(number)
      autoscale_max_throughput = optional(number)
      containers = optional(list(object({
        name                   = string
        partition_key_paths    = list(string)
        partition_key_kind     = optional(string, "Hash")
        throughput             = optional(number)
        autoscale_max_throughput = optional(number)
        default_ttl            = optional(number)
      })), [])
    })), [])

    # MongoDB API databases and collections (when kind = MongoDB)
    mongo_databases = optional(list(object({
      name                   = string
      throughput             = optional(number)
      autoscale_max_throughput = optional(number)
      collections = optional(list(object({
        name                   = string
        shard_key              = string
        throughput             = optional(number)
        autoscale_max_throughput = optional(number)
        default_ttl_seconds    = optional(number)
        indexes = optional(list(object({
          keys   = list(string)
          unique = optional(bool, false)
        })), [])
      })), [])
    })), [])
  })
}
