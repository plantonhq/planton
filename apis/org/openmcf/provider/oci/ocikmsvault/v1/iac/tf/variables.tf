variable "metadata" {
  description = "Cloud resource metadata"
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
  description = "OciKmsVault specification"
  type = object({
    compartment_id = object({
      value = string
    })

    display_name = optional(string, "")

    vault_type = string

    external_key_manager_metadata = optional(object({
      external_vault_endpoint_url = string
      oauth_metadata = object({
        client_app_id        = string
        client_app_secret    = string
        idcs_account_name_url = string
      })
      private_endpoint_id = string
    }), null)
  })
}
