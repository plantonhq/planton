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
  description = "CivoDatabase specification"
  type = object({
    # A human-readable name for the database instance (required, max 64 chars)
    db_instance_name = string

    # The database engine: "mysql" or "postgres" (required)
    engine = string

    # The engine version (e.g., "8.0" for MySQL, "16" for PostgreSQL) (required)
    engine_version = string

    # The Civo region (required)
    region = string

    # The plan/size identifier (e.g., "g3.db.small") (required)
    size_slug = string

    # Number of replica nodes (0 = master only, max 4) (optional, default 0)
    replicas = optional(number, 0)

    # Target private network ID or reference (required)
    network_id = object({
      value = optional(string)
      value_from = optional(object({
        kind       = string
        env        = optional(string)
        name       = string
        field_path = string
      }))
    })

    # Firewall rule IDs or references (optional)
    firewall_ids = optional(list(object({
      value = optional(string)
      value_from = optional(object({
        kind       = string
        env        = optional(string)
        name       = string
        field_path = string
      }))
    })), [])

    # Custom storage size in GiB (optional)
    storage_gib = optional(number)

    # Tags for the database instance (optional)
    tags = optional(list(string), [])
  })

  validation {
    condition     = length(var.spec.db_instance_name) <= 64
    error_message = "db_instance_name must not exceed 64 characters."
  }

  validation {
    condition     = contains(["mysql", "postgres"], var.spec.engine)
    error_message = "engine must be either 'mysql' or 'postgres'."
  }

  validation {
    condition     = can(regex("^[0-9]+(\\.[0-9]+)?$", var.spec.engine_version))
    error_message = "engine_version must match the pattern '^[0-9]+(\\.[0-9]+)?$' (e.g., '8.0' or '16')."
  }

  validation {
    condition     = var.spec.replicas >= 0 && var.spec.replicas <= 4
    error_message = "replicas must be between 0 and 4."
  }
}