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
  description = "Azure SQL Server specification"
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

    # SQL Server version ("2.0" or "12.0", default "12.0")
    version = optional(string, "12.0")

    # Minimum TLS version ("1.0" or "1.2", default "1.2")
    minimum_tls_version = optional(string, "1.2")

    # Whether the server is accessible over the public internet
    public_network_access_enabled = optional(bool, true)

    # Connection policy ("Default", "Proxy", "Redirect")
    connection_policy = optional(string, "Default")

    # Databases to create (each with its own compute SKU and storage)
    databases = optional(list(object({
      name                = string
      sku_name            = string
      max_size_gb         = optional(number)
      collation           = optional(string, "SQL_Latin1_General_CP1_CI_AS")
      zone_redundant      = optional(bool, false)
      license_type        = optional(string)
      storage_account_type = optional(string, "Geo")
    })), [])

    # Firewall rules for public access
    firewall_rules = optional(list(object({
      name             = string
      start_ip_address = string
      end_ip_address   = string
    })), [])
  })
}
