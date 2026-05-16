# Auth0Client Variables
# This file defines all input variables for the Auth0Client Terraform module
# These variables map to the Auth0ClientSpec protobuf message

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
  description = "Auth0Client specification"
  type = object({
    # application_type defines the type of application being registered.
    # Required. One of: native, spa, regular_web, non_interactive
    application_type = string

    # description is an optional free-text description of the application.
    description = optional(string)

    # logo_uri is the URL of the application's logo.
    logo_uri = optional(string)

    # callbacks are the allowed callback URLs for the application.
    callbacks = optional(list(string))

    # allowed_logout_urls are URLs that Auth0 can redirect to after logout.
    allowed_logout_urls = optional(list(string))

    # web_origins are the allowed origins for web message response mode.
    web_origins = optional(list(string))

    # allowed_origins are CORS origins allowed for this application.
    allowed_origins = optional(list(string))

    # grant_types specifies which OAuth grant types this application can use.
    grant_types = optional(list(string))

    # oidc_conformant enables stricter OIDC-conformant behavior.
    oidc_conformant = optional(bool, true)

    # is_first_party indicates whether this is a first-party application.
    is_first_party = optional(bool, true)

    # cross_origin_authentication enables cross-origin authentication.
    cross_origin_authentication = optional(bool, false)

    # cross_origin_loc is the URL for cross-origin verification fallback.
    cross_origin_loc = optional(string)

    # sso enables Single Sign-On for this application.
    sso = optional(bool, true)

    # sso_disabled explicitly disables SSO for this application.
    sso_disabled = optional(bool, false)

    # custom_login_page is the custom HTML for the login page.
    custom_login_page = optional(string)

    # custom_login_page_on enables the custom login page.
    custom_login_page_on = optional(bool, false)

    # initiate_login_uri is the URL to initiate login (for OIDC third-party apps).
    initiate_login_uri = optional(string)

    # organization_usage determines how organizations are used with this app.
    # One of: deny, allow, require
    organization_usage = optional(string)

    # organization_require_behavior specifies when org is required.
    # One of: no_prompt, pre_login_prompt, post_login_prompt
    organization_require_behavior = optional(string)

    # jwt_configuration contains settings for JWT tokens issued to this client.
    jwt_configuration = optional(object({
      # lifetime_in_seconds is the expiration time for JWTs in seconds.
      lifetime_in_seconds = optional(number)

      # scopes is a map of custom scopes and their descriptions.
      scopes = optional(map(string))

      # alg is the algorithm used to sign the JWT.
      # One of: HS256, RS256, PS256
      alg = optional(string)

      # secret_encoded indicates if the client secret is base64 encoded.
      secret_encoded = optional(bool, false)
    }))

    # refresh_token contains settings for refresh token behavior.
    refresh_token = optional(object({
      # rotation_type determines refresh token rotation behavior.
      # One of: non-rotating, rotating
      rotation_type = optional(string)

      # expiration_type determines how refresh tokens expire.
      # One of: non-expiring, expiring
      expiration_type = optional(string)

      # token_lifetime is the absolute lifetime of a refresh token in seconds.
      token_lifetime = optional(number)

      # idle_token_lifetime is the inactivity timeout for refresh tokens.
      idle_token_lifetime = optional(number)

      # infinite_token_lifetime allows tokens to never expire.
      infinite_token_lifetime = optional(bool, false)

      # infinite_idle_token_lifetime allows tokens to never expire due to inactivity.
      infinite_idle_token_lifetime = optional(bool, false)

      # leeway is the clock skew leeway in seconds for token validation.
      leeway = optional(number)
    }))

    # native_social_login configures native social login for mobile apps.
    native_social_login = optional(object({
      apple = optional(object({
        enabled = optional(bool, false)
      }))
      facebook = optional(object({
        enabled = optional(bool, false)
      }))
    }))

    # mobile configures mobile-specific settings.
    mobile = optional(object({
      android = optional(object({
        app_package_name         = optional(string)
        sha256_cert_fingerprints = optional(list(string))
      }))
      ios = optional(object({
        team_id               = optional(string)
        app_bundle_identifier = optional(string)
      }))
    }))

    # client_metadata is a map of custom metadata key-value pairs.
    client_metadata = optional(map(string))

    # client_aliases are alternative identifiers for this client.
    client_aliases = optional(list(string))

    # is_token_endpoint_ip_header_trusted determines if IP header is trusted.
    is_token_endpoint_ip_header_trusted = optional(bool, false)

    # oidc_backchannel_logout configures OIDC back-channel logout.
    oidc_backchannel_logout = optional(object({
      backchannel_logout_urls = optional(list(string))
    }))

    # enabled_connections limits which connections this app can use.
    # StringValueOrRef fields are flattened to plain strings by the tfvars generator.
    enabled_connections = optional(list(string))

    # api_grants configures which APIs this client is authorized to access.
    # For M2M applications using client_credentials grant, at least one API grant is typically required.
    # Each entry creates an auth0_client_grant resource linking this client to an API.
    api_grants = optional(list(object({
      # audience is the API identifier the client is authorized to access.
      # Required. For Auth0 Management API: "https://{tenant}.{region}.auth0.com/api/v2/"
      # StringValueOrRef fields are flattened to plain strings by the tfvars generator.
      audience = string

      # scopes are the permissions granted for this API.
      scopes = optional(list(string))

      # allow_any_organization determines if any organization can be used with this grant.
      allow_any_organization = optional(bool, false)

      # organization_usage defines whether organizations can be used with client credentials exchanges.
      # One of: deny, allow, require
      organization_usage = optional(string)
    })))
  })
}


