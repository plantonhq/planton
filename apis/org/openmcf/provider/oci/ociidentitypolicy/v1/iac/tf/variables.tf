variable "metadata" {
  description = "Cloud resource metadata"
  type = object({
    name = string
    id   = optional(string, "")
    org  = optional(string, "")
    env  = optional(string, "")
    labels      = optional(map(string), {})
    annotations = optional(map(string), {})
    tags        = optional(list(string), [])
  })
}

variable "spec" {
  description = "OciIdentityPolicy specification"
  type = object({
    compartment_id = object({
      value = string
    })

    name = optional(string, "")

    description = string

    statements = list(string)

    version_date = optional(string, "")
  })
}
