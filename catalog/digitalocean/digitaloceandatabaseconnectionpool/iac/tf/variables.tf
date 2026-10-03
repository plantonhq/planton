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
  description = "DigitalOceanDatabaseConnectionPool specification"
  type = object({
    cluster   = string
    pool_name = string
    mode      = string
    size      = number
    db_name   = string
    user      = optional(string, "")
  })
}