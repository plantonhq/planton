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
  description = "DigitalOceanDatabaseFirewall specification"
  type = object({
    cluster                = string
    ip_rules               = optional(list(string), [])
    droplet_ids            = optional(list(string), [])
    kubernetes_cluster_ids = optional(list(string), [])
    app_ids                = optional(list(string), [])
    tags                   = optional(list(string), [])
  })
}