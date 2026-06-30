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
  description = "Azure Container App Environment specification"
  type = object({
    # The Azure region where the environment will be created
    region = string

    # The Azure Resource Group name
    resource_group = string

    # The name of the Container App Environment
    name = string

    # Existing subnet ID for VNet injection (optional)
    infrastructure_subnet_id = optional(string)

    # Log Analytics Workspace ID for centralized logging (optional)
    log_analytics_workspace_id = optional(string)

    # Enable internal load balancing mode (default: false)
    internal_load_balancer_enabled = optional(bool, false)

    # Enable zone redundancy (default: false)
    zone_redundancy_enabled = optional(bool, false)

    # Dedicated workload profiles (optional)
    workload_profiles = optional(list(object({
      name                  = string
      workload_profile_type = string
      minimum_count         = optional(number)
      maximum_count         = optional(number)
    })), [])
  })
}
