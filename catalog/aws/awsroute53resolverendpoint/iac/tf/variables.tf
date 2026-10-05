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
  description = "AwsRoute53ResolverEndpoint specification"
  type = object({
    region = string
    direction = string
    ip_addresses = list(object({
      subnet_id = string
      ip = optional(string, "")
      ipv6 = optional(string, "")
    }))
    security_group_ids = list(string)
    endpoint_type = optional(string, "")
    protocols = optional(list(string), [])
    rni_enhanced_metrics_enabled = optional(bool)
    target_name_server_metrics_enabled = optional(bool)
    rules = optional(list(object({
      name = string
      domain_name = string
      rule_type = string
      target_ips = optional(list(object({
        ip = optional(string, "")
        ipv6 = optional(string, "")
        port = optional(number, 0)
        protocol = optional(string, "")
      })), [])
      vpc_ids = optional(list(string), [])
    })), [])
  })
}
