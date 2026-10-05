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
  description = "CloudflareSecretsStoreSecret specification"
  type = object({
    account_id = string
    store_id   = string
    name       = string
    value      = string
    scopes     = list(string)
    comment    = optional(string, "")
  })
}
