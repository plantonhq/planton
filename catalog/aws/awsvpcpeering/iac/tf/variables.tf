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
  description = "AwsVpcPeering specification"
  type = object({
    region = string
    request = optional(object({
      vpc_id = string
      peer_vpc_id = string
      peer_owner_id = optional(string, "")
      peer_region = optional(string, "")
      auto_accept = optional(bool, false)
      requester_allow_remote_vpc_dns_resolution = optional(bool, false)
      accepter_allow_remote_vpc_dns_resolution = optional(bool, false)
    }))
    accept = optional(object({
      vpc_peering_connection_id = string
      auto_accept = optional(bool, false)
      accepter_allow_remote_vpc_dns_resolution = optional(bool, false)
    }))
  })
}