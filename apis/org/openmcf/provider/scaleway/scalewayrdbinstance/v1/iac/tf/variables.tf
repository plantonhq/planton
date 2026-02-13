variable "metadata" {
  description = "Metadata for the resource, including name and labels"
  type = object({
    name    = string
    id      = optional(string)
    org     = optional(string)
    env     = optional(string)
    labels  = optional(map(string))
    tags    = optional(list(string))
    version = optional(object({ id = string, message = string }))
  })
}

variable "spec" {
  description = "Scaleway RDB Instance specification"
  type = object({
    # Region where the instance will be created (e.g., "fr-par")
    region = string

    # Engine and version (e.g., "PostgreSQL-16", "MySQL-8")
    engine = string

    # Node type (e.g., "DB-DEV-S", "db-gp-xs")
    node_type = string

    # Private Network ID (resolved from StringValueOrRef before Terraform runs)
    private_network_id = optional(string, "")

    # High availability
    is_ha_cluster = optional(bool, false)

    # Storage configuration
    volume_type      = optional(string, "lssd")
    volume_size_in_gb = optional(number, 0)

    # Backup configuration
    disable_backup                   = optional(bool, false)
    backup_schedule_frequency_hours  = optional(number, 0)
    backup_schedule_retention_days   = optional(number, 0)

    # Security
    encryption_at_rest = optional(bool, false)

    # ACL rules (empty = no ACL resource created)
    acl_rules = optional(list(object({
      ip          = string
      description = optional(string, "")
    })), [])

    # Initial admin user
    admin_user     = string
    admin_password = string

    # Logical databases to create
    databases = optional(list(object({
      name = string
    })), [])

    # Additional users with optional privileges
    users = optional(list(object({
      name     = string
      password = string
      is_admin = optional(bool, false)
      privileges = optional(list(object({
        database_name = string
        permission    = string
      })), [])
    })), [])

    # Engine settings
    settings      = optional(map(string), {})
    init_settings = optional(map(string), {})
  })
}

variable "scaleway_access_key" {
  description = "Scaleway access key for API authentication"
  type        = string
  sensitive   = true
}

variable "scaleway_secret_key" {
  description = "Scaleway secret key for API authentication"
  type        = string
  sensitive   = true
}

variable "scaleway_project_id" {
  description = "Scaleway project ID (optional, defaults from provider)"
  type        = string
  default     = ""
}

variable "scaleway_organization_id" {
  description = "Scaleway organization ID (optional, defaults from provider)"
  type        = string
  default     = ""
}
