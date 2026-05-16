# Auth0Connection Variables
# This file defines all input variables for the Auth0Connection Terraform module
# These variables map to the Auth0ConnectionSpec protobuf message

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
  description = "Auth0Connection specification"
  type = object({
    # strategy is the identity provider strategy/type for this connection.
    # Required. One of: auth0, google-oauth2, facebook, github, linkedin, twitter,
    # microsoft-account, apple, samlp, oidc, waad, ad, adfs
    strategy = string

    # display_name is the human-readable name shown in the Auth0 Universal Login page.
    display_name = optional(string)

    # enabled_clients is a list of Auth0 application client IDs that can use this connection.
    # StringValueOrRef fields are flattened to plain strings by the tfvars generator.
    enabled_clients = optional(list(string))

    # is_domain_connection indicates whether this connection can be used for identifier-first flows.
    is_domain_connection = optional(bool, false)

    # realms are the identifiers that can be used for this connection in authentication requests.
    realms = optional(list(string))

    # show_as_button controls whether this connection shows as a button on the Universal Login page.
    show_as_button = optional(bool, true)

    # metadata is a map of custom metadata key-value pairs to store with the connection.
    metadata = optional(map(string))

    # database_options configures database connection behavior (strategy: auth0)
    database_options = optional(object({
      # password_policy: none, low, fair, good, excellent
      password_policy = optional(string, "good")

      # requires_username determines if users must provide a username in addition to email.
      requires_username = optional(bool, false)

      # disable_signup prevents new user signups through this connection.
      disable_signup = optional(bool, false)

      # brute_force_protection enables protection against brute force login attacks.
      brute_force_protection = optional(bool, true)

      # password_history_size is the number of previous passwords to check against (0-24).
      password_history_size = optional(number, 5)

      # password_no_personal_info prevents passwords containing user's personal information.
      password_no_personal_info = optional(bool, true)

      # password_dictionary enables checking passwords against a dictionary of common passwords.
      password_dictionary = optional(bool, true)

      # mfa_enabled enables Multi-Factor Authentication for this connection.
      mfa_enabled = optional(bool, false)
    }))

    # social_options configures social identity provider connections
    social_options = optional(object({
      # client_id is the OAuth client ID from the social provider.
      client_id = string

      # client_secret is the OAuth client secret from the social provider.
      client_secret = string

      # scopes is a list of OAuth scopes to request from the social provider.
      scopes = optional(list(string))

      # allowed_audiences restricts which audiences can use this connection.
      allowed_audiences = optional(list(string))

      # upstream_params are custom parameters to pass to the upstream social provider.
      upstream_params = optional(map(string))
    }))

    # saml_options configures SAML enterprise connections (strategy: samlp)
    saml_options = optional(object({
      # sign_in_endpoint is the SAML Identity Provider's Single Sign-On URL.
      sign_in_endpoint = string

      # signing_cert is the X.509 signing certificate from the Identity Provider.
      signing_cert = string

      # sign_out_endpoint is the SAML Identity Provider's Single Logout URL.
      sign_out_endpoint = optional(string)

      # entity_id is the unique identifier for the Identity Provider.
      entity_id = optional(string)

      # protocol_binding specifies how SAML requests should be sent.
      protocol_binding = optional(string)

      # user_id_attribute is the SAML attribute to use as the user identifier.
      user_id_attribute = optional(string)

      # sign_request indicates whether Auth0 should sign SAML requests.
      sign_request = optional(bool, false)

      # signature_algorithm is the algorithm used for signing SAML requests.
      signature_algorithm = optional(string, "rsa-sha256")

      # digest_algorithm is the algorithm used for digest in SAML signatures.
      digest_algorithm = optional(string, "sha256")

      # attribute_mappings maps SAML attributes to Auth0 user profile fields.
      attribute_mappings = optional(map(string))
    }))

    # oidc_options configures OpenID Connect enterprise connections (strategy: oidc)
    oidc_options = optional(object({
      # issuer is the OIDC issuer URL (the "iss" claim value).
      issuer = string

      # client_id is the OAuth client ID from the OIDC provider.
      client_id = string

      # client_secret is the OAuth client secret from the OIDC provider.
      client_secret = optional(string)

      # scopes is a list of OIDC scopes to request.
      scopes = optional(list(string))

      # type specifies the OIDC flow type (front_channel or back_channel).
      type = optional(string, "front_channel")

      # authorization_endpoint overrides the authorization endpoint from discovery.
      authorization_endpoint = optional(string)

      # token_endpoint overrides the token endpoint from discovery.
      token_endpoint = optional(string)

      # userinfo_endpoint overrides the userinfo endpoint from discovery.
      userinfo_endpoint = optional(string)

      # jwks_uri overrides the JWKS URI from discovery.
      jwks_uri = optional(string)

      # attribute_mappings maps OIDC claims to Auth0 user profile fields.
      attribute_mappings = optional(map(string))
    }))

    # azure_ad_options configures Azure AD/Entra ID enterprise connections (strategy: waad)
    azure_ad_options = optional(object({
      # client_id is the Application (client) ID from Azure AD app registration.
      client_id = string

      # client_secret is the client secret from Azure AD app registration.
      client_secret = string

      # domain is the Azure AD tenant domain.
      domain = string

      # tenant_id is the Azure AD tenant ID (Directory ID).
      tenant_id = optional(string)

      # use_common_endpoint allows users from any Azure AD tenant.
      use_common_endpoint = optional(bool, false)

      # max_groups_to_retrieve limits the number of groups retrieved from Azure AD.
      max_groups_to_retrieve = optional(number, 50)

      # should_trust_email_verified indicates whether to trust Azure AD's email_verified claim.
      should_trust_email_verified = optional(bool, true)

      # api_enable_users enables the ability to retrieve users from the Azure AD directory.
      api_enable_users = optional(bool, false)
    }))
  })
}

