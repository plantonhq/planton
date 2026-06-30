locals {
  resource_name = coalesce(try(var.metadata.name, null), "awscognitoidentityprovider")

  provider_type = try(var.spec.provider_type, var.spec.providerType, "")

  # Build provider_details based on which typed config is present.
  # OAuth providers use snake_case keys, SAML uses PascalCase keys.
  google_details = try(var.spec.google, null) != null ? {
    client_id        = var.spec.google.clientId
    client_secret    = var.spec.google.clientSecret
    authorize_scopes = var.spec.google.authorizeScopes
  } : null

  facebook_details = try(var.spec.facebook, null) != null ? merge({
    client_id        = var.spec.facebook.clientId
    client_secret    = var.spec.facebook.clientSecret
    authorize_scopes = var.spec.facebook.authorizeScopes
  }, try(var.spec.facebook.apiVersion, "") != "" ? {
    api_version = var.spec.facebook.apiVersion
  } : {}) : null

  login_with_amazon_details = try(var.spec.loginWithAmazon, null) != null ? {
    client_id        = var.spec.loginWithAmazon.clientId
    client_secret    = var.spec.loginWithAmazon.clientSecret
    authorize_scopes = var.spec.loginWithAmazon.authorizeScopes
  } : null

  sign_in_with_apple_details = try(var.spec.signInWithApple, null) != null ? {
    client_id        = var.spec.signInWithApple.clientId
    team_id          = var.spec.signInWithApple.teamId
    key_id           = var.spec.signInWithApple.keyId
    private_key      = var.spec.signInWithApple.privateKey
    authorize_scopes = var.spec.signInWithApple.authorizeScopes
  } : null

  oidc_details = try(var.spec.oidc, null) != null ? merge(
    {
      client_id   = var.spec.oidc.clientId
      oidc_issuer = var.spec.oidc.oidcIssuer
    },
    try(var.spec.oidc.authorizeScopes, "") != "" ? { authorize_scopes = var.spec.oidc.authorizeScopes } : {},
    try(var.spec.oidc.clientSecret, "") != "" ? { client_secret = var.spec.oidc.clientSecret } : {},
    try(var.spec.oidc.attributesRequestMethod, "") != "" ? { attributes_request_method = var.spec.oidc.attributesRequestMethod } : {},
    try(var.spec.oidc.authorizeUrl, "") != "" ? { authorize_url = var.spec.oidc.authorizeUrl } : {},
    try(var.spec.oidc.tokenUrl, "") != "" ? { token_url = var.spec.oidc.tokenUrl } : {},
    try(var.spec.oidc.attributesUrl, "") != "" ? { attributes_url = var.spec.oidc.attributesUrl } : {},
    try(var.spec.oidc.jwksUri, "") != "" ? { jwks_uri = var.spec.oidc.jwksUri } : {},
  ) : null

  saml_details = try(var.spec.saml, null) != null ? merge(
    try(var.spec.saml.metadataFile, "") != "" ? { MetadataFile = var.spec.saml.metadataFile } : {},
    try(var.spec.saml.metadataUrl, "") != "" ? { MetadataURL = var.spec.saml.metadataUrl } : {},
    try(var.spec.saml.idpSignOut, false) ? { IDPSignout = "true" } : {},
    try(var.spec.saml.idpInit, false) ? { IDPInit = "true" } : {},
    try(var.spec.saml.encryptedResponses, false) ? { EncryptedResponses = "true" } : {},
    try(var.spec.saml.requestSigningAlgorithm, "") != "" ? { RequestSigningAlgorithm = var.spec.saml.requestSigningAlgorithm } : {},
  ) : null

  # Select the correct provider_details based on which config is non-null.
  provider_details = coalesce(
    local.google_details,
    local.facebook_details,
    local.login_with_amazon_details,
    local.sign_in_with_apple_details,
    local.oidc_details,
    local.saml_details,
    {}
  )

  attribute_mapping = try(var.spec.attributeMapping, null) != null ? var.spec.attributeMapping : null

  idp_identifiers = try(var.spec.idpIdentifiers, null) != null ? var.spec.idpIdentifiers : null
}
