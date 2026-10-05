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
  description = "AwsNetworkAcl specification"
  type = object({
    region = string
    vpc_id = string
    ingress = optional(list(object({
      rule_no = optional(number, 0)
      action = optional(string, "")
      protocol = string
      cidr_block = optional(string, "")
      ipv6_cidr_block = optional(string, "")
      from_port = optional(number, 0)
      to_port = optional(number, 0)
      icmp_type = optional(number)
      icmp_code = optional(number)
    })), [])
    egress = optional(list(object({
      rule_no = optional(number, 0)
      action = optional(string, "")
      protocol = string
      cidr_block = optional(string, "")
      ipv6_cidr_block = optional(string, "")
      from_port = optional(number, 0)
      to_port = optional(number, 0)
      icmp_type = optional(number)
      icmp_code = optional(number)
    })), [])
    subnet_ids = optional(list(string), [])
  })
}