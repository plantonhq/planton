# Auth0Connection

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `auth0.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

Auth0ConnectionSpec defines the configuration for an Auth0 Connection.
A connection in Auth0 represents an identity provider - either a user database,
social identity provider (Google, Facebook, GitHub), or enterprise identity provider (SAML, OIDC, LDAP).

This spec focuses on the 80/20 use case: configuring common connection types
with sensible defaults while exposing the essential configuration options.

Supported connection strategies:
- Database: auth0, Username-Password-Authentication
- Social: google-oauth2, facebook, github, linkedin, twitter, microsoft-account
- Enterprise: samlp, oidc, ad, adfs, waad

## Example

```yaml
# Test Manifest for Auth0Connection
# This manifest is used for local development and testing of the IaC modules.
# Do NOT use this manifest for production deployments.

apiVersion: auth0.planton.dev/v1alpha1
kind: Auth0Connection
metadata:
  name: test-user-database
  org: test-org
  env: development
  labels:
    purpose: testing
    component: auth0connection
spec:
  strategy: auth0
  display_name: Test Database Connection
  enabled_clients:
    - value: "test-client-id-001"
    - value: "test-client-id-002"
  show_as_button: true
  database_options:
    password_policy: good
    requires_username: false
    disable_signup: false
    brute_force_protection: true
    # The three advanced password options below require a paid Auth0
    # "password-advanced-options" entitlement (the test tenant must have it).
    password_history_size: 5
    password_no_personal_info: true
    password_dictionary: true
    mfa_enabled: false
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.strategy` | `string` | yes |  |  |
| `spec.displayName` | `string` |  |  |  |
| `spec.enabledClients` | `[]string \| valueFrom` |  |  | Auth0Client (`status.outputs.client_id`) |
| `spec.isDomainConnection` | `bool` |  |  |  |
| `spec.realms` | `[]string` |  |  |  |
| `spec.showAsButton` | `bool` |  |  |  |
| `spec.metadata` | `map<string, string>` |  |  |  |
| `spec.databaseOptions` | `Auth0DatabaseOptions` |  |  |  |
| `spec.databaseOptions.passwordPolicy` | `string` |  |  |  |
| `spec.databaseOptions.requiresUsername` | `bool` |  |  |  |
| `spec.databaseOptions.disableSignup` | `bool` |  |  |  |
| `spec.databaseOptions.bruteForceProtection` | `bool` |  |  |  |
| `spec.databaseOptions.passwordHistorySize` | `int32` |  |  |  |
| `spec.databaseOptions.passwordNoPersonalInfo` | `bool` |  |  |  |
| `spec.databaseOptions.passwordDictionary` | `bool` |  |  |  |
| `spec.databaseOptions.mfaEnabled` | `bool` |  |  |  |
| `spec.socialOptions` | `Auth0SocialOptions` |  |  |  |
| `spec.socialOptions.clientId` | `string` | yes |  |  |
| `spec.socialOptions.clientSecret` | `string` (sensitive) | yes |  |  |
| `spec.socialOptions.scopes` | `[]string` |  |  |  |
| `spec.socialOptions.allowedAudiences` | `[]string` |  |  |  |
| `spec.socialOptions.upstreamParams` | `map<string, string>` |  |  |  |
| `spec.samlOptions` | `Auth0SamlOptions` |  |  |  |
| `spec.samlOptions.signInEndpoint` | `string` | yes |  |  |
| `spec.samlOptions.signingCert` | `string` | yes |  |  |
| `spec.samlOptions.signOutEndpoint` | `string` |  |  |  |
| `spec.samlOptions.entityId` | `string` |  |  |  |
| `spec.samlOptions.protocolBinding` | `string` |  |  |  |
| `spec.samlOptions.userIdAttribute` | `string` |  |  |  |
| `spec.samlOptions.signRequest` | `bool` |  |  |  |
| `spec.samlOptions.signatureAlgorithm` | `string` |  |  |  |
| `spec.samlOptions.digestAlgorithm` | `string` |  |  |  |
| `spec.samlOptions.attributeMappings` | `map<string, string>` |  |  |  |
| `spec.oidcOptions` | `Auth0OidcOptions` |  |  |  |
| `spec.oidcOptions.issuer` | `string` | yes |  |  |
| `spec.oidcOptions.clientId` | `string` | yes |  |  |
| `spec.oidcOptions.clientSecret` | `string` (sensitive) |  |  |  |
| `spec.oidcOptions.scopes` | `[]string` |  |  |  |
| `spec.oidcOptions.type` | `string` |  |  |  |
| `spec.oidcOptions.authorizationEndpoint` | `string` |  |  |  |
| `spec.oidcOptions.tokenEndpoint` | `string` |  |  |  |
| `spec.oidcOptions.userinfoEndpoint` | `string` |  |  |  |
| `spec.oidcOptions.jwksUri` | `string` |  |  |  |
| `spec.oidcOptions.attributeMappings` | `map<string, string>` |  |  |  |
| `spec.azureAdOptions` | `Auth0AzureAdOptions` |  |  |  |
| `spec.azureAdOptions.clientId` | `string` | yes |  |  |
| `spec.azureAdOptions.clientSecret` | `string` (sensitive) | yes |  |  |
| `spec.azureAdOptions.domain` | `string` | yes |  |  |
| `spec.azureAdOptions.tenantId` | `string` |  |  |  |
| `spec.azureAdOptions.useCommonEndpoint` | `bool` |  |  |  |
| `spec.azureAdOptions.maxGroupsToRetrieve` | `int32` |  |  |  |
| `spec.azureAdOptions.shouldTrustEmailVerified` | `bool` |  |  |  |
| `spec.azureAdOptions.apiEnableUsers` | `bool` |  |  |  |

## Field Details

### spec.strategy

`string` · required

strategy is the identity provider strategy/type for this connection.
This determines how users authenticate and what additional configuration is required.

Database strategies:
- "auth0": Auth0's hosted database (default, recommended for most use cases)

Social strategies:
- "google-oauth2": Google OAuth 2.0
- "facebook": Facebook Login
- "github": GitHub OAuth
- "linkedin": LinkedIn OAuth
- "twitter": Twitter OAuth
- "microsoft-account": Microsoft Account (personal)
- "apple": Sign in with Apple

Enterprise strategies:
- "samlp": SAML Protocol
- "oidc": OpenID Connect
- "waad": Windows Azure Active Directory (Entra ID)
- "ad": Active Directory/LDAP
- "adfs": Active Directory Federation Services

https://registry.terraform.io/providers/auth0/auth0/latest/docs/resources/connection#strategy

- rule: {"required":true,"string":{"in":["auth0","google-oauth2","facebook","github","linkedin","twitter","microsoft-account","apple","samlp","oidc","waad","ad","adfs"]}}

### spec.displayName

`string`

display_name is the human-readable name shown in the Auth0 Universal Login page.
This is the name users will see when choosing how to log in.
If not specified, Auth0 will use a default name based on the strategy.
Example: "Google", "Company SSO", "Sign up with Email"

### spec.enabledClients

`[]string | valueFrom`

enabled_clients is a list of Auth0 application client IDs that can use this connection.
Only applications in this list will show this connection as a login option.
If empty, no applications will be able to use this connection.

You can provide either:
- Direct client ID: {value: "abc123clientID"}
- Reference to Auth0Client component: {value_from: {kind: Auth0Client, name: "my-app"}}

When using references, the client_id is automatically resolved from the
Auth0Client's status.outputs.client_id field.

- references: Auth0Client (`status.outputs.client_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: Auth0Client, name: <that resource's name>, fieldPath: status.outputs.client_id}} -- a bare string does not parse

### spec.isDomainConnection

`bool`

is_domain_connection indicates whether this connection can be used for identifier-first
authentication flows. When true, Auth0 will attempt to discover the appropriate
connection based on the user's email domain.
Default: false

### spec.realms

`[]string`

realms are the identifiers that can be used for this connection in authentication requests.
This is useful for routing users to the correct connection.
If not specified, defaults to the connection name.

### spec.showAsButton

`bool`

show_as_button controls whether this connection shows as a button on the Universal Login page.
When false, the connection will only be used if explicitly requested or discovered via domain.
Default: true

### spec.metadata

`map<string, string>`

metadata is a map of custom metadata key-value pairs to store with the connection.
This can be used to store integration-specific identifiers or configuration.
Maximum 10 key-value pairs, with keys and values up to 255 characters each.

### spec.databaseOptions

`Auth0DatabaseOptions`

database_options configures database connection behavior.
Only applicable when strategy is "auth0" (Auth0 database).

### spec.databaseOptions.passwordPolicy

`string`

password_policy defines the password complexity requirements.
- "none": No password requirements
- "low": At least 6 characters
- "fair": At least 8 characters, including lower and upper case
- "good": At least 8 characters, including lower, upper, numeric
- "excellent": At least 10 characters, including lower, upper, numeric, special
Default: "good"

- rule: {"string":{"in":["","none","low","fair","good","excellent"]}}

### spec.databaseOptions.requiresUsername

`bool`

requires_username determines if users must provide a username in addition to email.
When true, users sign up with both username and email.
Default: false

### spec.databaseOptions.disableSignup

`bool`

disable_signup prevents new user signups through this connection.
Useful when you only want existing users to log in, or when onboarding is done programmatically.
Default: false

### spec.databaseOptions.bruteForceProtection

`bool`

brute_force_protection enables protection against brute force login attacks.
When enabled, Auth0 will block login attempts after multiple failures.
Default: true (recommended)

### spec.databaseOptions.passwordHistorySize

`int32`

password_history_size is the number of previous passwords to check against.
Users cannot reuse passwords from their history. Set to 0 to disable.
Valid range: 0-24
Requires Auth0's paid "password-advanced-options" entitlement; leaving it
unset avoids a 403 on free/lower-tier tenants.
Default: 0 (disabled)

- rule: {"int32":{"lte":24,"gte":0}}

### spec.databaseOptions.passwordNoPersonalInfo

`bool`

password_no_personal_info prevents passwords containing user's personal information
(name, username, email). Recommended for security.
Requires Auth0's paid "password-advanced-options" entitlement; leaving it
unset avoids a 403 on free/lower-tier tenants.
Default: false (disabled)

### spec.databaseOptions.passwordDictionary

`bool`

password_dictionary enables checking passwords against a dictionary of common passwords.
When true, users cannot use common/weak passwords.
Requires Auth0's paid "password-advanced-options" entitlement; leaving it
unset avoids a 403 on free/lower-tier tenants.
Default: false (disabled)

### spec.databaseOptions.mfaEnabled

`bool`

mfa_enabled enables Multi-Factor Authentication for this connection.
When true, users will be prompted for a second factor during login.
Default: false

### spec.socialOptions

`Auth0SocialOptions`

social_options configures social identity provider connections.
Only applicable when strategy is a social provider (google-oauth2, facebook, github, etc.).

### spec.socialOptions.clientId

`string` · required

client_id is the OAuth client ID from the social provider.
This is obtained from the provider's developer console.
Required for all social connections.

- rule: {"required":true}

### spec.socialOptions.clientSecret

`string` · required · sensitive

client_secret is the OAuth client secret from the social provider.
This is obtained from the provider's developer console.
Required for all social connections.

- rule: {"required":true}

### spec.socialOptions.scopes

`[]string`

scopes is a list of OAuth scopes to request from the social provider.
These determine what user information Auth0 can access.
If not specified, default scopes for the strategy will be used.
Example for Google: ["openid", "profile", "email"]

### spec.socialOptions.allowedAudiences

`[]string`

allowed_audiences restricts which audiences (applications) can use this connection.
Only applicable for providers that support audience restrictions.

### spec.socialOptions.upstreamParams

`map<string, string>`

upstream_params are custom parameters to pass to the upstream social provider.
Useful for provider-specific features like login hints or prompt types.
Example: {"login_hint": "user@example.com", "prompt": "select_account"}

### spec.samlOptions

`Auth0SamlOptions`

saml_options configures SAML enterprise connections.
Only applicable when strategy is "samlp".

### spec.samlOptions.signInEndpoint

`string` · required

sign_in_endpoint is the SAML Identity Provider's Single Sign-On URL.
This is where Auth0 sends SAML authentication requests.
Also known as "SSO URL" or "Login URL".
Required for SAML connections.

- rule: {"required":true}

### spec.samlOptions.signingCert

`string` · required

signing_cert is the X.509 signing certificate from the Identity Provider.
Used to verify the signature on SAML responses.
Should be in PEM format (including BEGIN/END CERTIFICATE headers).
Required for SAML connections.

- rule: {"required":true}

### spec.samlOptions.signOutEndpoint

`string`

sign_out_endpoint is the SAML Identity Provider's Single Logout URL.
This is where Auth0 sends logout requests.
Optional but recommended for complete logout functionality.

### spec.samlOptions.entityId

`string`

entity_id is the unique identifier for the Identity Provider.
Also known as "Issuer" or "EntityID".
If not specified, Auth0 will attempt to derive it from the metadata.

### spec.samlOptions.protocolBinding

`string`

protocol_binding specifies how SAML requests should be sent.
- "urn:oasis:names:tc:SAML:2.0:bindings:HTTP-Redirect": URL-encoded in query string
- "urn:oasis:names:tc:SAML:2.0:bindings:HTTP-POST": Form POST
Default: HTTP-Redirect

- rule: {"string":{"in":["","urn:oasis:names:tc:SAML:2.0:bindings:HTTP-Redirect","urn:oasis:names:tc:SAML:2.0:bindings:HTTP-POST"]}}

### spec.samlOptions.userIdAttribute

`string`

user_id_attribute is the SAML attribute to use as the user identifier.
Default: "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/nameidentifier"

### spec.samlOptions.signRequest

`bool`

sign_request indicates whether Auth0 should sign SAML requests.
Required by some Identity Providers.
Default: false

### spec.samlOptions.signatureAlgorithm

`string`

signature_algorithm is the algorithm used for signing SAML requests.
- "rsa-sha256": RSA with SHA-256 (recommended)
- "rsa-sha1": RSA with SHA-1 (deprecated, use only for legacy systems)
Default: rsa-sha256

- rule: {"string":{"in":["","rsa-sha256","rsa-sha1"]}}

### spec.samlOptions.digestAlgorithm

`string`

digest_algorithm is the algorithm used for digest in SAML signatures.
- "sha256": SHA-256 (recommended)
- "sha1": SHA-1 (deprecated, use only for legacy systems)
Default: sha256

- rule: {"string":{"in":["","sha256","sha1"]}}

### spec.samlOptions.attributeMappings

`map<string, string>`

attribute_mappings maps SAML attributes to Auth0 user profile fields.
Keys are Auth0 profile fields (email, name, given_name, family_name, nickname, picture).
Values are SAML attribute names from the IdP response.

### spec.oidcOptions

`Auth0OidcOptions`

oidc_options configures OpenID Connect enterprise connections.
Only applicable when strategy is "oidc".

### spec.oidcOptions.issuer

`string` · required

issuer is the OIDC issuer URL (the "iss" claim value).
Auth0 will use OIDC Discovery to fetch configuration from /.well-known/openid-configuration
Required for OIDC connections.

- rule: {"required":true}

### spec.oidcOptions.clientId

`string` · required

client_id is the OAuth client ID from the OIDC provider.
Obtained from the provider's application registration.
Required for OIDC connections.

- rule: {"required":true}

### spec.oidcOptions.clientSecret

`string` · sensitive

client_secret is the OAuth client secret from the OIDC provider.
Required for OIDC connections using authorization code flow.

### spec.oidcOptions.scopes

`[]string`

scopes is a list of OIDC scopes to request.
"openid" is always requested implicitly.
Common scopes: "profile", "email", "address", "phone"
Default: ["openid", "profile", "email"]

### spec.oidcOptions.type

`string`

type specifies the OIDC flow type.
- "front_channel": Authorization Code Flow (recommended for web apps)
- "back_channel": Uses token endpoint directly
Default: front_channel

- rule: {"string":{"in":["","front_channel","back_channel"]}}

### spec.oidcOptions.authorizationEndpoint

`string`

authorization_endpoint overrides the authorization endpoint from discovery.
Only set this if the provider's discovery document is incorrect.

### spec.oidcOptions.tokenEndpoint

`string`

token_endpoint overrides the token endpoint from discovery.
Only set this if the provider's discovery document is incorrect.

### spec.oidcOptions.userinfoEndpoint

`string`

userinfo_endpoint overrides the userinfo endpoint from discovery.
Only set this if the provider's discovery document is incorrect.

### spec.oidcOptions.jwksUri

`string`

jwks_uri overrides the JWKS URI from discovery.
Only set this if the provider's discovery document is incorrect.

### spec.oidcOptions.attributeMappings

`map<string, string>`

attribute_mappings maps OIDC claims to Auth0 user profile fields.
Keys are Auth0 profile fields (email, name, given_name, family_name, nickname, picture).
Values are OIDC claim names from the ID token or userinfo response.

### spec.azureAdOptions

`Auth0AzureAdOptions`

azure_ad_options configures Azure AD/Entra ID enterprise connections.
Only applicable when strategy is "waad".

### spec.azureAdOptions.clientId

`string` · required

client_id is the Application (client) ID from Azure AD app registration.
Found in Azure Portal > Azure Active Directory > App Registrations > Your App > Overview
Required for Azure AD connections.

- rule: {"required":true}

### spec.azureAdOptions.clientSecret

`string` · required · sensitive

client_secret is the client secret from Azure AD app registration.
Found in Azure Portal > Azure Active Directory > App Registrations > Your App > Certificates & secrets
Required for Azure AD connections.

- rule: {"required":true}

### spec.azureAdOptions.domain

`string` · required

domain is the Azure AD tenant domain.
This can be the primary domain (e.g., "contoso.onmicrosoft.com")
or a custom domain (e.g., "contoso.com").
Required for Azure AD connections.

- rule: {"required":true}

### spec.azureAdOptions.tenantId

`string`

tenant_id is the Azure AD tenant ID (Directory ID).
Found in Azure Portal > Azure Active Directory > Overview
If not specified, "common" is used, allowing any Azure AD tenant.

### spec.azureAdOptions.useCommonEndpoint

`bool`

use_common_endpoint determines whether to use the common endpoint.
When true, allows users from any Azure AD tenant (multi-tenant app).
When false, restricts to the specified tenant.
Default: false

### spec.azureAdOptions.maxGroupsToRetrieve

`int32`

max_groups_to_retrieve limits the number of groups retrieved from Azure AD.
Azure AD can return many groups; set this to limit the groups in the user profile.
Set to 0 for no limit.
Default: 50

- rule: {"int32":{"gte":0}}

### spec.azureAdOptions.shouldTrustEmailVerified

`bool`

should_trust_email_verified indicates whether to trust Azure AD's email_verified claim.
When true, Auth0 trusts that Azure AD has verified the user's email.
Default: true

### spec.azureAdOptions.apiEnableUsers

`bool`

api_enable_users enables the ability to retrieve users from the Azure AD directory.
When true, you can use Auth0's Management API to list users from Azure AD.
Requires additional Azure AD permissions (Directory.Read.All).
Default: false

## Outputs

Reference an output from another manifest as `valueFrom: {kind: Auth0Connection, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.id` | `string` | id is the unique identifier of the Auth0 connection. This is the primary identifier used in Auth0 APIs and Management Dashboard. Format: "con_" followed by a unique string (e.g., "con_0000000000000001"). |
| `status.outputs.name` | `string` | name is the unique name of the connection within the Auth0 tenant. This name is used in authentication URLs and API calls. It is automatically generated from metadata.name or can be customized. |
| `status.outputs.strategy` | `string` | strategy is the identity provider strategy type. Reflects the connection type configured in the spec. Examples: "auth0", "google-oauth2", "samlp", "oidc", "waad" |
| `status.outputs.is_enabled` | `string` | is_enabled indicates whether the connection is currently enabled. A disabled connection cannot be used for authentication. |
| `status.outputs.provisioning_ticket_url` | `string` | provisioning_ticket_url is the URL used for self-service connection setup. Only available for certain enterprise connections (like SAML or OIDC). Users can visit this URL to complete connection configuration. |
| `status.outputs.callback_url` | `string` | callback_url is the Auth0 callback URL for this connection. For social and enterprise connections, this URL must be registered with the identity provider. Format: https://{tenant}.auth0.com/login/callback |
| `status.outputs.metadata_url` | `string` | metadata_url is the SAML metadata URL for this connection. Only available for SAML connections. Identity Providers can use this to configure the Service Provider. Format: https://{tenant}.auth0.com/samlp/metadata/{connection_name} |
| `status.outputs.entity_id` | `string` | entity_id is the SAML Service Provider Entity ID. Only available for SAML connections. This is the unique identifier Auth0 uses as a SAML SP. Format: urn:auth0:{tenant}:{connection_name} |
| `status.outputs.enabled_client_ids` | `[]string` | enabled_client_ids is the list of Auth0 application client IDs that can use this connection. |
| `status.outputs.realms` | `[]string` | realms is the list of realms/domains associated with this connection. Used for identifier-first authentication flows. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.enabledClients` | Auth0Client | `status.outputs.client_id` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| Auth0Client | `spec.enabledConnections` | `status.outputs.name` |
| Auth0User | `spec.connectionName` | `status.outputs.name` |

## See Also

- [Overview](../README.md)
