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
  description = "DigitalOceanDatabaseReplica specification"
  type = object({
    cluster          = string
    replica_name     = string
    region           = string
    size             = string
    vpc              = optional(string, "")
    storage_size_mib = optional(number, 0)
    tags             = optional(list(string), [])
  })
}