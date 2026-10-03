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
  description = "AwsRoute53ResolverFirewall specification"
  type = object({
    region = string
    domain_lists = optional(list(object({
      name = string
      domains = optional(list(string), [])
    })), [])
    rules = optional(list(object({
      name = string
      priority = number
      action = string
      domain_list_name = optional(string, "")
      domain_list_id = optional(string, "")
      dns_threat_protection = optional(string, "")
      confidence_threshold = optional(string, "")
      block_response = optional(string, "")
      block_override_domain = optional(string, "")
      block_override_ttl = optional(number)
      firewall_domain_redirection_action = optional(string, "")
      q_type = optional(string, "")
    })), [])
    vpc_associations = optional(list(object({
      name = string
      vpc_id = string
      priority = number
      mutation_protection = optional(string, "")
    })), [])
  })
}
