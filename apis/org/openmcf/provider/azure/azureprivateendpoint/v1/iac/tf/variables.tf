variable "metadata" {
  description = "Resource metadata"
  type = object({
    name    = string
    id      = optional(string, "")
    org     = optional(string, "")
    env     = optional(string, "")
    labels  = optional(map(string), {})
    tags    = optional(map(string), {})
    version = optional(string, "")
  })
}

variable "spec" {
  description = "AzurePrivateEndpoint spec"
  type = object({
    region                         = string
    resource_group                 = string
    name                           = string
    subnet_id                      = string
    private_connection_resource_id = string
    subresource_names              = optional(list(string), [])
    private_dns_zone_id            = optional(string, null)
  })
}
