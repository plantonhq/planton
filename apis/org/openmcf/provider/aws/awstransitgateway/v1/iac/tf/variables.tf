variable "provider_config" {
  description = "AWS provider configuration"
  type = object({
    account_id        = string
    access_key_id     = string
    secret_access_key = string
    region            = string
    session_token     = optional(string, "")
  })
}

variable "metadata" {
  description = "Resource metadata"
  type = object({
    name = string
    org  = optional(string, "")
    env  = optional(string, "")
    id   = optional(string, "")
  })
}

variable "spec" {
  description = "AwsTransitGateway spec"
  type = object({
    # The AWS region where the resource will be created.
    region                            = string
    description                       = optional(string, "")
    amazon_side_asn                   = optional(number, 64512)
    default_route_table_association   = optional(bool, true)
    default_route_table_propagation   = optional(bool, true)
    dns_support                       = optional(bool, true)
    vpn_ecmp_support                  = optional(bool, true)
    auto_accept_shared_attachments    = optional(bool, false)
    security_group_referencing_support = optional(bool, false)
    multicast_support                 = optional(bool, false)
    transit_gateway_cidr_blocks       = optional(list(string), [])
    vpc_attachments = list(object({
      name                           = string
      vpc_id                         = string
      subnet_ids                     = list(string)
      dns_support                    = optional(bool, true)
      ipv6_support                   = optional(bool, false)
      appliance_mode_support         = optional(bool, false)
      default_route_table_association = optional(bool, true)
      default_route_table_propagation = optional(bool, true)
    }))
  })
}
