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
    kafka_settings = optional(object({
      auto_create_topics_enable = optional(bool)
      log_retention_hours       = optional(number)
      num_partitions            = optional(number)
    }))
    kms_key_id = optional(object({
      value = string
    }))
    private_endpoint_settings = optional(object({
      subnet_id = object({
        value = string
      })
      nsg_ids = optional(list(object({
        value = string
      })), [])
      private_endpoint_ip = optional(string, "")
    }))
    streams = optional(list(object({
      name               = string
      partitions         = number
      retention_in_hours = optional(number)
    })), [])
  })
}
