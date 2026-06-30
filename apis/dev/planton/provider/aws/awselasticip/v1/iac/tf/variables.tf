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
  description = "AwsElasticIp specification"
  type = object({
    region = string
    public_ipv4_pool = optional(string, "")
    address = optional(string, "")
    network_border_group = optional(string, "")
  })
}
