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
  description = "Auth0ResourceServer specification"
  type = object({
    identifier                                      = string
    name                                            = optional(string, "")
    signing_alg                                     = optional(string, "")
    allow_offline_access                            = optional(bool)
    token_lifetime                                  = optional(number, 0)
    token_lifetime_for_web                          = optional(number, 0)
    skip_consent_for_verifiable_first_party_clients = optional(bool)
    enforce_policies                                = optional(bool)
    token_dialect                                   = optional(string, "")
    scopes = optional(list(object({
      name        = string
      description = optional(string, "")
    })), [])
    allow_online_access                         = optional(bool)
    allow_online_access_with_ephemeral_sessions = optional(bool)
    consent_policy                              = optional(string)
    token_lifetime_for_anonymous_access_tokens  = optional(number)
    verification_location                       = optional(string)
    signing_secret                              = string
    access_token = optional(object({
      claims_mapping = optional(object({
        custom_claims = optional(list(object({
          name       = string
          expression = string
        })), [])
      }))
    }))
    authorization_details = optional(list(object({
      type    = optional(string)
      disable = optional(bool)
    })), [])
    authorization_policy = optional(object({
      policy_id = optional(string)
    }))
    proof_of_possession = optional(object({
      disable      = optional(bool)
      mechanism    = optional(string)
      required     = optional(bool)
      required_for = optional(string)
    }))
    subject_type_authorization = optional(object({
      user = optional(object({
        policy = optional(string)
      }))
      client = optional(object({
        policy = optional(string)
      }))
      anonymous_user = optional(object({
        policy = optional(string)
      }))
    }))
    token_encryption = optional(object({
      disable = optional(bool)
      format  = optional(string)
      encryption_key = optional(object({
        algorithm = string
        pem       = string
        kid       = optional(string)
        name      = optional(string)
      }))
    }))
    third_party_client_default_grants = optional(list(object({
      subject_type                = string
      scopes                      = optional(list(string), [])
      authorization_details_types = optional(list(string), [])
      allow_all_scopes            = optional(bool)
      organization_usage          = optional(string)
      allow_any_organization      = optional(bool)
    })), [])
  })
}
