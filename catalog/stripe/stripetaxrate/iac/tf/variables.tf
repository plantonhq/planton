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
  description = "StripeTaxRate specification"
  type = object({
    display_name = string
    percentage   = optional(number, 0)
    inclusive    = optional(bool, false)
    country      = optional(string, "")
    state        = optional(string, "")
    jurisdiction = optional(string, "")
    description  = optional(string, "")
    tax_type     = optional(string, "")
    active       = optional(bool)
    metadata     = optional(map(string), {})
  })
}
