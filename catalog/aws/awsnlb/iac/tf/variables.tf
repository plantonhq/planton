variable "metadata" {
  description = "Catalog object metadata"
  type = object({
    name = string
    id = optional(string, "")
    org = optional(string, "")
    env = optional(string, "")
    labels = optional(map(string), {})
    annotations = optional(map(string), {})
    tags = optional(list(string), [])
  })
}

variable "spec" {
  description = "AwsNlb specification"
  type = object({
    region = string
    subnet_mappings = list(object({
      subnet_id = string
      allocation_id = optional(string, "")
      private_ipv4_address = optional(string, "")
      ipv6_address = optional(string, "")
    }))
    security_groups = optional(list(string), [])
    internal = optional(bool, false)
    delete_protection_enabled = optional(bool, false)
    cross_zone_load_balancing_enabled = optional(bool, false)
    ip_address_type = optional(string, "")
    dns_record_client_routing_policy = optional(string, "")
    zonal_shift_enabled = optional(bool, false)
    enforce_security_group_inbound_rules_on_private_link_traffic = optional(string, "")
    access_logs = optional(object({
      bucket = string
      prefix = optional(string, "")
    }))
    dns = optional(object({
      enabled = optional(bool, false)
      route53_zone_id = optional(string, "")
      hostnames = optional(list(string), [])
    }))
    minimum_load_balancer_capacity_units = optional(number)
    secondary_ips_auto_assigned_per_subnet = optional(number)
    enable_prefix_for_ipv6_source_nat = optional(string, "")
  })
}