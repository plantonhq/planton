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
  description = "AzureCognitiveAccountProject specification"
  type = object({
    cognitive_account_id = string
    name                 = string
    region               = string
    identity = object({
      type         = string
      identity_ids = optional(list(string), [])
    })
    description  = optional(string, "")
    display_name = optional(string, "")
    tags         = optional(map(string), {})
  })
}