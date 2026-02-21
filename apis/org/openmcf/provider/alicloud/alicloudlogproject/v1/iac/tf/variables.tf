variable "metadata" {
  description = "Metadata for the resource, including name and labels"
  type = object({
    name   = string
    id     = optional(string)
    org    = optional(string)
    env    = optional(string)
    labels = optional(map(string))
    tags   = optional(list(string))
  })
}

variable "spec" {
  description = "Alibaba Cloud SLS Log Project specification"
  type = object({
    region            = string
    project_name      = string
    description       = optional(string, "")
    resource_group_id = optional(string, "")
    tags              = optional(map(string), {})
    log_stores = optional(list(object({
      name                  = string
      retention_days        = optional(number, 30)
      shard_count           = optional(number, 2)
      auto_split            = optional(bool, true)
      max_split_shard_count = optional(number, 64)
      enable_index          = optional(bool, true)
      append_meta           = optional(bool, true)
    })), [])
  })

  validation {
    condition     = length(var.spec.project_name) >= 3 && length(var.spec.project_name) <= 63
    error_message = "project_name must be between 3 and 63 characters."
  }
}
