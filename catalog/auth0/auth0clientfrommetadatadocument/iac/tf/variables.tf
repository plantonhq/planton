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
  description = "Auth0ClientFromMetadataDocument specification"
  type = object({
    external_client_id                                   = string
    external_client_id_version                           = optional(number)
    app_type                                             = optional(string)
    grant_types                                          = optional(list(string), [])
    description                                          = optional(string)
    allowed_origins                                      = optional(list(string), [])
    web_origins                                          = optional(list(string), [])
    oidc_conformant                                      = optional(bool)
    require_proof_of_possession                          = optional(bool)
    skip_non_verifiable_callback_uri_confirmation_prompt = optional(bool)
    redirection_policy                                   = optional(string)
    organization_discovery_methods                       = optional(list(string), [])
    default_organization = optional(object({
      organization_id = string
      flows           = list(string)
    }))
    client_metadata = optional(map(string), {})
    jwt_configuration = optional(object({
      alg                 = optional(string)
      lifetime_in_seconds = optional(number)
    }))
    refresh_token = optional(object({
      rotation_type                = optional(string)
      expiration_type              = optional(string)
      leeway                       = optional(number)
      token_lifetime               = optional(number)
      infinite_token_lifetime      = optional(bool)
      idle_token_lifetime          = optional(number)
      infinite_idle_token_lifetime = optional(bool)
    }))
    token_quota = optional(object({
      client_credentials = object({
        enforce  = optional(bool)
        per_day  = optional(number)
        per_hour = optional(number)
      })
    }))
  })
}
