variable "metadata" {
  description = "Catalog object metadata"
  type = object({
    name        = string
    id          = optional(string, "")
    org         = optional(string, "")
    env         = optional(string, "")
    labels      = optional(map(string), {})
    annotations = optional(map(string), {})
    tags        = optional(list(string), [])
  })
}

variable "spec" {
  description = "CloudflareWorkflow specification"
  type = object({
    account_id    = string
    workflow_name = string
    class_name    = string
    script_name   = string
    default_retention = optional(object({
      error_retention   = optional(string, "")
      success_retention = optional(string, "")
    }))
    limits = optional(object({
      steps = optional(number)
    }))
    schedules = optional(list(object({
      cron = string
    })), [])
  })
}
