# Auth0ResourceServer Variables
# This file defines all input variables for the Auth0ResourceServer Terraform module
# These variables map to the Auth0ResourceServerSpec protobuf message

variable "metadata" {
  description = "Metadata for the resource, including name and labels"
  type = object({
    name    = string
    id      = optional(string)
    org     = optional(string)
    env     = optional(string)
    labels  = optional(map(string))
    tags    = optional(list(string))
    version = optional(object({ id = string, message = string }))
  })
}

variable "spec" {
  description = "Auth0ResourceServer specification"
  type = object({
    # identifier is the unique identifier for the resource server.
    # This value is used as the "audience" parameter for authorization calls.
    # Required. Cannot be changed once set.
    identifier = string

    # name is a friendly display name for the resource server.
    name = optional(string)

    # signing_alg is the algorithm used to sign access tokens for this API.
    # One of: RS256, HS256, PS256
    signing_alg = optional(string)

    # allow_offline_access indicates whether refresh tokens can be issued.
    allow_offline_access = optional(bool, false)

    # token_lifetime is the duration (in seconds) that access tokens remain valid.
    # Range: 0 to 2592000 (30 days)
    token_lifetime = optional(number)

    # token_lifetime_for_web is the duration for tokens issued via implicit/hybrid flows.
    # Range: 0 to 2592000 (30 days)
    token_lifetime_for_web = optional(number)

    # skip_consent_for_verifiable_first_party_clients indicates whether to skip
    # the consent prompt for first-party applications.
    skip_consent_for_verifiable_first_party_clients = optional(bool, true)

    # enforce_policies enables RBAC authorization policies for this API.
    enforce_policies = optional(bool, false)

    # token_dialect determines the format of access tokens issued for this API.
    # One of: access_token, access_token_authz, rfc9068_profile, rfc9068_profile_authz
    token_dialect = optional(string)

    # scopes defines the permissions that can be granted for this API.
    scopes = optional(list(object({
      # name is the scope identifier used in OAuth flows.
      name = string
      # description is a human-readable explanation of what this scope grants.
      description = optional(string)
    })))
  })
}
