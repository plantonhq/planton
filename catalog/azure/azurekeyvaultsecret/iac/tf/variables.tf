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
  description = "AzureKeyVaultSecret specification"
  type = object({
    name            = string
    key_vault_id    = string
    value           = string
    content_type    = optional(string, "")
    not_before_date = optional(string)
    expiration_date = optional(string)
    tags            = optional(map(string), {})
  })
}