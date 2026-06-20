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
  description = "AwsSubnet specification"
  # StringValueOrRef fields (vpc_id, route_table_id, routes[].target_id) arrive as
  # plain strings: the tofu generator flattens StringValueOrRef to its string value
  # (see pkg/iac/tofu/generators typerules), and the orchestrator resolves any
  # value_from reference before the module runs.
  type = object({
    region = string

    vpc_id = string

    availability_zone = string

    cidr_block = string

    map_public_ip_on_launch = optional(bool, false)

    assign_ipv6_address_on_creation = optional(bool, false)

    ipv6_cidr_block = optional(string, "")

    enable_dns64 = optional(bool, false)

    enable_resource_name_dns_a_record_on_launch = optional(bool, false)

    enable_resource_name_dns_aaaa_record_on_launch = optional(bool, false)

    private_dns_hostname_type_on_launch = optional(string, "")

    route_table_id = optional(string, "")

    routes = optional(list(object({
      destination_cidr_block      = optional(string, "")
      destination_ipv6_cidr_block = optional(string, "")
      destination_prefix_list_id  = optional(string, "")
      target_type                 = string
      target_id                   = string
    })), [])
  })
}
