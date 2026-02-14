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
  description = "Azure Event Hub Namespace specification"
  type = object({
    region                        = string
    resource_group                = string
    name                          = string
    sku                           = optional(string, "Standard")
    capacity                      = optional(number, 1)
    auto_inflate_enabled          = optional(bool, false)
    maximum_throughput_units       = optional(number)
    zone_redundant                = optional(bool, false)
    minimum_tls_version           = optional(string, "1.2")
    public_network_access_enabled = optional(bool, true)
    event_hubs = optional(list(object({
      name              = string
      partition_count   = number
      message_retention = optional(number, 1)
      consumer_groups = optional(list(object({
        name          = string
        user_metadata = optional(string)
      })), [])
    })), [])
  })
}
