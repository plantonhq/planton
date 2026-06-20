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
  description = "AwsVpc specification"
  type = object({
    region = string

    cidr_block = optional(string, "")

    secondary_ipv4_cidr_blocks = optional(list(string), [])

    ipv4_ipam_pool_id = optional(string, "")

    ipv4_netmask_length = optional(number, 0)

    instance_tenancy = optional(string, "")

    # AWS defaults DNS support to on; unset keeps it on (see spec.proto).
    enable_dns_support = optional(bool, true)

    enable_dns_hostnames = optional(bool, false)

    enable_network_address_usage_metrics = optional(bool, false)

    assign_generated_ipv6_cidr_block = optional(bool, false)

    ipv6_cidr_block = optional(string, "")

    ipv6_cidr_block_network_border_group = optional(string, "")

    ipv6_ipam_pool_id = optional(string, "")

    ipv6_netmask_length = optional(number, 0)
  })
}
