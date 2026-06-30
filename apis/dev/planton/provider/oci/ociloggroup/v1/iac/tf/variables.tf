variable "metadata" {
  type = object({
    name   = string
    id     = optional(string, "")
    org    = optional(string, "")
    env    = optional(string, "")
    labels = optional(map(string), {})
  })
}

variable "spec" {
  type = object({
    compartment_id = object({
      value = string
    })
    description = optional(string, "")

    logs = optional(list(object({
      display_name = string
      log_type     = string
      is_enabled   = optional(bool)
      retention_duration = optional(number)

      configuration = optional(object({
        service = string
        resource = object({
          value = string
        })
        category   = string
        parameters = optional(map(string), {})
        compartment_id = optional(object({
          value = string
        }))
      }))
    })), [])
  })
}
