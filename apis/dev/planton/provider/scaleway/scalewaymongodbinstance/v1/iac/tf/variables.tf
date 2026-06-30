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
  description = "Scaleway MongoDB Instance specification"
  type = object({
    # Region where the instance will be created (e.g., "fr-par")
    region = string

    # MongoDB version (e.g., "7.0.12")
    version = string

    # Node type (e.g., "MGDB-PLAY2-NANO", "MGDB-PRO2-S")
    node_type = string

    # Number of nodes: 1 (standalone) or 3 (replica set)
    node_number = number

    # Private Network ID (resolved from StringValueOrRef before Terraform runs)
    private_network_id = optional(string, "")

    # Whether to also create a public endpoint when PN is attached
    enable_public_network = optional(bool, false)

    # Volume configuration
    volume_type      = optional(string, "sbs_5k")
    volume_size_in_gb = optional(number, 0)

    # Snapshot schedule configuration
    enable_snapshot_schedule            = optional(bool, false)
    snapshot_schedule_frequency_hours   = optional(number, 0)
    snapshot_schedule_retention_days    = optional(number, 0)

    # Initial admin user
    admin_user     = string
    admin_password = string

    # Additional users with roles
    users = optional(list(object({
      name     = string
      password = string
      roles = optional(list(object({
        role          = string
        database_name = optional(string, "")
        any_database  = optional(bool, false)
      })), [])
    })), [])

    # MongoDB settings
    settings = optional(map(string), {})
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
