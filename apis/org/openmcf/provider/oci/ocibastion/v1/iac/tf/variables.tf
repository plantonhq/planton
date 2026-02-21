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
  description = "OciBastion specification"
  type = object({
    compartment_id = object({
      value = string
    })

    target_subnet_id = object({
      value = string
    })

    display_name = optional(string, "")

    client_cidr_block_allow_list = optional(list(string), [])

    max_session_ttl_in_seconds = optional(number, null)

    is_dns_proxy_enabled = optional(bool, null)
  })
}
