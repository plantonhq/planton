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
  description = "AzureMongoClusterUser specification"
  type = object({
    mongo_cluster_id = string
    object_id        = string
    principal_type   = string
    roles = list(object({
      database = string
      role     = string
    }))
  })
}