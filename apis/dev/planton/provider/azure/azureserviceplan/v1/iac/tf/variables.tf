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
  description = "Azure Service Plan specification"
  type = object({
    # The Azure region where the Service Plan will be created
    region = string

    # The Azure Resource Group name
    resource_group = string

    # The name of the Service Plan
    name = string

    # The operating system type: "Linux" or "Windows"
    os_type = optional(string, "Linux")

    # The SKU name (e.g., "P1v3", "B1", "Y1", "EP1")
    sku_name = string

    # Number of VM instances (workers) for the plan
    worker_count = optional(number)

    # Enable availability zone balancing
    zone_balancing_enabled = optional(bool, false)

    # Enable per-site scaling
    per_site_scaling_enabled = optional(bool, false)

    # Maximum elastic worker count (for EP* SKUs)
    maximum_elastic_worker_count = optional(number)
  })
}
