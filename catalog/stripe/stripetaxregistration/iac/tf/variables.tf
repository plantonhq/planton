variable "metadata" {
  description = "Cloud resource metadata"
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
  description = "StripeTaxRegistration specification"
  type = object({
    country                = string
    type                   = optional(string, "")
    active_from            = optional(number, 0)
    expires_at             = optional(number)
    place_of_supply_scheme = optional(string, "")
    province               = optional(string, "")
    state                  = optional(string, "")
    jurisdiction           = optional(string, "")
    state_sales_tax_elections = optional(list(object({
      type         = optional(string, "")
      jurisdiction = optional(string, "")
    })), [])
  })
}
