variable "metadata" {
  description = "Cloud resource metadata"
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
  description = "AwsSecurityGroup specification"
  type = object({
    region = string
    vpc_id = string
    description = string
    ingress = optional(list(object({
      protocol = string
      from_port = optional(number, 0)
      to_port = optional(number, 0)
      ipv4_cidrs = optional(list(string), [])
      ipv6_cidrs = optional(list(string), [])
      source_security_group_ids = optional(list(string), [])
      destination_security_group_ids = optional(list(string), [])
      self_reference = optional(bool, false)
      description = optional(string, "")
    })), [])
    egress = optional(list(object({
      protocol = string
      from_port = optional(number, 0)
      to_port = optional(number, 0)
      ipv4_cidrs = optional(list(string), [])
      ipv6_cidrs = optional(list(string), [])
      source_security_group_ids = optional(list(string), [])
      destination_security_group_ids = optional(list(string), [])
      self_reference = optional(bool, false)
      description = optional(string, "")
    })), [])
  })
}
