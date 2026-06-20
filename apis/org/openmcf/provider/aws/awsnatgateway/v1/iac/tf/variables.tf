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
  description = "AwsNatGateway specification"
  # StringValueOrRef fields (subnet_id, allocation_id, secondary_allocation_ids[])
  # arrive as plain strings: the tofu generator flattens StringValueOrRef to its
  # string value (singular -> string, repeated -> list(string); see
  # pkg/iac/tofu/generators typerules), and the orchestrator resolves any
  # value_from reference before the module runs.
  type = object({
    region = string

    connectivity_type = string

    subnet_id = string

    allocation_id = optional(string, "")

    private_ip = optional(string, "")

    secondary_allocation_ids = optional(list(string), [])

    secondary_private_ip_addresses = optional(list(string), [])

    secondary_private_ip_address_count = optional(number, 0)
  })
}
