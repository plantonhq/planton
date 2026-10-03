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
  description = "AwsRoute53ResolverQueryLog specification"
  type = object({
    region = string
    destination_arn = string
    vpc_ids = optional(list(string), [])
  })
}
