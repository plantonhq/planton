variable "metadata" {
  description = "Catalog object metadata"
  type = object({
    name        = string
    id          = optional(string, "")
    org         = optional(string, "")
    env         = optional(string, "")
    labels      = optional(map(string), {})
    annotations = optional(map(string), {})
    tags        = optional(list(string), [])
  })
}

variable "spec" {
  description = "AwsOpenSearchServerlessCollection specification"
  type = object({
    region                         = string
    description                    = optional(string, "")
    type                           = optional(string)
    standby_replicas               = optional(string)
    collection_group_name          = optional(string, "")
    serverless_vector_acceleration = optional(string, "")
    encryption = optional(object({
      kms_key_arn = optional(string, "")
    }))
    network = optional(object({
      allow_from_public  = optional(bool)
      vpc_endpoint_ids   = optional(list(string), [])
      include_dashboards = optional(bool)
    }))
    data_access = optional(list(object({
      principals             = list(string)
      collection_permissions = optional(list(string), [])
      index_permissions      = optional(list(string), [])
      index_patterns         = optional(list(string), [])
    })), [])
    retention_rules = optional(list(object({
      index_patterns      = list(string)
      min_index_retention = optional(string, "")
      unlimited           = optional(bool, false)
    })), [])
  })
}