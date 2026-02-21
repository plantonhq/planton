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
  description = "OciPublicIp specification"
  type = object({
    compartment_id = object({
      value = string
    })

    lifetime = string

    display_name = optional(string, "")

    private_ip_id = optional(object({
      value = string
    }), null)

    public_ip_pool_id = optional(object({
      value = string
    }), null)
  })
}
