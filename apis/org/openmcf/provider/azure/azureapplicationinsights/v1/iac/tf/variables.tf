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
  description = "Azure Application Insights specification"
  type = object({
    region              = string
    resource_group      = string
    name                = string
    application_type    = optional(string, "web")
    workspace_id        = string
    retention_in_days   = optional(number, 90)
    daily_data_cap_in_gb = optional(number, 100)
    sampling_percentage = optional(number, 100)
  })
}
