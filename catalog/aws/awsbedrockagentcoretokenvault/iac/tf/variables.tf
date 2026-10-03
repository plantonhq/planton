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
  description = "AwsBedrockAgentCoreTokenVault specification"
  type = object({
    region         = string
    token_vault_id = optional(string, "")
    key_type       = string
    kms_key_arn    = optional(string, "")
  })
}
