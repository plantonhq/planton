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
  description = "DigitalOceanDatabaseUser specification"
  type = object({
    cluster           = string
    user_name         = string
    mysql_auth_plugin = optional(string, "")
    settings = optional(object({
      kafka_acls = optional(list(object({
        topic      = string
        permission = string
      })), [])
      opensearch_acls = optional(list(object({
        index      = string
        permission = string
      })), [])
    }))
  })
}