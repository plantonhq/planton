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
  description = "AwsConfigAggregator specification"
  type = object({
    region = string
    aggregation = optional(object({
      account_source = optional(object({
        account_ids = list(string)
        all_regions = optional(bool, false)
        regions = optional(list(string), [])
      }))
      organization_source = optional(object({
        role_arn = string
        all_regions = optional(bool, false)
        regions = optional(list(string), [])
      }))
    }))
    authorizations = optional(list(object({
      account_id = string
      authorized_aws_region = string
    })), [])
  })
}
