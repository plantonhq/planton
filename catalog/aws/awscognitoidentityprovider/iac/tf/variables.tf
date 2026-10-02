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
  description = "AwsCognitoIdentityProvider specification"
  type = object({
    # The AWS region where the resource will be created.
    # Example: "us-west-2", "eu-west-1"
    region = string

    # The ID of the Cognito User Pool to attach this identity provider to.
    # Format: "{region}_{poolId}" (e.g., "us-east-1_Ab1Cd2EfG").
    #
    # This field is ForceNew: changing it requires replacing the identity provider.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    user_pool_id = string

    # Display name for this identity provider within the User Pool. Must be
    # unique within the pool. Referenced by User Pool Clients in their
    # `supported_identity_providers` list.
    #
    # For the SOCIAL provider types AWS requires the name to EQUAL the type
    # ("Google", "Facebook", "LoginWithAmazon", "SignInWithApple") -- custom
    # names are an OIDC/SAML feature.
    #
    # Examples: "Google", "CorpOkta", "AzureAD-SAML"
    #
    # 1-32 UTF-8 characters. This field is ForceNew.
    provider_name = string

    # The type of identity provider, exactly as the AWS API spells it. Valid
    # values: "Google", "Facebook", "LoginWithAmazon", "SignInWithApple",
    # "OIDC", "SAML". Determines which provider configuration field to populate
    # (google, facebook, login_with_amazon, sign_in_with_apple, oidc, or saml).
    #
    # This field is ForceNew: changing it requires replacing the identity provider.
    provider_type = string

    # Google OAuth 2.0 configuration. Set when provider_type is Google.
    google = optional(object({
      # Google OAuth 2.0 client ID from the Google Cloud Console.
      client_id = string

      # Google OAuth 2.0 client secret from the Google Cloud Console.
      client_secret = string

      # Space-separated OAuth scopes (e.g., "email profile openid").
      authorize_scopes = string
    }))

    # Facebook Login configuration. Set when provider_type is Facebook.
    facebook = optional(object({
      # Facebook App ID from the Facebook Developer Portal.
      client_id = string

      # Facebook App Secret from the Facebook Developer Portal.
      client_secret = string

      # Comma-separated OAuth scopes (e.g., "email,public_profile").
      # Note: Facebook uses comma-separated scopes, unlike other providers
      # which use space-separated.
      authorize_scopes = string

      # Facebook Graph API version (e.g., "v17.0"). When omitted, Cognito uses
      # the latest version it supports.
      api_version = optional(string, "")
    }))

    # Login with Amazon configuration. Set when provider_type is LoginWithAmazon.
    login_with_amazon = optional(object({
      # Login with Amazon client ID from the Amazon Developer Console.
      client_id = string

      # Login with Amazon client secret from the Amazon Developer Console.
      client_secret = string

      # Space-separated OAuth scopes (e.g., "profile postal_code").
      authorize_scopes = string
    }))

    # Sign in with Apple configuration. Set when provider_type is SignInWithApple.
    sign_in_with_apple = optional(object({
      # Apple Services ID configured in the Apple Developer Portal.
      client_id = string

      # Apple Developer Team ID (10-character alphanumeric string).
      team_id = string

      # Key ID for the Apple private key.
      key_id = string

      # Apple private key in PEM format. Used to generate the client_secret JWT.
      private_key = string

      # Space-separated OAuth scopes (e.g., "email name").
      authorize_scopes = string
    }))

    # Generic OIDC configuration. Set when provider_type is OIDC.
    oidc = optional(object({
      # OIDC client ID registered with the identity provider.
      client_id = string

      # OIDC issuer URL (e.g., "https://login.microsoftonline.com/{tenant}/v2.0").
      # Cognito auto-discovers authorize, token, userinfo, and JWKS endpoints from
      # the issuer's .well-known/openid-configuration document.
      oidc_issuer = string

      # Space-separated OAuth/OIDC scopes (e.g., "openid email profile").
      authorize_scopes = optional(string, "")

      # OIDC client secret. Optional for public clients that use PKCE.
      client_secret = optional(string, "")

      # HTTP method for the userinfo endpoint: "GET" or "POST".
      # When omitted, Cognito defaults to "GET".
      attributes_request_method = optional(string, "")

      # Override the auto-discovered authorization endpoint URL.
      authorize_url = optional(string, "")

      # Override the auto-discovered token endpoint URL.
      token_url = optional(string, "")

      # Override the auto-discovered userinfo endpoint URL.
      attributes_url = optional(string, "")

      # Override the auto-discovered JWKS endpoint URL.
      jwks_uri = optional(string, "")

      # When true, Cognito appends the requested attributes as query parameters
      # to the userinfo (attributes_url) request instead of relying on the
      # provider returning them by default. Only some providers need this.
      attributes_url_add_attributes = optional(bool, false)
    }))

    # SAML 2.0 configuration. Set when provider_type is SAML.
    saml = optional(object({
      # SAML metadata XML content as a string. Use this for inline metadata.
      # Mutually exclusive with metadata_url.
      metadata_file = optional(string, "")

      # URL pointing to the IdP's SAML metadata document. Cognito fetches the
      # metadata from this URL. Mutually exclusive with metadata_file.
      metadata_url = optional(string, "")

      # Enable single logout (SLO). When true, Cognito signs the user out of
      # the SAML IdP when they sign out of the User Pool.
      idp_sign_out = optional(bool, false)

      # Enable IdP-initiated SSO. When true, the IdP can start the sign-in
      # flow directly without the user first visiting the Cognito hosted UI.
      idp_init = optional(bool, false)

      # Require encrypted SAML assertions from the IdP.
      encrypted_responses = optional(bool, false)

      # SAML request signing algorithm (e.g., "rsa-sha256"). When omitted,
      # Cognito uses the default algorithm.
      request_signing_algorithm = optional(string, "")
    }))

    # Maps identity provider attributes to Cognito User Pool attributes.
    # Keys are Cognito user pool attribute names (e.g., "email", "username",
    # "given_name"). Values are provider-specific attribute names or paths
    # (e.g., "sub", "email").
    #
    # When omitted, AWS applies default mappings based on the provider type.
    attribute_mapping = optional(map(string), {})

    # Alternative identifiers for this identity provider. These can be used
    # in the login endpoint's `idp_identifier` parameter to redirect to this
    # provider without exposing the provider_name.
    #
    # Maximum 50 identifiers, each 1-40 characters.
    idp_identifiers = optional(list(string), [])
  })
}
