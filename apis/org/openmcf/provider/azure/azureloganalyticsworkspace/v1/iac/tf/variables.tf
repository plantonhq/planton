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
  description = "Azure Log Analytics Workspace specification"
  type = object({
    # The Azure region where the workspace will be deployed
    region = string

    # The Azure Resource Group name
    resource_group = string

    # The name of the Log Analytics Workspace
    name = string

    # The pricing tier (SKU) of the workspace
    sku = optional(string, "PerGB2018")

    # The number of days to retain data (30-730)
    retention_in_days = optional(number, 30)

    # The daily ingestion quota in GB (-1 for unlimited)
    daily_quota_gb = optional(number, -1)
  })
}
