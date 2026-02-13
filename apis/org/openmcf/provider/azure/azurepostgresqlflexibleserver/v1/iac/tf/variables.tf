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
  description = "Azure PostgreSQL Flexible Server specification"
  type = object({
    # The Azure region
    region = string

    # The Azure Resource Group name
    resource_group = string

    # The server name (globally unique)
    name = string

    # Administrator login name
    administrator_login = string

    # Administrator password
    administrator_password = string

    # PostgreSQL version ("12"-"17", default "16")
    version = optional(string, "16")

    # SKU name (e.g., "GP_Standard_D2s_v3")
    sku_name = string

    # Storage size in MB (min 32768)
    storage_mb = number

    # Enable storage auto-grow
    auto_grow_enabled = optional(bool, false)

    # Delegated subnet ID for VNet integration (optional)
    delegated_subnet_id = optional(string)

    # Private DNS zone ID (optional)
    private_dns_zone_id = optional(string)

    # Availability zone for primary ("1", "2", "3")
    zone = optional(string)

    # High availability configuration
    high_availability = optional(object({
      # Mode: "ZoneRedundant" or "SameZone"
      mode = string

      # Standby availability zone
      standby_availability_zone = optional(string)
    }))

    # Backup retention days (7-35, default 7)
    backup_retention_days = optional(number, 7)

    # Enable geo-redundant backup
    geo_redundant_backup_enabled = optional(bool, false)

    # Databases to create
    databases = optional(list(object({
      name      = string
      charset   = optional(string, "UTF8")
      collation = optional(string, "en_US.utf8")
    })), [])

    # Firewall rules for public access
    firewall_rules = optional(list(object({
      name             = string
      start_ip_address = string
      end_ip_address   = string
    })), [])
  })
}
