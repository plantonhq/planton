# AWS Cognito Identity Provider

Deploys an external identity provider federated into an Amazon Cognito User Pool. Users can sign in through social providers (Google, Facebook, Amazon, Apple), enterprise OIDC (Okta, Azure AD, Auth0), or SAML 2.0 federation. Cognito maps the provider's user attributes to the pool's schema and issues tokens for your application.

## What Gets Created

- **Cognito Identity Provider** — a federated IdP configuration attached to an existing User Pool. The IdP enables users to authenticate via the external provider; Cognito handles token exchange, attribute mapping, and user linking.

## Supported Provider Types

| Provider Type | Use Case | Configuration Message |
|---------------|----------|------------------------|
| `Google` | Google OAuth 2.0 | `google` |
| `Facebook` | Facebook Login | `facebook` |
| `LoginWithAmazon` | Login with Amazon | `loginWithAmazon` |
| `SignInWithApple` | Sign in with Apple | `signInWithApple` |
| `OIDC` | Generic OpenID Connect (Okta, Azure AD, Auth0) | `oidc` |
| `SAML` | SAML 2.0 federation (Azure AD, Salesforce, ADFS) | `saml` |

The `provider_type` field is a proto enum that matches AWS API values exactly. Each provider type has its own typed configuration message — there is no flat map; the schema is strongly typed for each IdP.

## Child Resource Relationship

**This is a child resource of `AwsCognitoUserPool`.** The User Pool must exist before creating identity providers. You cannot create an IdP without a pool.

After creating an identity provider, add its `provider_name` to the User Pool Client's `supported_identity_providers` list to enable federated sign-in for that client.

## ForceNew Fields

All identity fields are **ForceNew** — changing any of them destroys and recreates the identity provider:

| Field | ForceNew | Notes |
|-------|----------|-------|
| `userPoolId` | Yes | Provider cannot be moved between pools |
| `providerName` | Yes | Must be unique within a User Pool |
| `providerType` | Yes | Changing type requires replacement |

## Attribute Mapping

Attribute mapping is **optional**. When omitted, AWS applies default mappings based on the provider type. When specified, keys are Cognito User Pool attribute names (e.g., `email`, `username`, `given_name`); values are provider-specific attribute names or paths (e.g., `sub`, `email`, or SAML claim URIs).

## Spec Fields Reference

| Field | Type | Required | ForceNew | Description |
|-------|------|----------|----------|-------------|
| `userPoolId` | `StringValueOrRef` | Yes | Yes | User Pool ID (format: `{region}_{poolId}`). Can reference `AwsCognitoUserPool` via `valueFrom`. |
| `providerName` | `string` | Yes | Yes | Display name (1–32 chars). Unique within pool. Referenced in User Pool Client's `supported_identity_providers`. |
| `providerType` | `enum` | Yes | Yes | One of: `Google`, `Facebook`, `LoginWithAmazon`, `SignInWithApple`, `OIDC`, `SAML`. |
| `google` | `AwsCognitoIdpGoogleConfig` | When `providerType: Google` | — | Google OAuth config. |
| `facebook` | `AwsCognitoIdpFacebookConfig` | When `providerType: Facebook` | — | Facebook Login config. |
| `loginWithAmazon` | `AwsCognitoIdpLoginWithAmazonConfig` | When `providerType: LoginWithAmazon` | — | Login with Amazon config. |
| `signInWithApple` | `AwsCognitoIdpSignInWithAppleConfig` | When `providerType: SignInWithApple` | — | Sign in with Apple config. |
| `oidc` | `AwsCognitoIdpOidcConfig` | When `providerType: OIDC` | — | Generic OIDC config. |
| `saml` | `AwsCognitoIdpSamlConfig` | When `providerType: SAML` | — | SAML 2.0 config. |
| `attributeMapping` | `map<string, string>` | No | No | Maps provider attributes to Cognito attributes. Optional; AWS defaults apply when omitted. |
| `idpIdentifiers` | `string[]` | No | No | Alternative identifiers (max 50, 1–40 chars each) for `idp_identifier` in login endpoint. |

### Provider Configuration Details

**Google** (`google`):

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `clientId` | `string` | Yes | OAuth client ID from Google Cloud Console |
| `clientSecret` | `string` | Yes | OAuth client secret |
| `authorizeScopes` | `string` | Yes | Space-separated scopes (e.g., `"email profile openid"`) |

**Facebook** (`facebook`):

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `clientId` | `string` | Yes | App ID from Facebook Developer Portal |
| `clientSecret` | `string` | Yes | App Secret |
| `authorizeScopes` | `string` | Yes | Comma-separated scopes (e.g., `"email,public_profile"`) |
| `apiVersion` | `string` | No | Graph API version (e.g., `"v17.0"`). Omit for Cognito default. |

**Sign in with Apple** (`signInWithApple`):

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `clientId` | `string` | Yes | Apple Services ID |
| `teamId` | `string` | Yes | Apple Developer Team ID (10-char alphanumeric) |
| `keyId` | `string` | Yes | Key ID for the Apple private key |
| `privateKey` | `string` | Yes | Apple private key in PEM format |
| `authorizeScopes` | `string` | Yes | Space-separated scopes (e.g., `"email name"`) |

**OIDC** (`oidc`):

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `clientId` | `string` | Yes | OIDC client ID |
| `oidcIssuer` | `string` | Yes | Issuer URL. Cognito auto-discovers endpoints from `.well-known/openid-configuration`. |
| `authorizeScopes` | `string` | No | Space-separated scopes (e.g., `"openid email profile"`) |
| `clientSecret` | `string` | No | Optional for public clients using PKCE |
| `attributesRequestMethod` | `string` | No | `"GET"` or `"POST"` for userinfo. Default: `"GET"`. |
| `authorizeUrl` | `string` | No | Override auto-discovered authorization URL |
| `tokenUrl` | `string` | No | Override auto-discovered token URL |
| `attributesUrl` | `string` | No | Override auto-discovered userinfo URL |
| `jwksUri` | `string` | No | Override auto-discovered JWKS URL |

**SAML** (`saml`):

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `metadataFile` | `string` | One of metadataFile/metadataUrl | Inline SAML metadata XML |
| `metadataUrl` | `string` | One of metadataFile/metadataUrl | URL to IdP's SAML metadata document |
| `idpSignOut` | `bool` | No | Enable single logout (SLO) |
| `idpInit` | `bool` | No | Enable IdP-initiated SSO |
| `encryptedResponses` | `bool` | No | Require encrypted SAML assertions |
| `requestSigningAlgorithm` | `string` | No | SAML request signing algorithm (e.g., `"rsa-sha256"`) |

## Stack Outputs

| Output | Type | Description |
|--------|------|-------------|
| `provider_name` | `string` | The name registered in the User Pool. Add this to User Pool Client's `supported_identity_providers` to enable federated sign-in. |
| `provider_type` | `string` | The provider type (e.g., `"Google"`, `"OIDC"`, `"SAML"`). Informational. |

## Prerequisites

- **AwsCognitoUserPool** must exist before creating identity providers
- AWS credentials with permissions for `cognito-idp:*`
- OAuth/OIDC/SAML credentials from the external IdP (client IDs, secrets, metadata URLs, etc.)

## Related Components

- [AwsCognitoUserPool](../awscognitouserpool/v1/README.md) — parent resource; must be created first
- [AwsCognitoUserPoolClient](../awscognitouserpool/v1/README.md) — add `provider_name` to `supported_identity_providers` to enable federated sign-in
