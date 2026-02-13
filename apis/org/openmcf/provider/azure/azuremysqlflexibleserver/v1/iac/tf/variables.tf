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
  description = "Azure MySQL Flexible Server specification"
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

    # MySQL version ("5.7" or "8.0.21", default "8.0.21")
    version = optional(string, "8.0.21")

    # SKU name (e.g., "GP_Standard_D2ds_v4")
    sku_name = string

    # Storage size in GB (min 20)
    storage_size_gb = number

    # Enable storage auto-grow (MySQL defaults to true)
    auto_grow_enabled = optional(bool, true)

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

    # Backup retention days (1-35, default 7)
    backup_retention_days = optional(number, 7)

    # Enable geo-redundant backup
    geo_redundant_backup_enabled = optional(bool, false)

    # Databases to create
    databases = optional(list(object({
      name      = string
      charset   = optional(string, "utf8mb4")
      collation = optional(string, "utf8mb4_0900_ai_ci")
    })), [])

    # Firewall rules for public access
    firewall_rules = optional(list(object({
      name             = string
      start_ip_address = string
      end_ip_address   = string
    })), [])
  })
}
