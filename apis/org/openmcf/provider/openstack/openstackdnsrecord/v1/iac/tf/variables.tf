variable "metadata" {
  description = "Metadata for the resource"
  type = object({
    name    = string
    id      = optional(string)
    org     = optional(string)
    env     = optional(string)
    labels  = optional(map(string))
    tags    = optional(list(string))
    version = optional(object({ id = string, message = string }))
  })
}

variable "spec" {
  description = "OpenStackDnsRecordSpec"
  type = object({
    zone_id = object({
      value = string
    })
    record_name = string
    type        = number
    values      = list(string)
    ttl         = optional(number)
    description = optional(string, "")
    region      = optional(string, "")
  })
}
