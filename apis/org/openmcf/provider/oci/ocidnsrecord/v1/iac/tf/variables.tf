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
    zone_name_or_id = object({
      value = string
    })
    domain = string
    rtype  = string
    view_id = optional(object({
      value = string
    }))
    items = list(object({
      rdata = string
      ttl   = number
    }))
  })
}
