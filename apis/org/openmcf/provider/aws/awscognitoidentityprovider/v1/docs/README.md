# AWS Cognito Identity Provider — Technical Reference

## Federation Model

Amazon Cognito User Pools act as a federation hub that aggregates multiple external identity providers (IdPs) into a single user directory. Each identity provider resource configures a trust relationship between Cognito and one external IdP. Users authenticate via the external IdP; Cognito receives the IdP's assertion or tokens, maps attributes to the user pool schema, and issues Cognito tokens (ID, access, refresh) to the application.

Cognito supports three federation categories:

1. **Social providers** (Google, Facebook, Login with Amazon, Sign in with Apple): OAuth 2.0 flows with provider-specific endpoints and scopes.
2. **Generic OIDC** (Okta, Azure AD, Auth0, Keycloak): OpenID Connect discovery from `.well-known/openid-configuration`.
3. **SAML 2.0** (Enterprise IdPs): Metadata exchange, SP-initiated and IdP-initiated SSO.

The federation flow is always the same from the application's perspective: redirect to Cognito's hosted UI or `/oauth2/authorize`, optionally specify `identity_provider` or `idp_identifier` to skip the IdP picker, and receive Cognito tokens on callback. The complexity of OAuth, OIDC, or SAML is handled by Cognito.

---

## OAuth 2.0 Flow (Social Providers)

Social providers (Google, Facebook, Amazon, Apple) use the OAuth 2.0 Authorization Code flow:

1. User selects the social provider on the Cognito hosted UI.
2. Cognito redirects the user to the provider's authorization endpoint (e.g., `https://accounts.google.com/o/oauth2/v2/auth`) with `client_id`, `redirect_uri` (Cognito's callback), `response_type=code`, and `scope` from `authorize_scopes`.
3. User authenticates at the provider and grants consent.
4. Provider redirects back to Cognito with an authorization code.
5. Cognito exchanges the code for the provider's access token at the provider's token endpoint (POST).
6. Cognito calls the provider's userinfo/attributes endpoint (e.g., `https://people.googleapis.com/v1/people/me?personFields=`) with the access token to fetch user attributes.
7. Cognito applies attribute mapping, creates or updates the federated user, and issues Cognito tokens.

Provider-specific behavior:

- **Google**: Uses `https://accounts.google.com` as issuer; Cognito auto-discovers `authorize_url`, `token_url`, `attributes_url` from `oidc_issuer`. Scopes are space-separated (e.g., `email profile openid`).
- **Facebook**: Uses comma-separated scopes (e.g., `public_profile,email`). Optional `api_version` (e.g., `v17.0`) pins the Graph API version.
- **Login with Amazon**: Uses `https://api.amazon.com` for token and user profile. Scopes are space-separated.
- **Sign in with Apple**: Uses `client_secret` JWT signed with the Apple private key; no static client secret. `team_id`, `key_id`, `private_key` are required.

---

## OIDC Flow (Generic OIDC Providers)

For generic OIDC providers (Okta, Azure AD, Auth0, Keycloak), Cognito:

1. Fetches `https://{oidc_issuer}/.well-known/openid-configuration`.
2. Extracts `authorization_endpoint`, `token_endpoint`, `userinfo_endpoint`, `jwks_uri`.

When `oidc_issuer` is the only endpoint-related field provided, Cognito performs full auto-discovery. You can override individual endpoints with `authorize_url`, `token_url`, `attributes_url`, `jwks_uri` when the IdP's discovery document is incorrect or incomplete.

OIDC providers must support `client_secret_post` client authentication and HTTPS. If the IdP is a public client (PKCE-only), `client_secret` can be omitted.

**Userinfo method**: The `attributes_request_method` field controls whether Cognito uses GET or POST when calling the userinfo endpoint. Default is GET. Some IdPs (e.g., older Auth0) require POST.

**Scopes**: `authorize_scopes` is optional. When omitted, Cognito typically requests `openid email profile`. Specify explicitly for IdPs that require additional scopes.

---

## SAML 2.0 Flow

SAML federation works in two modes:

### SP-Initiated SSO

1. User visits the application and clicks "Sign in with SAML".
2. Application redirects to Cognito's `/oauth2/authorize` with `identity_provider` set to the provider name.
3. Cognito generates a SAML `<AuthnRequest>` and redirects the user to the IdP's SSO endpoint (from the IdP's metadata).
4. User authenticates at the IdP.
5. IdP POSTs a SAML `<Response>` to Cognito's Assertion Consumer Service (ACS) URL: `https://{domain}.auth.{region}.amazoncognito.com/saml2/idpresponse`.
6. Cognito validates the assertion, decrypts if `EncryptedResponses` is true, maps attributes, and issues Cognito tokens.

### IdP-Initiated SSO

When `idp_init` is true, the IdP can start the flow directly:

1. User visits the IdP's application portal and clicks a link to the app.
2. IdP POSTs a SAML `<Response>` to Cognito's ACS URL without a prior `<AuthnRequest>`.
3. Cognito validates the assertion and issues tokens.

### Metadata Exchange

SAML configuration requires either:

- **metadata_url**: URL to the IdP's metadata XML. Cognito fetches it periodically.
- **metadata_file**: Inline metadata XML content. Use when the IdP is not reachable from the internet or for static metadata.

The metadata must include:

- Entity ID
- SSO endpoint (HTTP-Redirect or HTTP-POST binding)
- X.509 certificate(s) for assertion signing
- Optional: SLO endpoint (if `idp_sign_out` is true)

---

## Attribute Mapping

Attribute mapping assigns IdP attribute names to Cognito user pool attributes. Keys are Cognito attribute names (e.g., `email`, `username`, `given_name`, `family_name`); values are IdP claim names or SAML attribute URIs.

### Default mappings

When `attribute_mapping` is omitted, Cognito applies provider-specific defaults:

| Provider | `username` source | Default mapped attributes |
|----------|-------------------|---------------------------|
| Google | `sub` | `email`, `email_verified`, `name`, `given_name`, `family_name`, `picture` |
| Facebook | `id` | `email`, `name`, `picture` |
| Login with Amazon | `user_id` | `email`, `name` |
| Sign in with Apple | `sub` | `email`, `name` (from first sign-in only) |
| OIDC | `sub` | Standard OIDC claims (`email`, `email_verified`, `name`, etc.) |
| SAML | `NameID` | None; required attributes must be mapped explicitly |

### Username derivation

Cognito derives the federated user's `username` from a fixed source per provider type and prepends the provider name: `{ProviderName}_{source_value}`. For example, `Google_123456789012345678901`. The source attribute cannot be overridden in the mapping; only custom attributes and standard attributes other than `username` can be mapped.

### Custom attributes

Mapped attributes must be writable by the app client. Custom attributes must be mutable. Immutable custom attributes cannot be updated on federated sign-in; Cognito returns an error if the IdP sends a value for a mapped immutable attribute.

### Mapped values

- Maximum value length: 2,048 characters.
- Multi-valued attributes: Cognito flattens to `[value1,value2,value3]` (comma-delimited, URL-encoded).
- Blank values: If the IdP sends a blank value for a mapped attribute, Cognito clears the attribute.

---

## Token Exchange

Cognito receives provider tokens and does not pass them through to the application. The flow:

1. **Social providers**: Cognito receives access token from the provider's token endpoint, uses it to call the userinfo endpoint, then discards it.
2. **OIDC**: Cognito receives ID token and access token. Validates the ID token via JWKS, uses access token for userinfo if needed. Maps claims to user pool attributes.
3. **SAML**: Cognito receives the SAML assertion, validates signature and conditions, extracts attributes from the assertion.

Cognito then issues its own tokens:

- **ID token**: Cognito-signed JWT with `sub`, `email`, `cognito:username`, and mapped attributes.
- **Access token**: OAuth scopes and group memberships.
- **Refresh token**: Opaque; used to obtain new ID/access tokens without re-authentication.

The application never sees the provider's tokens. Optionally, you can map `access_token` or `id_token` to a custom attribute (max 2,048 chars) if you need to pass them through to backend logic.

---

## Provider-Specific Details Keys

The AWS API expects a flat `ProviderDetails` map. Keys vary by provider type.

### Google

| Key | Required | Description |
|-----|----------|-------------|
| `client_id` | Yes | OAuth client ID from Google Cloud Console |
| `client_secret` | Yes | OAuth client secret |
| `authorize_scopes` | Yes | Space-separated scopes (e.g., `email profile openid`) |

Cognito auto-discovers: `authorize_url`, `token_url`, `attributes_url`, `oidc_issuer`. These appear in Describe response only.

### Facebook

| Key | Required | Description |
|-----|----------|-------------|
| `client_id` | Yes | Facebook App ID |
| `client_secret` | Yes | Facebook App Secret |
| `authorize_scopes` | Yes | Comma-separated scopes (e.g., `public_profile,email`) |
| `api_version` | No | Graph API version (e.g., `v17.0`). Omit for Cognito default |

### Login with Amazon

| Key | Required | Description |
|-----|----------|-------------|
| `client_id` | Yes | LWA client ID from Amazon Developer Console |
| `client_secret` | Yes | LWA client secret |
| `authorize_scopes` | Yes | Space-separated scopes (e.g., `profile postal_code`) |

### Sign in with Apple

| Key | Required | Description |
|-----|----------|-------------|
| `client_id` | Yes | Apple Services ID (bundle ID) |
| `team_id` | Yes | Apple Developer Team ID (10-char alphanumeric) |
| `key_id` | Yes | Key ID for the Apple private key |
| `private_key` | Yes | PEM-encoded private key (used to sign client_secret JWT) |
| `authorize_scopes` | Yes | Space-separated scopes (e.g., `email name`) |

### OIDC

| Key | Required | Description |
|-----|----------|-------------|
| `client_id` | Yes | OIDC client ID |
| `oidc_issuer` | Yes | Issuer URL (e.g., `https://login.microsoftonline.com/{tenant}/v2.0`). Used for discovery |
| `authorize_scopes` | No | Space-separated scopes (default: `openid email profile`) |
| `client_secret` | No | Required for confidential clients; omit for public clients |
| `attributes_request_method` | No | `GET` or `POST` for userinfo. Default: `GET` |
| `authorize_url` | No | Override auto-discovered authorization endpoint |
| `token_url` | No | Override auto-discovered token endpoint |
| `attributes_url` | No | Override auto-discovered userinfo endpoint |
| `jwks_uri` | No | Override auto-discovered JWKS endpoint |

### SAML

| Key | Required | Description |
|-----|----------|-------------|
| `MetadataFile` | One of | Inline SAML metadata XML (quotes escaped) |
| `MetadataURL` | One of | URL to IdP metadata document |
| `IDPSignout` | No | `"true"` to enable single logout |
| `IDPInit` | No | `"true"` to enable IdP-initiated SSO |
| `EncryptedResponses` | No | `"true"` to require encrypted SAML assertions |
| `RequestSigningAlgorithm` | No | Algorithm for signing AuthnRequest (e.g., `rsa-sha256`) |

---

## ForceNew Behavior

The following fields are immutable in the AWS API. Changing any of them requires destroying and recreating the identity provider:

| Field | Reason |
|-------|--------|
| `user_pool_id` | The provider is scoped to a single pool. Moving it would require a new resource. |
| `provider_name` | The provider name is the primary identifier; it's used in `supported_identity_providers` and in URLs. Renaming would break existing references. |
| `provider_type` | Provider type determines the schema of `ProviderDetails`. Changing from Google to OIDC would require a different configuration structure. |

---

## ActiveEncryptionCertificate (SAML Quirk)

When you configure a SAML provider with `EncryptedResponses: true`, Cognito must publish its encryption certificate in the SP metadata so the IdP can encrypt assertions. AWS generates this certificate and stores it in the identity provider configuration.

On `DescribeIdentityProvider`, the response includes `ActiveEncryptionCertificate` in `ProviderDetails`. This field is **read-only** and **auto-populated by AWS**. You cannot set it in `CreateIdentityProvider` or `UpdateIdentityProvider`. Terraform and Pulumi typically ignore it in state to avoid drift, since the value is managed by AWS.

If you use `MetadataFile` with inline metadata, the IdP must obtain the SP metadata from Cognito (e.g., `https://{domain}.auth.{region}.amazoncognito.com/saml2/metadata`) to get the encryption certificate. Configure the IdP to use that metadata URL for encryption, or ensure your IdP supports importing the certificate from Cognito's SP metadata.

---

## Coordination Pattern: Identity Providers and User Pool Clients

Identity providers and user pool clients are separate resources:

1. **AwsCognitoIdentityProvider** creates the IdP configuration attached to a user pool.
2. **AwsCognitoUserPool** (or equivalent) defines app clients with `supported_identity_providers`.

To enable federated sign-in for a client, add the IdP's `provider_name` to the client's `supported_identity_providers` list. For example:

```yaml
# User Pool Client
spec:
  clients:
    - name: web-app
      supportedIdentityProviders:
        - COGNITO
        - Google
        - CorpSSO
```

`COGNITO` enables native username/password sign-in. `Google` and `CorpSSO` are the `provider_name` values from the corresponding AwsCognitoIdentityProvider resources.

The identity provider must exist before it can be referenced. Deployment order: User Pool → Identity Providers → User Pool Clients (or ensure clients are updated after IdPs are created).

---

## Security Considerations

### Client secret handling

- **Storage**: Client secrets are stored in Cognito and returned in API responses. Do not log or expose them. Use secrets management (e.g., AWS Secrets Manager, environment variables) when provisioning.
- **Rotation**: Rotating a provider's client secret requires updating the identity provider. Plan for zero-downtime rotation if the IdP supports it.

### SAML assertion encryption

- When `EncryptedResponses` is true, the IdP must encrypt the SAML assertion with Cognito's public key. Obtain the key from Cognito's SP metadata.
- If the IdP does not support encryption, leave `EncryptedResponses` false.

### SAML request signing

- `RequestSigningAlgorithm` (e.g., `rsa-sha256`) signs the `<AuthnRequest>` sent to the IdP. Some enterprise IdPs require signed requests.
- Cognito uses its own signing key; the IdP must trust Cognito's SP metadata certificate.

### IdP identifiers

- `idp_identifiers` are alternative names for the provider. Use them in the `idp_identifier` query parameter of the authorize endpoint to redirect to this provider without exposing the real `provider_name`.
- Domain format identifiers (e.g., `example.com`) enable email-address matching for SAML: users with `user@example.com` can be routed to the matching IdP automatically.
