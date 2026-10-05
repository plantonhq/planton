# CloudflareZeroTrustAccessApplication

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `cloudflare.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

CloudflareZeroTrustAccessApplicationSpec defines a Cloudflare Zero Trust Access
application: the protected resource (a self-hosted web app, a SaaS app, an SSH
/VNC/RDP target, an app launcher, a warp/biso/bookmark entry, an
infrastructure target, or an MCP endpoint) that Cloudflare Access guards. The
application binds one or more standalone Access policies (referenced by ID) to
the resource and configures how users reach and authenticate to it.

An application is account-scoped or zone-scoped (exactly one of account_id or
zone_id). Account scope is the common case and lets the app reuse account-level
policies and groups.

## Example

```yaml
apiVersion: cloudflare.planton.dev/v1alpha1
kind: CloudflareZeroTrustAccessApplication
metadata:
  name: test-access-app
spec:
  accountId: "0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d"
  name: internal-dashboard
  type: self_hosted
  domain: dashboard.example.com
  sessionDuration: 24h
  policies:
    - policy:
        value: "00000000000000000000000000000000"
      precedence: 1
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.accountId` | `string` |  |  |  |
| `spec.zoneId` | `string \| valueFrom` |  |  | CloudflareDnsZone (`status.outputs.zone_id`) |
| `spec.name` | `string` | yes |  |  |
| `spec.type` | `enum` |  |  |  |
| `spec.domain` | `string` |  |  |  |
| `spec.policies` | `[]CloudflareZeroTrustAccessApplicationPolicyRef` |  |  |  |
| `spec.policies[].policy` | `string \| valueFrom` | yes |  | CloudflareZeroTrustAccessPolicy (`status.outputs.policy_id`) |
| `spec.policies[].precedence` | `uint32` |  |  |  |
| `spec.destinations` | `[]CloudflareZeroTrustAccessApplicationDestination` |  |  |  |
| `spec.destinations[].type` | `string` |  |  |  |
| `spec.destinations[].uri` | `string` |  |  |  |
| `spec.destinations[].cidr` | `string` |  |  |  |
| `spec.destinations[].hostname` | `string` |  |  |  |
| `spec.destinations[].l4Protocol` | `string` |  |  |  |
| `spec.destinations[].portRange` | `string` |  |  |  |
| `spec.destinations[].vnetId` | `string` |  |  |  |
| `spec.destinations[].mcpServerId` | `string` |  |  |  |
| `spec.allowedIdps` | `[]string \| valueFrom` |  |  |  |
| `spec.autoRedirectToIdentity` | `bool` |  |  |  |
| `spec.sessionDuration` | `string` |  |  |  |
| `spec.tags` | `[]string` |  |  |  |
| `spec.customPages` | `[]string` |  |  |  |
| `spec.appLauncherVisible` | `bool` |  |  |  |
| `spec.skipAppLauncherLoginPage` | `bool` |  |  |  |
| `spec.appLauncherLogoUrl` | `string` |  |  |  |
| `spec.bgColor` | `string` |  |  |  |
| `spec.headerBgColor` | `string` |  |  |  |
| `spec.logoUrl` | `string` |  |  |  |
| `spec.landingPageDesign` | `CloudflareZeroTrustAccessApplicationLandingPageDesign` |  |  |  |
| `spec.landingPageDesign.title` | `string` |  |  |  |
| `spec.landingPageDesign.message` | `string` |  |  |  |
| `spec.landingPageDesign.imageUrl` | `string` |  |  |  |
| `spec.landingPageDesign.buttonColor` | `string` |  |  |  |
| `spec.landingPageDesign.buttonTextColor` | `string` |  |  |  |
| `spec.footerLinks` | `[]CloudflareZeroTrustAccessApplicationFooterLink` |  |  |  |
| `spec.footerLinks[].name` | `string` | yes |  |  |
| `spec.footerLinks[].url` | `string` | yes |  |  |
| `spec.allowAuthenticateViaWarp` | `bool` |  |  |  |
| `spec.allowIframe` | `bool` |  |  |  |
| `spec.optionsPreflightBypass` | `bool` |  |  |  |
| `spec.readServiceTokensFromHeader` | `string` |  |  |  |
| `spec.sameSiteCookieAttribute` | `string` |  |  |  |
| `spec.serviceAuth401Redirect` | `bool` |  |  |  |
| `spec.skipInterstitial` | `bool` |  |  |  |
| `spec.enableBindingCookie` | `bool` |  |  |  |
| `spec.httpOnlyCookieAttribute` | `bool` |  |  |  |
| `spec.pathCookieAttribute` | `bool` |  |  |  |
| `spec.customDenyMessage` | `string` |  |  |  |
| `spec.customDenyUrl` | `string` |  |  |  |
| `spec.customNonIdentityDenyUrl` | `string` |  |  |  |
| `spec.corsHeaders` | `CloudflareZeroTrustAccessApplicationCorsHeaders` |  |  |  |
| `spec.corsHeaders.allowAllHeaders` | `bool` |  |  |  |
| `spec.corsHeaders.allowAllMethods` | `bool` |  |  |  |
| `spec.corsHeaders.allowAllOrigins` | `bool` |  |  |  |
| `spec.corsHeaders.allowCredentials` | `bool` |  |  |  |
| `spec.corsHeaders.allowedHeaders` | `[]string` |  |  |  |
| `spec.corsHeaders.allowedMethods` | `[]enum` |  |  |  |
| `spec.corsHeaders.allowedOrigins` | `[]string` |  |  |  |
| `spec.corsHeaders.maxAge` | `int32` |  |  |  |
| `spec.mfaConfig` | `CloudflareZeroTrustAccessApplicationMfaConfig` |  |  |  |
| `spec.mfaConfig.allowedAuthenticators` | `[]enum` |  |  |  |
| `spec.mfaConfig.mfaDisabled` | `bool` |  |  |  |
| `spec.mfaConfig.sessionDuration` | `string` |  |  |  |
| `spec.oauthConfiguration` | `CloudflareZeroTrustAccessApplicationOauthConfiguration` |  |  |  |
| `spec.oauthConfiguration.enabled` | `bool` |  |  |  |
| `spec.oauthConfiguration.dynamicClientRegistration` | `CloudflareZeroTrustAccessApplicationOauthDcr` |  |  |  |
| `spec.oauthConfiguration.dynamicClientRegistration.enabled` | `bool` |  |  |  |
| `spec.oauthConfiguration.dynamicClientRegistration.allowAnyOnLocalhost` | `bool` |  |  |  |
| `spec.oauthConfiguration.dynamicClientRegistration.allowAnyOnLoopback` | `bool` |  |  |  |
| `spec.oauthConfiguration.dynamicClientRegistration.allowedUris` | `[]string` |  |  |  |
| `spec.oauthConfiguration.grant` | `CloudflareZeroTrustAccessApplicationOauthGrant` |  |  |  |
| `spec.oauthConfiguration.grant.accessTokenLifetime` | `string` |  |  |  |
| `spec.oauthConfiguration.grant.sessionDuration` | `string` |  |  |  |
| `spec.targetCriteria` | `[]CloudflareZeroTrustAccessApplicationTargetCriteria` |  |  |  |
| `spec.targetCriteria[].port` | `uint32` | yes |  |  |
| `spec.targetCriteria[].protocol` | `enum` | yes |  |  |
| `spec.targetCriteria[].targetAttributes` | `[]CloudflareZeroTrustAccessApplicationTargetAttribute` | yes |  |  |
| `spec.targetCriteria[].targetAttributes[].name` | `string` | yes |  |  |
| `spec.targetCriteria[].targetAttributes[].values` | `[]string` | yes |  |  |
| `spec.saasApp` | `CloudflareZeroTrustAccessSaasApp` |  |  |  |
| `spec.saasApp.authType` | `enum` |  |  |  |
| `spec.saasApp.consumerServiceUrl` | `string` |  |  |  |
| `spec.saasApp.spEntityId` | `string` |  |  |  |
| `spec.saasApp.nameIdFormat` | `enum` |  |  |  |
| `spec.saasApp.nameIdTransformJsonata` | `string` |  |  |  |
| `spec.saasApp.samlAttributeTransformJsonata` | `string` |  |  |  |
| `spec.saasApp.defaultRelayState` | `string` |  |  |  |
| `spec.saasApp.customAttributes` | `[]CloudflareZeroTrustAccessSaasCustomAttribute` |  |  |  |
| `spec.saasApp.customAttributes[].name` | `string` |  |  |  |
| `spec.saasApp.customAttributes[].friendlyName` | `string` |  |  |  |
| `spec.saasApp.customAttributes[].nameFormat` | `string` |  |  |  |
| `spec.saasApp.customAttributes[].required` | `bool` |  |  |  |
| `spec.saasApp.customAttributes[].source` | `CloudflareZeroTrustAccessSaasAttributeSource` |  |  |  |
| `spec.saasApp.customAttributes[].source.name` | `string` |  |  |  |
| `spec.saasApp.customAttributes[].source.nameByIdp` | `[]CloudflareZeroTrustAccessSaasSourceNameByIdp` |  |  |  |
| `spec.saasApp.customAttributes[].source.nameByIdp[].idpId` | `string \| valueFrom` |  |  |  |
| `spec.saasApp.customAttributes[].source.nameByIdp[].sourceName` | `string` |  |  |  |
| `spec.saasApp.redirectUris` | `[]string` |  |  |  |
| `spec.saasApp.grantTypes` | `[]enum` |  |  |  |
| `spec.saasApp.scopes` | `[]enum` |  |  |  |
| `spec.saasApp.groupFilterRegex` | `string` |  |  |  |
| `spec.saasApp.appLauncherUrl` | `string` |  |  |  |
| `spec.saasApp.accessTokenLifetime` | `string` |  |  |  |
| `spec.saasApp.allowPkceWithoutClientSecret` | `bool` |  |  |  |
| `spec.saasApp.customClaims` | `[]CloudflareZeroTrustAccessSaasCustomClaim` |  |  |  |
| `spec.saasApp.customClaims[].name` | `string` |  |  |  |
| `spec.saasApp.customClaims[].required` | `bool` |  |  |  |
| `spec.saasApp.customClaims[].scope` | `enum` |  |  |  |
| `spec.saasApp.customClaims[].source` | `CloudflareZeroTrustAccessSaasClaimSource` |  |  |  |
| `spec.saasApp.customClaims[].source.name` | `string` |  |  |  |
| `spec.saasApp.customClaims[].source.nameByIdp` | `map<string, string>` |  |  |  |
| `spec.saasApp.hybridAndImplicitOptions` | `CloudflareZeroTrustAccessSaasHybridImplicitOptions` |  |  |  |
| `spec.saasApp.hybridAndImplicitOptions.returnAccessTokenFromAuthorizationEndpoint` | `bool` |  |  |  |
| `spec.saasApp.hybridAndImplicitOptions.returnIdTokenFromAuthorizationEndpoint` | `bool` |  |  |  |
| `spec.saasApp.refreshTokenOptions` | `CloudflareZeroTrustAccessSaasRefreshTokenOptions` |  |  |  |
| `spec.saasApp.refreshTokenOptions.lifetime` | `string` |  |  |  |
| `spec.scimConfig` | `CloudflareZeroTrustAccessScimConfig` |  |  |  |
| `spec.scimConfig.idpUid` | `string \| valueFrom` | yes |  |  |
| `spec.scimConfig.remoteUri` | `string` | yes |  |  |
| `spec.scimConfig.enabled` | `bool` |  |  |  |
| `spec.scimConfig.deactivateOnDelete` | `bool` |  |  |  |
| `spec.scimConfig.authentication` | `CloudflareZeroTrustAccessScimAuthentication` |  |  |  |
| `spec.scimConfig.authentication.scheme` | `enum` | yes |  |  |
| `spec.scimConfig.authentication.user` | `string` |  |  |  |
| `spec.scimConfig.authentication.password` | `string` (sensitive) |  |  |  |
| `spec.scimConfig.authentication.token` | `string` (sensitive) |  |  |  |
| `spec.scimConfig.authentication.clientId` | `string` |  |  |  |
| `spec.scimConfig.authentication.clientSecret` | `string` (sensitive) |  |  |  |
| `spec.scimConfig.authentication.authorizationUrl` | `string` |  |  |  |
| `spec.scimConfig.authentication.tokenUrl` | `string` |  |  |  |
| `spec.scimConfig.authentication.scopes` | `[]string` |  |  |  |
| `spec.scimConfig.mappings` | `[]CloudflareZeroTrustAccessScimMapping` |  |  |  |
| `spec.scimConfig.mappings[].schema` | `string` | yes |  |  |
| `spec.scimConfig.mappings[].enabled` | `bool` |  |  |  |
| `spec.scimConfig.mappings[].filter` | `string` |  |  |  |
| `spec.scimConfig.mappings[].strictness` | `enum` |  |  |  |
| `spec.scimConfig.mappings[].transformJsonata` | `string` |  |  |  |
| `spec.scimConfig.mappings[].operations` | `CloudflareZeroTrustAccessScimMappingOperations` |  |  |  |
| `spec.scimConfig.mappings[].operations.create` | `bool` |  |  |  |
| `spec.scimConfig.mappings[].operations.update` | `bool` |  |  |  |
| `spec.scimConfig.mappings[].operations.delete` | `bool` |  |  |  |

## Field Details

### spec.accountId

`string`

The Cloudflare account ID that owns this application. Set this for an
account-scoped application. Mutually exclusive with zone_id.

- rule: account_id must be a 32-character hex string

### spec.zoneId

`string | valueFrom`

The Cloudflare zone this application is scoped to, as a literal zone ID or a
reference to a CloudflareDnsZone. Mutually exclusive with account_id.

- references: CloudflareDnsZone (`status.outputs.zone_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: CloudflareDnsZone, name: <that resource's name>, fieldPath: status.outputs.zone_id}} -- a bare string does not parse

### spec.name

`string` · required

The display name of the Access application.

- rule: {"string":{"minLen":"1"}}

### spec.type

`enum`

The application type, which determines which other fields apply. Defaults to
self_hosted.

- rule: {"enum":{"definedOnly":true}}

Allowed values (use exactly as shown):

- `application_type_unspecified` -- Unspecified — treated as self_hosted.
- `self_hosted` -- A self-hosted web application fronted by Cloudflare.
- `saas` -- A SaaS application Cloudflare federates into (SAML/OIDC).
- `ssh` -- An SSH target brokered through Access.
- `vnc` -- A VNC target brokered through Access.
- `app_launcher` -- An Access app launcher.
- `warp` -- A WARP enrollment application.
- `biso` -- A browser-isolation (BISO) application.
- `bookmark` -- A bookmark tile in the app launcher.
- `dash_sso` -- The Cloudflare dashboard SSO application.
- `infrastructure` -- An infrastructure (targets) application.
- `rdp` -- An RDP target brokered through Access.
- `mcp` -- A Model Context Protocol (MCP) application.
- `mcp_portal` -- An MCP portal application.
- `proxy_endpoint` -- A proxy-endpoint application.

### spec.domain

`string`

The primary fully-qualified domain protected by this application (e.g.
"app.example.com"). Required for self_hosted/ssh/vnc/rdp; ignored for types
that don't front a single hostname. The hostname must sit on a zone this
account owns AND that zone must be ACTIVE (registrar-delegated): Cloudflare
rejects the create with "access.api.error.invalid_request: domain does not
belong to zone" while the zone is still pending (measured live).

### spec.policies

`[]CloudflareZeroTrustAccessApplicationPolicyRef`

The policies that govern this application, in evaluation order. Each entry
references a standalone CloudflareZeroTrustAccessPolicy by ID; precedence
(lower first) breaks ties, otherwise list order applies.

### spec.policies[].policy

`string | valueFrom` · required

The policy to attach, as a literal policy ID or a reference to a
CloudflareZeroTrustAccessPolicy.

- references: CloudflareZeroTrustAccessPolicy (`status.outputs.policy_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: CloudflareZeroTrustAccessPolicy, name: <that resource's name>, fieldPath: status.outputs.policy_id}} -- a bare string does not parse

### spec.policies[].precedence

`uint32`

Evaluation order (lower is evaluated first). Leave 0 to use list order.

### spec.destinations

`[]CloudflareZeroTrustAccessApplicationDestination`

Additional destinations this application serves beyond `domain` (public URLs
and private network targets). The modern replacement for self-hosted domain
lists; applies to self_hosted/ssh/vnc/rdp/mcp(_portal).

### spec.destinations[].type

`string`

Destination kind: "public" (a public URI) or "private" (a network target
reachable over a Cloudflare Tunnel virtual network). Empty defaults to public.

- rule: destination type must be "public", "private", or "via_mcp_server_portal"

### spec.destinations[].uri

`string`

For public destinations: the URI (e.g. "https://app.example.com").

### spec.destinations[].cidr

`string`

For private destinations: the CIDR the target falls within.

### spec.destinations[].hostname

`string`

For private destinations: the target hostname.

### spec.destinations[].l4Protocol

`string`

For private destinations: the layer-4 protocol ("tcp" or "udp").

- rule: l4_protocol must be "tcp" or "udp"

### spec.destinations[].portRange

`string`

For private destinations: the port range (e.g. "8080" or "8080-8090").

### spec.destinations[].vnetId

`string`

For private destinations: the Cloudflare Tunnel virtual-network ID.

### spec.destinations[].mcpServerId

`string`

For via_mcp_server_portal destinations: the MCP server ID.

### spec.allowedIdps

`[]string | valueFrom`

Identity providers allowed to authenticate to this application, by IdP ID
(literal or reference to another resource's output). Empty allows all
configured IdPs.

- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.autoRedirectToIdentity

`bool`

Automatically redirect the user to the identity provider when only one is
configured, skipping the IdP chooser. Defaults to false.

### spec.sessionDuration

`string`

How long a user session is valid before re-authentication, as a duration
string (e.g. "24h"). Empty uses the account default.

### spec.tags

`[]string`

Free-form tags used to organize applications in the Access dashboard.

### spec.customPages

`[]string`

Names of custom Access pages (block / identity-denied) to render for this
application.

### spec.appLauncherVisible

`bool` · optional (explicit presence)

Whether the application appears in the Access app launcher. Applies to
self_hosted/ssh/vnc/rdp/saas/bookmark. Omit to use the provider default.

### spec.skipAppLauncherLoginPage

`bool`

For app_launcher applications, skip the app-launcher login page. Defaults to
false.

### spec.appLauncherLogoUrl

`string`

App-launcher logo URL (app_launcher type only).

### spec.bgColor

`string`

App-launcher background color (app_launcher type only).

### spec.headerBgColor

`string`

App-launcher header background color (app_launcher type only).

### spec.logoUrl

`string`

Logo URL shown on the Access login page.

### spec.landingPageDesign

`CloudflareZeroTrustAccessApplicationLandingPageDesign`

Visual design of the app-launcher landing page (app_launcher type only).

### spec.landingPageDesign.title

`string`

Title text.

### spec.landingPageDesign.message

`string`

Body message.

### spec.landingPageDesign.imageUrl

`string`

Hero image URL.

### spec.landingPageDesign.buttonColor

`string`

Primary button background color.

### spec.landingPageDesign.buttonTextColor

`string`

Primary button text color.

### spec.footerLinks

`[]CloudflareZeroTrustAccessApplicationFooterLink`

Footer links shown on the app-launcher page (app_launcher type only).

### spec.footerLinks[].name

`string` · required

The link label.

- rule: {"required":true}

### spec.footerLinks[].url

`string` · required

The link URL.

- rule: {"required":true}

### spec.allowAuthenticateViaWarp

`bool`

Allow users to authenticate to this application from the WARP client.
Applies to self_hosted/ssh/vnc/rdp/saas/dash_sso.

### spec.allowIframe

`bool`

Allow this application to be rendered inside an iframe. Applies to
self_hosted/ssh/vnc/rdp/mcp_portal.

### spec.optionsPreflightBypass

`bool`

Bypass Access for CORS preflight (OPTIONS) requests. Cannot be combined with
cors_headers. Self-hosted types.

### spec.readServiceTokensFromHeader

`string`

Read the Access service-token from this request header instead of the default.
Self-hosted types.

### spec.sameSiteCookieAttribute

`string`

SameSite attribute for the Access cookie ("lax", "strict", or "none"). Empty
uses the Cloudflare default. Self-hosted types.

- rule: same_site_cookie_attribute must be "lax", "strict", or "none"

### spec.serviceAuth401Redirect

`bool`

Return a 401 instead of redirecting to login when a service-auth request is
unauthenticated. Self-hosted types.

### spec.skipInterstitial

`bool`

Skip the Access interstitial confirmation page (e.g. for SSH). Self-hosted
types.

### spec.enableBindingCookie

`bool`

Bind the Access cookie to the user's device. Self-hosted types.

### spec.httpOnlyCookieAttribute

`bool` · optional (explicit presence)

Add the HttpOnly flag to the Access cookie. Self-hosted types. Omit to use the
provider default.

### spec.pathCookieAttribute

`bool`

Scope the Access cookie to the application path. Self-hosted types.

### spec.customDenyMessage

`string`

Custom message shown to users who are denied access.

### spec.customDenyUrl

`string`

URL to redirect denied users to (instead of the default block page).

### spec.customNonIdentityDenyUrl

`string`

URL to redirect denied non-identity (service-token) requests to.

### spec.corsHeaders

`CloudflareZeroTrustAccessApplicationCorsHeaders`

Cross-origin resource sharing (CORS) settings for the application. Self-hosted
types. Mutually exclusive with options_preflight_bypass.

### spec.corsHeaders.allowAllHeaders

`bool`

Allow all request headers cross-origin (mutually exclusive with
allowed_headers).

### spec.corsHeaders.allowAllMethods

`bool`

Allow all methods cross-origin (mutually exclusive with allowed_methods).

### spec.corsHeaders.allowAllOrigins

`bool`

Allow all origins cross-origin (mutually exclusive with allowed_origins).

### spec.corsHeaders.allowCredentials

`bool`

Allow credentials (cookies/authorization) on cross-origin requests. Cannot be
combined with allow_all_origins.

### spec.corsHeaders.allowedHeaders

`[]string`

Explicit list of allowed request headers.

### spec.corsHeaders.allowedMethods

`[]enum`

Explicit list of allowed methods.

- rule: {"repeated":{"items":{"enum":{"definedOnly":true}}}}

Allowed values (use exactly as shown):

- `method_unspecified` -- Unspecified (invalid).
- `GET`
- `POST`
- `HEAD`
- `PUT`
- `DELETE`
- `CONNECT`
- `OPTIONS`
- `TRACE`
- `PATCH`

### spec.corsHeaders.allowedOrigins

`[]string`

Explicit list of allowed origins.

### spec.corsHeaders.maxAge

`int32`

Preflight cache lifetime in seconds (-1 to 86400).

- rule: max_age must be between -1 and 86400

### spec.mfaConfig

`CloudflareZeroTrustAccessApplicationMfaConfig`

Application-level multi-factor requirements (self_hosted/ssh/vnc/rdp).

### spec.mfaConfig.allowedAuthenticators

`[]enum`

The set of authenticators a user may use to satisfy MFA.

- rule: allowed_authenticators must be totp, biometrics, or security_key
- rule: {"repeated":{"items":{"enum":{"definedOnly":true}}}}

Allowed values (use exactly as shown):

- `authenticator_unspecified` -- Unspecified (invalid).
- `totp` -- Time-based one-time password.
- `biometrics` -- Platform biometrics.
- `security_key` -- Hardware/passkey security key.

### spec.mfaConfig.mfaDisabled

`bool`

Disable MFA enforcement (overrides allowed_authenticators).

### spec.mfaConfig.sessionDuration

`string`

Re-prompt for MFA after this duration (e.g. "8h").

### spec.oauthConfiguration

`CloudflareZeroTrustAccessApplicationOauthConfiguration`

OAuth/MCP authorization-server settings (for MCP applications acting as an
OAuth provider).

### spec.oauthConfiguration.enabled

`bool`

Whether the OAuth authorization server is enabled.

### spec.oauthConfiguration.dynamicClientRegistration

`CloudflareZeroTrustAccessApplicationOauthDcr`

Dynamic client registration settings.

### spec.oauthConfiguration.dynamicClientRegistration.enabled

`bool`

Whether dynamic client registration is enabled.

### spec.oauthConfiguration.dynamicClientRegistration.allowAnyOnLocalhost

`bool`

Allow registering clients with any localhost redirect URI.

### spec.oauthConfiguration.dynamicClientRegistration.allowAnyOnLoopback

`bool`

Allow registering clients with any loopback redirect URI.

### spec.oauthConfiguration.dynamicClientRegistration.allowedUris

`[]string`

Explicit list of allowed redirect URIs.

### spec.oauthConfiguration.grant

`CloudflareZeroTrustAccessApplicationOauthGrant`

Token / session grant settings.

### spec.oauthConfiguration.grant.accessTokenLifetime

`string`

Access-token lifetime (e.g. "1h").

### spec.oauthConfiguration.grant.sessionDuration

`string`

Session duration (e.g. "24h").

### spec.targetCriteria

`[]CloudflareZeroTrustAccessApplicationTargetCriteria`

Target selection criteria for rdp/infrastructure applications (which targets,
ports, and attributes this application brokers).

### spec.targetCriteria[].port

`uint32` · required

The target port.

- rule: {"required":true,"uint32":{"lte":65535}}

### spec.targetCriteria[].protocol

`enum` · required

The protocol.

- rule: protocol must be SSH or RDP
- rule: {"required":true,"enum":{"definedOnly":true}}

Allowed values (use exactly as shown):

- `protocol_unspecified` -- Unspecified (invalid).
- `SSH` -- Secure Shell.
- `RDP` -- Remote Desktop Protocol.

### spec.targetCriteria[].targetAttributes

`[]CloudflareZeroTrustAccessApplicationTargetAttribute` · required

Target attributes (e.g. "hostname" -> ["db-1","db-2"]) used to match the
brokered targets. At least one attribute is required.

- rule: {"repeated":{"minItems":"1"}}

### spec.targetCriteria[].targetAttributes[].name

`string` · required

The attribute name (e.g. "hostname").

- rule: {"required":true}

### spec.targetCriteria[].targetAttributes[].values

`[]string` · required

The attribute values.

- rule: {"repeated":{"minItems":"1"}}

### spec.saasApp

`CloudflareZeroTrustAccessSaasApp`

SaaS application settings (SAML or OIDC). Required when type is saas or
dash_sso; ignored otherwise.

### spec.saasApp.authType

`enum`

The federation protocol. Defaults to saml.

- rule: {"enum":{"definedOnly":true}}

Allowed values (use exactly as shown):

- `auth_type_unspecified` -- Unspecified — defaults to saml.
- `saml` -- SAML 2.0.
- `oidc` -- OpenID Connect.

### spec.saasApp.consumerServiceUrl

`string`

The SAML Assertion Consumer Service (ACS) URL the IdP posts assertions to.

### spec.saasApp.spEntityId

`string`

The Service Provider entity ID.

### spec.saasApp.nameIdFormat

`enum`

The SAML NameID format.

- rule: {"enum":{"definedOnly":true}}

Allowed values (use exactly as shown):

- `name_id_format_unspecified` -- Unspecified — provider default.
- `email` -- Use the user's email as the NameID.
- `id` -- Use the user's unique ID as the NameID.

### spec.saasApp.nameIdTransformJsonata

`string`

A JSONata expression transforming the NameID before it is sent.

### spec.saasApp.samlAttributeTransformJsonata

`string`

A JSONata expression transforming SAML attributes before they are sent.

### spec.saasApp.defaultRelayState

`string`

The default RelayState passed to the IdP.

### spec.saasApp.customAttributes

`[]CloudflareZeroTrustAccessSaasCustomAttribute`

Custom SAML attributes to include in the assertion.

### spec.saasApp.customAttributes[].name

`string`

The attribute name.

### spec.saasApp.customAttributes[].friendlyName

`string`

A friendly name for the attribute.

### spec.saasApp.customAttributes[].nameFormat

`string`

The SAML name format URN (e.g.
"urn:oasis:names:tc:SAML:2.0:attrname-format:basic").

- rule: name_format must be a SAML attrname-format URN (unspecified, basic, or uri) when set

### spec.saasApp.customAttributes[].required

`bool`

Whether the attribute is required in the assertion.

### spec.saasApp.customAttributes[].source

`CloudflareZeroTrustAccessSaasAttributeSource`

The source of the attribute value.

### spec.saasApp.customAttributes[].source.name

`string`

The source attribute name at the IdP.

### spec.saasApp.customAttributes[].source.nameByIdp

`[]CloudflareZeroTrustAccessSaasSourceNameByIdp`

Per-IdP source-name overrides.

### spec.saasApp.customAttributes[].source.nameByIdp[].idpId

`string | valueFrom`

The identity-provider ID, as a literal or a reference to another resource's
output.

- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.saasApp.customAttributes[].source.nameByIdp[].sourceName

`string`

The source attribute name to use for that IdP.

### spec.saasApp.redirectUris

`[]string`

Allowed OIDC redirect URIs.

### spec.saasApp.grantTypes

`[]enum`

OIDC grant types the app supports.

- rule: {"repeated":{"items":{"enum":{"definedOnly":true}}}}

Allowed values (use exactly as shown):

- `grant_type_unspecified` -- Unspecified (invalid).
- `authorization_code`
- `authorization_code_with_pkce`
- `refresh_tokens`
- `hybrid`
- `implicit`

### spec.saasApp.scopes

`[]enum`

OIDC scopes the app requests.

- rule: {"repeated":{"items":{"enum":{"definedOnly":true}}}}

Allowed values (use exactly as shown):

- `scope_unspecified` -- Unspecified (invalid).
- `openid`
- `groups`
- `email`
- `profile`

### spec.saasApp.groupFilterRegex

`string`

A regex selecting which groups are emitted in the OIDC token.

### spec.saasApp.appLauncherUrl

`string`

The app-launcher URL for the OIDC app.

### spec.saasApp.accessTokenLifetime

`string`

OIDC access-token lifetime (e.g. "10m").

### spec.saasApp.allowPkceWithoutClientSecret

`bool`

Allow PKCE without a client secret.

### spec.saasApp.customClaims

`[]CloudflareZeroTrustAccessSaasCustomClaim`

Custom OIDC claims to include in the token.

### spec.saasApp.customClaims[].name

`string`

The claim name.

### spec.saasApp.customClaims[].required

`bool`

Whether the claim is required.

### spec.saasApp.customClaims[].scope

`enum`

The OIDC scope the claim belongs to.

- rule: {"enum":{"definedOnly":true}}

Allowed values (use exactly as shown):

- `scope_unspecified` -- Unspecified (invalid).
- `openid`
- `groups`
- `email`
- `profile`

### spec.saasApp.customClaims[].source

`CloudflareZeroTrustAccessSaasClaimSource`

The source of the claim value.

### spec.saasApp.customClaims[].source.name

`string`

The source attribute name at the IdP.

### spec.saasApp.customClaims[].source.nameByIdp

`map<string, string>`

Per-IdP source-name overrides (IdP ID -> source attribute name).

### spec.saasApp.hybridAndImplicitOptions

`CloudflareZeroTrustAccessSaasHybridImplicitOptions`

Hybrid/implicit flow options.

### spec.saasApp.hybridAndImplicitOptions.returnAccessTokenFromAuthorizationEndpoint

`bool`

Return an access token from the authorization endpoint.

### spec.saasApp.hybridAndImplicitOptions.returnIdTokenFromAuthorizationEndpoint

`bool`

Return an ID token from the authorization endpoint.

### spec.saasApp.refreshTokenOptions

`CloudflareZeroTrustAccessSaasRefreshTokenOptions`

Refresh-token options.

### spec.saasApp.refreshTokenOptions.lifetime

`string`

Refresh-token lifetime (e.g. "30d").

### spec.scimConfig

`CloudflareZeroTrustAccessScimConfig`

SCIM provisioning settings (push users/groups from an IdP to the application).

### spec.scimConfig.idpUid

`string | valueFrom` · required

The IdP UID supplying the SCIM source, as a literal or a reference to another
resource's output.

- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.scimConfig.remoteUri

`string` · required

The SCIM remote URI Cloudflare provisions against.

- rule: {"required":true}

### spec.scimConfig.enabled

`bool`

Whether SCIM provisioning is enabled.

### spec.scimConfig.deactivateOnDelete

`bool`

De-provision (deactivate) users in the application when they are removed at
the source.

### spec.scimConfig.authentication

`CloudflareZeroTrustAccessScimAuthentication`

Authentication used to reach the SCIM endpoint.

### spec.scimConfig.authentication.scheme

`enum` · required

The authentication scheme.

- rule: scheme must be httpbasic, oauthbearertoken, oauth2, or access_service_token
- rule: {"required":true,"enum":{"definedOnly":true}}

Allowed values (use exactly as shown):

- `scheme_unspecified` -- Unspecified (invalid).
- `httpbasic` -- HTTP basic auth (user + password).
- `oauthbearertoken` -- OAuth bearer token.
- `oauth2` -- OAuth 2.0 client credentials.
- `access_service_token` -- Cloudflare Access service token.

### spec.scimConfig.authentication.user

`string`

Basic-auth username (httpbasic).

### spec.scimConfig.authentication.password

`string` · sensitive

Basic-auth password (httpbasic).

### spec.scimConfig.authentication.token

`string` · sensitive

Bearer token (oauthbearertoken).

### spec.scimConfig.authentication.clientId

`string`

OAuth client ID (oauth2).

### spec.scimConfig.authentication.clientSecret

`string` · sensitive

OAuth client secret (oauth2).

### spec.scimConfig.authentication.authorizationUrl

`string`

OAuth authorization URL (oauth2).

### spec.scimConfig.authentication.tokenUrl

`string`

OAuth token URL (oauth2).

### spec.scimConfig.authentication.scopes

`[]string`

OAuth scopes (oauth2).

### spec.scimConfig.mappings

`[]CloudflareZeroTrustAccessScimMapping`

Attribute/group mapping rules.

### spec.scimConfig.mappings[].schema

`string` · required

The SCIM schema URI this mapping applies to.

- rule: {"required":true}

### spec.scimConfig.mappings[].enabled

`bool`

Whether this mapping is enabled.

### spec.scimConfig.mappings[].filter

`string`

A filter expression selecting which resources the mapping applies to.

### spec.scimConfig.mappings[].strictness

`enum`

Mapping strictness.

- rule: {"enum":{"definedOnly":true}}

Allowed values (use exactly as shown):

- `strictness_unspecified` -- Unspecified — provider default.
- `strict` -- Strictly apply the mapping.
- `passthrough` -- Pass through unmapped attributes.

### spec.scimConfig.mappings[].transformJsonata

`string`

A JSONata transform applied to the mapped resource.

### spec.scimConfig.mappings[].operations

`CloudflareZeroTrustAccessScimMappingOperations`

Per-operation toggles (create/update/delete).

### spec.scimConfig.mappings[].operations.create

`bool`

Apply the mapping on create.

### spec.scimConfig.mappings[].operations.update

`bool`

Apply the mapping on update.

### spec.scimConfig.mappings[].operations.delete

`bool`

Apply the mapping on delete.

## Validation Rules

- `spec.account_xor_zone`: set exactly one of account_id or zone_id
- `spec.domain_required_for_self_hosted`: domain is required for self_hosted, ssh, vnc, and rdp application types

## Outputs

Reference an output from another manifest as `valueFrom: {kind: CloudflareZeroTrustAccessApplication, name: <resource-name>, fieldPath: status.outputs.<output>}`. A sensitive output is a secret the resource generates: on Planton it is kept in the organization's secret store and the output holds a `$secret/` reference, so feed it only to a sensitive field.

| Output | Type | Description |
|---|---|---|
| `status.outputs.application_id` | `string` | The unique ID of the Access application. |
| `status.outputs.aud` | `string` | The application's audience (AUD) tag. Downstream services and Workers use this to validate the Cloudflare Access JWT for requests to this application. |
| `status.outputs.domain` | `string` | The primary domain protected by this application (echoes the input domain when set, otherwise the provider-resolved value). |
| `status.outputs.saas_client_id` | `string` | For SaaS (OIDC) applications: the issued OAuth client ID. |
| `status.outputs.saas_client_secret` | `string` (sensitive) | For SaaS (OIDC) applications: the issued OAuth client secret. |
| `status.outputs.saas_public_key` | `string` | For SaaS (SAML) applications: the IdP-facing public key (certificate). |
| `status.outputs.saas_sso_endpoint` | `string` | For SaaS (SAML) applications: the single sign-on (SSO) endpoint URL. |
| `status.outputs.saas_idp_entity_id` | `string` | For SaaS (SAML) applications: the IdP entity ID. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.zoneId` | CloudflareDnsZone | `status.outputs.zone_id` |
| `spec.policies[].policy` | CloudflareZeroTrustAccessPolicy | `status.outputs.policy_id` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| CloudflareZeroTrustTunnel | `spec.ingress[].originRequest.access.audTag` | `status.outputs.aud` |
| CloudflareZeroTrustTunnel | `spec.originRequest.access.audTag` | `status.outputs.aud` |

## See Also

- [Overview](../README.md)
