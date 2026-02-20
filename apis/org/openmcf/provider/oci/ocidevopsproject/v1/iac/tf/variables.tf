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
    notification_topic_id = object({
      value = string
    })
    description = optional(string, "")
  })
}
