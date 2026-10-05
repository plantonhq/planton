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
  description = "AwsCloudMapNamespace specification"
  type = object({
    region = string
    type = string
    vpc_id = optional(string, "")
    description = optional(string, "")
    services = optional(list(object({
      name = string
      description = optional(string, "")
      dns_config = optional(object({
        records = list(object({
          type = string
          ttl = number
        }))
        routing_policy = optional(string, "")
      }))
      health_check_config = optional(object({
        type = optional(string, "")
        resource_path = optional(string, "")
        failure_threshold = optional(number, 0)
      }))
      health_check_custom_config = optional(object({}))
      force_destroy = optional(bool, false)
      instances = optional(list(object({
        instance_id = string
        ip = optional(string, "")
        ipv6 = optional(string, "")
        port = optional(number, 0)
        cname = optional(string, "")
        alias_dns_name = optional(string, "")
        ec2_instance_id = optional(string, "")
        custom_attributes = optional(map(string), {})
      })), [])
    })), [])
  })
}
