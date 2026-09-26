# Auth0Client

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `auth0.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

Auth0ClientSpec defines the configuration for an Auth0 Application (Client).
In Auth0, "Applications" represent the clients that interact with Auth0 for authentication.
Applications are the entry point for users to authenticate and authorize access to APIs.

This spec covers the 80/20 use case for configuring Auth0 applications:
- Web applications (SPAs, traditional server-rendered apps)
- Mobile/native applications
- Machine-to-Machine (M2M) applications for API access

Supported application types:
- native: Mobile, desktop, or CLI applications
- spa: Single Page Applications (JavaScript running in browser)
- regular_web: Traditional server-side web applications
- non_interactive: Machine-to-Machine (M2M) API clients

https://auth0.com/docs/get-started/applications

## Example

```yaml
# Test Manifest for Auth0Client
# This manifest is used for local development and testing of the IaC modules.
# Do NOT use this manifest for production deployments.

apiVersion: auth0.planton.dev/v1alpha1
kind: Auth0Client
metadata:
  name: test-spa-app
  org: test-org
  env: development
  labels:
    purpose: testing
    component: auth0client
spec:
  application_type: spa
  description: Test SPA Application for Development
  callbacks:
    - http://localhost:3000/callback
    - http://localhost:4200/callback
  allowed_logout_urls:
    - http://localhost:3000
    - http://localhost:4200
  web_origins:
    - http://localhost:3000
    - http://localhost:4200
  grant_types:
    - authorization_code
    - refresh_token
  oidc_conformant: true
  is_first_party: true
  jwt_configuration:
    lifetime_in_seconds: 36000
    alg: RS256
  refresh_token:
    rotation_type: rotating
    expiration_type: expiring
    token_lifetime: 2592000
    idle_token_lifetime: 1296000
    leeway: 60
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.applicationType` | `string` | yes |  |  |
| `spec.description` | `string` |  |  |  |
| `spec.logoUri` | `string` |  |  |  |
| `spec.callbacks` | `[]string` |  |  |  |
| `spec.allowedLogoutUrls` | `[]string` |  |  |  |
| `spec.webOrigins` | `[]string` |  |  |  |
| `spec.allowedOrigins` | `[]string` |  |  |  |
| `spec.grantTypes` | `[]string` |  |  |  |
| `spec.oidcConformant` | `bool` |  |  |  |
| `spec.isFirstParty` | `bool` |  |  |  |
| `spec.crossOriginAuthentication` | `bool` |  |  |  |
| `spec.crossOriginLoc` | `string` |  |  |  |
| `spec.sso` | `bool` |  |  |  |
| `spec.ssoDisabled` | `bool` |  |  |  |
| `spec.customLoginPage` | `string` |  |  |  |
| `spec.customLoginPageOn` | `bool` |  |  |  |
| `spec.initiateLoginUri` | `string` |  |  |  |
| `spec.organizationUsage` | `string` |  |  |  |
| `spec.organizationRequireBehavior` | `string` |  |  |  |
| `spec.jwtConfiguration` | `Auth0JwtConfiguration` |  |  |  |
| `spec.jwtConfiguration.lifetimeInSeconds` | `int32` |  |  |  |
| `spec.jwtConfiguration.scopes` | `map<string, string>` |  |  |  |
| `spec.jwtConfiguration.alg` | `string` |  |  |  |
| `spec.jwtConfiguration.secretEncoded` | `bool` |  |  |  |
| `spec.refreshToken` | `Auth0RefreshTokenConfiguration` |  |  |  |
| `spec.refreshToken.rotationType` | `string` |  |  |  |
| `spec.refreshToken.expirationType` | `string` |  |  |  |
| `spec.refreshToken.tokenLifetime` | `int32` |  |  |  |
| `spec.refreshToken.idleTokenLifetime` | `int32` |  |  |  |
| `spec.refreshToken.infiniteTokenLifetime` | `bool` |  |  |  |
| `spec.refreshToken.infiniteIdleTokenLifetime` | `bool` |  |  |  |
| `spec.refreshToken.leeway` | `int32` |  |  |  |
| `spec.nativeSocialLogin` | `Auth0NativeSocialLogin` |  |  |  |
| `spec.nativeSocialLogin.apple` | `Auth0NativeSocialLoginProvider` |  |  |  |
| `spec.nativeSocialLogin.apple.enabled` | `bool` |  |  |  |
| `spec.nativeSocialLogin.facebook` | `Auth0NativeSocialLoginProvider` |  |  |  |
| `spec.nativeSocialLogin.facebook.enabled` | `bool` |  |  |  |
| `spec.mobile` | `Auth0MobileConfiguration` |  |  |  |
| `spec.mobile.android` | `Auth0MobileAndroid` |  |  |  |
| `spec.mobile.android.appPackageName` | `string` |  |  |  |
| `spec.mobile.android.sha256CertFingerprints` | `[]string` |  |  |  |
| `spec.mobile.ios` | `Auth0MobileIos` |  |  |  |
| `spec.mobile.ios.teamId` | `string` |  |  |  |
| `spec.mobile.ios.appBundleIdentifier` | `string` |  |  |  |
| `spec.clientMetadata` | `map<string, string>` |  |  |  |
| `spec.clientAliases` | `[]string` |  |  |  |
| `spec.isTokenEndpointIpHeaderTrusted` | `bool` |  |  |  |
| `spec.oidcBackchannelLogout` | `Auth0OidcBackchannelLogout` |  |  |  |
| `spec.oidcBackchannelLogout.backchannelLogoutUrls` | `[]string` |  |  |  |
| `spec.enabledConnections` | `[]string \| valueFrom` |  |  | Auth0Connection (`status.outputs.name`) |
| `spec.apiGrants` | `[]Auth0ClientApiGrant` |  |  |  |
| `spec.apiGrants[].audience` | `string \| valueFrom` | yes |  | Auth0ResourceServer (`status.outputs.identifier`) |
| `spec.apiGrants[].scopes` | `[]string` |  |  |  |
| `spec.apiGrants[].allowAnyOrganization` | `bool` |  |  |  |
| `spec.apiGrants[].organizationUsage` | `string` |  |  |  |

## Field Details

### spec.applicationType

`string` · required

application_type defines the type of application being registered.
This determines the appropriate OAuth flows and security settings.

- "native": Mobile, desktop, or CLI applications that run natively on a device.
  Uses Authorization Code with PKCE. Cannot securely store secrets.
- "spa": Single Page Applications running JavaScript in the browser.
  Uses Authorization Code with PKCE. Cannot securely store secrets.
- "regular_web": Traditional web apps with server-side rendering.
  Can use Authorization Code flow and securely store client secrets.
- "non_interactive": Machine-to-Machine applications for API access.
  Uses Client Credentials flow. Designed for backend services.

https://auth0.com/docs/get-started/applications/application-types

- rule: {"required":true,"string":{"in":["native","spa","regular_web","non_interactive"]}}

### spec.description

`string`

description is an optional free-text description of the application.
Useful for documenting the purpose, owner, or other metadata.
Maximum 140 characters.

- rule: {"string":{"maxLen":"140"}}

### spec.logoUri

`string`

logo_uri is the URL of the application's logo.
This is displayed on the consent page and login page.
Must be a valid HTTPS URL to an image file.

### spec.callbacks

`[]string`

callbacks are the allowed callback URLs for the application.
After authentication, Auth0 will only redirect to URLs in this list.
For SPAs and native apps, include your development and production URLs.
Example: ["https://myapp.com/callback", "http://localhost:3000/callback"]

### spec.allowedLogoutUrls

`[]string`

allowed_logout_urls are URLs that Auth0 can redirect to after logout.
Must be registered here to be used with the logout endpoint.
Example: ["https://myapp.com", "http://localhost:3000"]

### spec.webOrigins

`[]string`

web_origins are the allowed origins for web message response mode.
Required for SPAs using popup or iframe-based authentication.
Example: ["https://myapp.com", "http://localhost:3000"]

### spec.allowedOrigins

`[]string`

allowed_origins are CORS origins allowed for this application.
Used for cross-origin requests from JavaScript applications.
Example: ["https://myapp.com"]

### spec.grantTypes

`[]string`

grant_types specifies which OAuth grant types this application can use.
If not specified, defaults are based on application_type.

Common grant types:
- "authorization_code": Standard OAuth 2.0 authorization code flow
- "implicit": Implicit flow (legacy, not recommended)
- "refresh_token": Allows obtaining refresh tokens
- "client_credentials": For M2M applications
- "password": Resource owner password grant (not recommended)
- "http://auth0.com/oauth/grant-type/password-realm": Password with realm
- "urn:ietf:params:oauth:grant-type:device_code": Device authorization flow

### spec.oidcConformant

`bool`

oidc_conformant enables stricter OIDC-conformant behavior.
When true, the application will follow OIDC specification strictly.
Recommended for new applications.
Default: true

### spec.isFirstParty

`bool`

is_first_party indicates whether this is a first-party application.
First-party apps are owned by the same entity as the Auth0 tenant.
They skip the consent prompt for users.
Default: true

### spec.crossOriginAuthentication

`bool`

cross_origin_authentication enables cross-origin authentication.
Required for embedded login forms in SPAs.
Only enable if you understand the security implications.
Default: false

### spec.crossOriginLoc

`string`

cross_origin_loc is the URL for cross-origin verification fallback.
Used with cross-origin authentication for certain browsers.

### spec.sso

`bool`

sso enables Single Sign-On for this application.
When enabled, users who are already logged in won't need to re-authenticate.
Default: true

### spec.ssoDisabled

`bool`

sso_disabled explicitly disables SSO for this application.
Set to true to require authentication for each session.
Default: false

### spec.customLoginPage

`string`

custom_login_page is the custom HTML for the login page.
Allows complete customization of the login experience.
Only used when custom_login_page_on is true.

### spec.customLoginPageOn

`bool`

custom_login_page_on enables the custom login page.
When true, uses custom_login_page instead of Universal Login.
Default: false

### spec.initiateLoginUri

`string`

initiate_login_uri is the URL to initiate login (for OIDC third-party apps).
Used for third-party initiated login flows.

### spec.organizationUsage

`string`

organization_usage determines how organizations are used with this app.
- "deny": Organizations cannot be used (default for most apps)
- "allow": Organizations can be used optionally
- "require": Organizations must be specified at login

- rule: {"string":{"in":["","deny","allow","require"]}}

### spec.organizationRequireBehavior

`string`

organization_require_behavior specifies when org is required.
- "no_prompt": Fail silently if no organization specified
- "pre_login_prompt": Show organization picker before login
- "post_login_prompt": Show organization picker after login
Only used when organization_usage is "require".

- rule: {"string":{"in":["","no_prompt","pre_login_prompt","post_login_prompt"]}}

### spec.jwtConfiguration

`Auth0JwtConfiguration`

jwt_configuration contains settings for JWT tokens issued to this client.

### spec.jwtConfiguration.lifetimeInSeconds

`int32`

lifetime_in_seconds is the expiration time for JWTs in seconds.
Default: 36000 (10 hours)
Range: 0-2592000 (30 days max)

- rule: {"int32":{"lte":2592000,"gte":0}}

### spec.jwtConfiguration.scopes

`map<string, string>`

scopes is a map of custom scopes and their descriptions.
These scopes will be available for this application.
Key: scope name, Value: scope description

### spec.jwtConfiguration.alg

`string`

alg is the algorithm used to sign the JWT.
- "HS256": HMAC using SHA-256 (symmetric, uses client secret)
- "RS256": RSA using SHA-256 (asymmetric, uses tenant keys)
- "PS256": RSA-PSS using SHA-256
Default: RS256 (recommended)

- rule: {"string":{"in":["","HS256","RS256","PS256"]}}

### spec.jwtConfiguration.secretEncoded

`bool`

secret_encoded indicates if the client secret is base64 encoded.
Only relevant when alg is HS256.
Default: false

### spec.refreshToken

`Auth0RefreshTokenConfiguration`

refresh_token contains settings for refresh token behavior.

### spec.refreshToken.rotationType

`string`

rotation_type determines refresh token rotation behavior.
- "non-rotating": Refresh tokens don't rotate (legacy behavior)
- "rotating": New refresh token issued with each use (recommended)
Default: non-rotating

- rule: {"string":{"in":["","non-rotating","rotating"]}}

### spec.refreshToken.expirationType

`string`

expiration_type determines how refresh tokens expire.
- "non-expiring": Tokens don't expire (not recommended)
- "expiring": Tokens expire based on configured lifetimes
Default: non-expiring

- rule: {"string":{"in":["","non-expiring","expiring"]}}

### spec.refreshToken.tokenLifetime

`int32`

token_lifetime is the absolute lifetime of a refresh token in seconds.
The token will expire after this time regardless of activity.
Only used when expiration_type is "expiring".
Default: 2592000 (30 days)

- rule: {"int32":{"gte":0}}

### spec.refreshToken.idleTokenLifetime

`int32`

idle_token_lifetime is the inactivity timeout for refresh tokens in seconds.
Token expires if not used within this time.
Only used when expiration_type is "expiring".
Default: 1296000 (15 days)

- rule: {"int32":{"gte":0}}

### spec.refreshToken.infiniteTokenLifetime

`bool`

infinite_token_lifetime allows tokens to never expire.
Only valid when expiration_type is "non-expiring".
Not recommended for security reasons.
Default: false

### spec.refreshToken.infiniteIdleTokenLifetime

`bool`

infinite_idle_token_lifetime allows tokens to never expire due to inactivity.
Only valid when expiration_type is "non-expiring".
Not recommended for security reasons.
Default: false

### spec.refreshToken.leeway

`int32`

leeway is the clock skew leeway in seconds for token validation.
Allows for slight differences in server clocks.
Default: 0

- rule: {"int32":{"gte":0}}

### spec.nativeSocialLogin

`Auth0NativeSocialLogin`

native_social_login configures native social login for mobile apps.
Only applicable for native application types.

### spec.nativeSocialLogin.apple

`Auth0NativeSocialLoginProvider`

apple configures Sign in with Apple native integration.

### spec.nativeSocialLogin.apple.enabled

`bool`

enabled determines if this native social provider is active.

### spec.nativeSocialLogin.facebook

`Auth0NativeSocialLoginProvider`

facebook configures Facebook native login integration.

### spec.nativeSocialLogin.facebook.enabled

`bool`

enabled determines if this native social provider is active.

### spec.mobile

`Auth0MobileConfiguration`

mobile configures mobile-specific settings.
Only applicable for native application types.

### spec.mobile.android

`Auth0MobileAndroid`

android configures Android-specific settings.

### spec.mobile.android.appPackageName

`string`

app_package_name is the Android application package name.
Example: "com.example.myapp"

### spec.mobile.android.sha256CertFingerprints

`[]string`

sha256_cert_fingerprints are the SHA-256 fingerprints of signing certificates.
Used for App Links and secure deep linking.
Example: ["D8:A0:..."]

### spec.mobile.ios

`Auth0MobileIos`

ios configures iOS-specific settings.

### spec.mobile.ios.teamId

`string`

team_id is the Apple Developer Team ID.
Required for universal links.

### spec.mobile.ios.appBundleIdentifier

`string`

app_bundle_identifier is the iOS application bundle identifier.
Example: "com.example.myapp"

### spec.clientMetadata

`map<string, string>`

client_metadata is a map of custom metadata key-value pairs.
Useful for storing application-specific configuration.
Maximum 10 key-value pairs.

### spec.clientAliases

`[]string`

client_aliases are alternative identifiers for this client.
Can be used in authentication requests instead of client_id.

### spec.isTokenEndpointIpHeaderTrusted

`bool`

is_token_endpoint_ip_header_trusted determines if IP header is trusted.
When true, Auth0 uses X-Forwarded-For header for IP-based features.
Default: false

### spec.oidcBackchannelLogout

`Auth0OidcBackchannelLogout`

oidc_backchannel_logout configures OIDC back-channel logout.

### spec.oidcBackchannelLogout.backchannelLogoutUrls

`[]string`

backchannel_logout_urls are the URLs to receive logout tokens.
Auth0 will POST a logout token to these URLs on logout.

### spec.enabledConnections

`[]string | valueFrom`

enabled_connections limits which connections this app can use.
If empty, all connections are available.

You can provide either:
- Direct connection name: {value: "Username-Password-Authentication"}
- Reference to Auth0Connection component: {value_from: {kind: Auth0Connection, name: "my-connection"}}

When using references, the connection name is automatically resolved from the
Auth0Connection's status.outputs.name field.

- references: Auth0Connection (`status.outputs.name`)
- rule: write as {value: <literal>} or {valueFrom: {kind: Auth0Connection, name: <that resource's name>, fieldPath: status.outputs.name}} -- a bare string does not parse

### spec.apiGrants

`[]Auth0ClientApiGrant`

api_grants configures which APIs this client is authorized to access.
For M2M applications using client_credentials grant, at least one API grant is typically required.
Each entry creates an auth0_client_grant resource linking this client to an API.

Without api_grants, an M2M application can authenticate but cannot access any APIs.
This is different from grant_types which only defines which OAuth flows are allowed.

Example - Authorize for Auth0 Management API:
  api_grants:
    - audience: "https://your-tenant.us.auth0.com/api/v2/"
      scopes:
        - read:users
        - read:user_idp_tokens

Example - Authorize for custom API:
  api_grants:
    - audience: "https://api.example.com/"
      scopes:
        - read:resources
        - write:resources

### spec.apiGrants[].audience

`string | valueFrom` · required

audience is the API identifier (Resource Server identifier) the client is authorized to access.
For Auth0 Management API: "https://{tenant}.{region}.auth0.com/api/v2/"
For custom APIs: the identifier you configured when creating the API in Auth0

You can provide either:
- Direct value: {value: "https://api.example.com/"}
- Reference to Auth0ResourceServer component: {value_from: {kind: Auth0ResourceServer, name: "my-api"}}

When using references, the audience is automatically resolved from the
Auth0ResourceServer's status.outputs.identifier field.

Example: "https://api.example.com/", "api.planton.live"

- references: Auth0ResourceServer (`status.outputs.identifier`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: Auth0ResourceServer, name: <that resource's name>, fieldPath: status.outputs.identifier}} -- a bare string does not parse

### spec.apiGrants[].scopes

`[]string`

scopes are the permissions granted for this API.
For Management API, common scopes include:
  - read:users, read:user_idp_tokens, create:users, update:users, delete:users
  - read:clients, update:clients, delete:clients
  - read:connections, update:connections
For custom APIs: the scopes you defined when creating the API in Auth0.
If empty, the client gets access to the API with no specific scopes
(valid for APIs that don't use scope-based authorization).

### spec.apiGrants[].allowAnyOrganization

`bool`

allow_any_organization determines if any organization can be used with this grant.
If false (default), the grant must be explicitly assigned to desired organizations.
Only relevant when using Auth0 Organizations feature.

### spec.apiGrants[].organizationUsage

`string`

organization_usage defines whether organizations can be used with client credentials
exchanges for this grant.
- "deny": Organizations cannot be used (default)
- "allow": Organizations can be used optionally
- "require": Organizations must be specified

- rule: {"string":{"in":["","deny","allow","require"]}}

## Outputs

Reference an output from another manifest as `valueFrom: {kind: Auth0Client, name: <resource-name>, fieldPath: status.outputs.<output>}`. A sensitive output is a secret the resource generates: on Planton it is kept in the organization's secret store and the output holds a `$secret/` reference, so feed it only to a sensitive field.

| Output | Type | Description |
|---|---|---|
| `status.outputs.id` | `string` | id is the unique identifier of the Auth0 client. This is used internally by Auth0 to identify the application. Format: A unique string identifier. |
| `status.outputs.client_id` | `string` | client_id is the OAuth 2.0 client identifier. This is the public identifier for the application. Used in authentication requests and as the "audience" for APIs. This value is safe to expose in client-side code. |
| `status.outputs.client_secret` | `string` (sensitive) | client_secret is the OAuth 2.0 client secret. Used for authenticating the application in confidential client flows. IMPORTANT: Keep this secret secure and never expose in client-side code. Only available for regular_web and non_interactive application types. |
| `status.outputs.name` | `string` | name is the name of the application. Derived from metadata.name in the Auth0Client resource. |
| `status.outputs.application_type` | `string` | application_type is the type of application. One of: native, spa, regular_web, non_interactive |
| `status.outputs.signing_keys` | `[]Auth0SigningKey` | signing_keys contains the signing keys for this client. Used for RS256 token signature verification. Contains certificate information for validating JWTs. |
| `status.outputs.signing_keys[].cert` | `string` | cert is the X.509 certificate in PEM format. |
| `status.outputs.signing_keys[].pkcs7` | `string` | pkcs7 is the PKCS#7 formatted certificate chain. |
| `status.outputs.signing_keys[].subject` | `string` | subject is the certificate subject. |
| `status.outputs.signing_keys[].thumbprint` | `string` | thumbprint is the SHA-1 thumbprint of the certificate. |
| `status.outputs.callback_url_template` | `string` | callback_url_template indicates if callback URL templating is enabled. When true, Auth0 allows parameterized callback URLs. NOTE: Not populated by default Terraform or Pulumi modules -- this attribute is only available via the Auth0 Management API and can be set by custom module overrides. |
| `status.outputs.allowed_clients` | `[]string` | allowed_clients lists clients allowed to perform delegation for this client. Used in legacy delegation flows. |
| `status.outputs.global` | `string` | global indicates if this is a global client (the tenant's default client). There is only one global client per tenant. NOTE: Not populated by default Terraform or Pulumi modules -- this attribute is only available via the Auth0 Management API and can be set by custom module overrides. |
| `status.outputs.token_endpoint_auth_method` | `string` | token_endpoint_auth_method is the authentication method for the token endpoint. Common values: "none", "client_secret_post", "client_secret_basic", "private_key_jwt" |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.enabledConnections` | Auth0Connection | `status.outputs.name` |
| `spec.apiGrants[].audience` | Auth0ResourceServer | `status.outputs.identifier` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| Auth0Connection | `spec.enabledClients` | `status.outputs.client_id` |

## See Also

- [Overview](../README.md)
