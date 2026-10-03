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
  description = "AwsManagedPrometheusScraper specification"
  type = object({
    region = string
    alias = optional(string, "")
    source_eks = optional(object({
      cluster_arn = string
      subnet_ids = list(string)
      security_group_ids = optional(list(string), [])
    }))
    source_vpc = optional(object({
      subnet_ids = list(string)
      security_group_ids = list(string)
    }))
    amp_workspace_arn = optional(string, "")
    cloudwatch_dataset_arn = optional(string, "")
    scrape_configuration = optional(string, "")
    role_configuration = optional(object({
      source_role_arn = optional(string, "")
      target_role_arn = optional(string, "")
    }))
    logging = optional(object({
      log_group_arn = string
      components = optional(list(string), [])
    }))
  })
}