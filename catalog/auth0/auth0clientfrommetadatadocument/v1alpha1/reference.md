# Auth0ClientFromMetadataDocument

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `auth0.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

Auth0ClientFromMetadataDocumentSpec registers an application in the Auth0
tenant the provider connection's credential belongs to from its Client ID
Metadata Document: a JSON file the application's owner hosts at an https URL
(external_client_id). Auth0 fetches the document, validates it, and registers
the application from it -- the path an MCP client takes to onboard itself,
and the one partner integrations take when they manage their own metadata and
keys. In sign-in flows the application presents the document's URL as its
client_id; Auth0 also assigns it a Management API id (output client_id,
"tpc_...") that grants and connections reference.

The document owns the application's identity, and this spec never states it:
its name (client_name), its redirect URIs (redirect_uris), its logo
(logo_uri), how it authenticates at the token endpoint
(token_endpoint_auth_method: "none" for a public client using PKCE, or
"private_key_jwt" with a jwks_uri for a confidential one) and its public keys
(jwks_uri) come from the document and are reported back as outputs. Three
settings are seeded from the document and may be set here over it:
app_type (the document's application_type), grant_types and description.
Everything else -- origins, token lifetimes, refresh-token rotation, the
redirect policy, organizations, metadata -- is the tenant's to set, and the
document has no say in it. Changing the document changes nothing in Auth0
until Auth0 fetches it again (external_client_id_version).

Every application registered this way is a third-party application in
strict mode (output third_party_security_mode): it signs people in only
through connections promoted to the domain level, reaches an API only
through an explicit client grant (never the Management API), and holds only
the authorization_code and refresh_token grants. Login flows for these
applications fail while the tenant has active Rules.

Unset fields are NOT MANAGED: the module sends only what the spec declares,
and Auth0 keeps what the document or the tenant set. Some settings cannot be
left alone by the provider once Auth0 holds a value for them; their comments
say so, and the kind's guide lists them for adoption.

Prerequisite: the tenant must allow registration from metadata documents
(Auth0TenantSettings' Client ID Metadata Document Registration setting).

Plans: registration itself is on every plan. A confidential client
(private_key_jwt) needs the Enterprise plan; the settings that need more say
so in their comments.

The credential needs create:clients, read:clients, update:clients and
delete:clients on the tenant's Management API (iac/permissions.yaml).

https://auth0.com/docs/get-started/auth0-overview/create-applications/register-applications-with-cimd
https://auth0.com/docs/get-started/applications/third-party-applications/security-controls
https://registry.terraform.io/providers/auth0/auth0/latest/docs/resources/client_cimd
https://www.pulumi.com/registry/packages/auth0/api-docs/clientcimd/

## Example

```yaml
# Auth0 Client From Metadata Document Test Manifest
# This file is used for testing the Auth0ClientFromMetadataDocument component.
#
# Prerequisites:
# 1. Set the following environment variables:
#    - AUTH0_DOMAIN: the test tenant's domain (e.g., your-test-tenant.auth0.com)
#    - AUTH0_CLIENT_ID: M2M application client ID
#    - AUTH0_CLIENT_SECRET: M2M application client secret
#
# 2. The M2M application must have these scopes:
#    - create:clients
#    - read:clients
#    - update:clients
#    - delete:clients
#
# 3. The tenant must allow Client ID Metadata Document registration
#    (Auth0TenantSettings), and a metadata document must be served at
#    externalClientId whose client_id is exactly that URL.

apiVersion: auth0.planton.dev/v1alpha1
kind: Auth0ClientFromMetadataDocument
metadata:
  name: test-mcp-client
  org: test-org
  env: development
  labels:
    purpose: testing
spec:
  # The URL the client serves its metadata document from
  externalClientId: https://mcp-client.example.com/.well-known/oauth-client-metadata

  # Raise to make Auth0 fetch the document again
  externalClientIdVersion: 1

  # Declared so plans stay clean when the document carries a description
  description: MCP client for end-to-end testing

  # Sign-in plus refresh tokens -- the two grants these applications may hold
  grantTypes:
    - authorization_code
    - refresh_token

  # Refresh tokens that rotate on every use and expire
  refreshToken:
    rotationType: rotating
    expirationType: expiring
    tokenLifetime: 2592000
    idleTokenLifetime: 1296000
    leeway: 0
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.externalClientId` | `string` | yes |  |  |
| `spec.externalClientIdVersion` | `int32` |  |  |  |
| `spec.appType` | `string` |  |  |  |
| `spec.grantTypes` | `[]string` |  |  |  |
| `spec.description` | `string` |  |  |  |
| `spec.allowedOrigins` | `[]string` |  |  |  |
| `spec.webOrigins` | `[]string` |  |  |  |
| `spec.oidcConformant` | `bool` |  |  |  |
| `spec.requireProofOfPossession` | `bool` |  |  |  |
| `spec.skipNonVerifiableCallbackUriConfirmationPrompt` | `bool` |  |  |  |
| `spec.redirectionPolicy` | `string` |  |  |  |
| `spec.organizationDiscoveryMethods` | `[]string` |  |  |  |
| `spec.defaultOrganization` | `Auth0ClientFromMetadataDocumentDefaultOrganization` |  |  |  |
| `spec.defaultOrganization.organizationId` | `string` | yes |  |  |
| `spec.defaultOrganization.flows` | `[]string` | yes |  |  |
| `spec.clientMetadata` | `map<string, string>` |  |  |  |
| `spec.jwtConfiguration` | `Auth0ClientFromMetadataDocumentJwtConfiguration` |  |  |  |
| `spec.jwtConfiguration.alg` | `string` |  |  |  |
| `spec.jwtConfiguration.lifetimeInSeconds` | `int32` |  |  |  |
| `spec.refreshToken` | `Auth0ClientFromMetadataDocumentRefreshToken` |  |  |  |
| `spec.refreshToken.rotationType` | `string` |  |  |  |
| `spec.refreshToken.expirationType` | `string` |  |  |  |
| `spec.refreshToken.leeway` | `int32` |  |  |  |
| `spec.refreshToken.tokenLifetime` | `int32` |  |  |  |
| `spec.refreshToken.infiniteTokenLifetime` | `bool` |  |  |  |
| `spec.refreshToken.idleTokenLifetime` | `int32` |  |  |  |
| `spec.refreshToken.infiniteIdleTokenLifetime` | `bool` |  |  |  |
| `spec.tokenQuota` | `Auth0ClientFromMetadataDocumentTokenQuota` |  |  |  |
| `spec.tokenQuota.clientCredentials` | `Auth0ClientFromMetadataDocumentTokenQuotaClientCredentials` | yes |  |  |
| `spec.tokenQuota.clientCredentials.enforce` | `bool` |  |  |  |
| `spec.tokenQuota.clientCredentials.perDay` | `int32` |  |  |  |
| `spec.tokenQuota.clientCredentials.perHour` | `int32` |  |  |  |

## Field Details

### spec.externalClientId

`string` · required

external_client_id is the https URL the application serves its Client ID
Metadata Document from, for example
"https://mcp-client.example.com/.well-known/oauth-client-metadata". It is
the application's client_id in every sign-in flow, and the document's own
client_id must be exactly this URL. Auth0 fetches the document without
following redirects, and refuses one larger than 5 KB.

Auth0 keys registrations by this URL: applying a URL the tenant has
already registered (in the dashboard, or by the client itself) takes that
registration over, and destroying this resource deletes it. Changing the
URL replaces the application -- a new client_id, so every grant and
connection that named the old one must be applied again.

A document declaring private_key_jwt (a confidential client, with a
jwks_uri on the document's own origin) needs the Enterprise plan; a public
client ("none", with PKCE) does not.

- rule: external_client_id must be the https:// URL the metadata document is served from (for example https://mcp-client.example.com/.well-known/oauth-client-metadata) -- Auth0 fetches client metadata only over HTTPS
- rule: external_client_id needs a path after the host (for example https://mcp-client.example.com/client.json) -- Auth0 refuses a document at the site's root
- rule: external_client_id cannot carry a query (?), a fragment (#), a user or password (user@host), port 0, or whitespace -- Auth0 refuses such document URLs
- rule: external_client_id cannot point at localhost, 127.0.0.1 or [::1] -- Auth0 fetches the document from the public internet
- rule: external_client_id cannot contain . or .. path segments (written plainly or as %2e) -- Auth0 refuses them; write the document's path directly
- rule: every % in external_client_id must start a two-digit hex escape (such as %20)
- rule: {"required":true,"string":{"maxLen":"120"}}

### spec.externalClientIdVersion

`int32` · optional (explicit presence)

external_client_id_version is a number that makes Auth0 fetch the
document again when it changes: change it (1, then 2, ...) after the
document's owner has changed the document, and the next apply re-registers
the application from it -- its name, redirect URIs, logo and keys, and its
application type and grant types, followed by this spec's own settings
applied over them again. Its value means nothing to Auth0; only a change
does. A key the owner rotates must arrive under a new kid: Auth0 keeps the
old key when new material reuses a kid. Unset, Auth0 keeps what it fetched
at registration.

### spec.appType

`string` · optional (explicit presence)

app_type is the kind of application Auth0 treats this as, seeded from the
document's application_type ("native" for native, "web" for
"regular_web"). Unset, the document's value stands; set, it overrides the
document until the next fetch (external_client_id_version), after which it
is applied again.
- "native": a desktop, mobile or command-line application
- "regular_web": a web application with a server
- "spa": a single-page application; the provider accepts it, while Auth0's
  CIMD documentation names only native and regular_web

- rule: {"string":{"in":["native","regular_web","spa"]}}

### spec.grantTypes

`[]string`

grant_types are the grants the application may use, seeded from the
document's grant_types. Auth0 allows these applications only
"authorization_code" (sign-in, with PKCE) and "refresh_token" (keeping a
session without signing in again; needed for refresh_token below to mean
anything). Empty, the document's list stands; set, it overrides the
document the way app_type does.

- rule: {"repeated":{"items":{"string":{"in":["authorization_code","refresh_token"]}}}}

### spec.description

`string` · optional (explicit presence)

description is a free-text description of the application, at most 140
characters, seeded from the document's description. A fresh registration
keeps the document's description while this is unset, but the provider
does not read an unset description as "keep" on an adopted application:
once Auth0's value is in state (after an import), every plan proposes to
clear it. Declare it when adopting (the document's words or your own).

- rule: {"string":{"maxLen":"140"}}

### spec.allowedOrigins

`[]string`

allowed_origins are the origins allowed to call Auth0 from a browser
(cross-origin requests), for example "https://mcp-client.example.com".
Wildcards such as "https://*.example.com" are accepted. Empty is not
managed. Declare the live value when adopting an existing application;
leaving it unset on an adopted application resets it.

### spec.webOrigins

`[]string`

web_origins are the origins allowed to use web-message response mode --
silent authentication and token renewal from a browser in a hidden frame.
Empty is not managed. Declare the live value when adopting an existing
application; leaving it unset on an adopted application resets it.

### spec.oidcConformant

`bool` · optional (explicit presence)

oidc_conformant is whether the application follows the OpenID Connect
specification strictly. Auth0 requires it for these applications, so it
can only be true; unset, Auth0 keeps it on.

- rule: oidc_conformant can only be true -- Auth0 requires applications registered from a metadata document to follow OpenID Connect strictly (leave it unset to keep it on)

### spec.requireProofOfPossession

`bool` · optional (explicit presence)

require_proof_of_possession makes sender-constrained tokens mandatory: the
application must prove on every token request that it holds its key, with
DPoP (a proof signed by the application's key) or mutual TLS, so a stolen
token is useless to anyone else. The APIs it calls must then accept
sender-constrained tokens (Auth0ResourceServer). Mutual TLS needs the
Enterprise plan with the Highly Regulated Identity add-on. The provider
does not read an unset value as "keep": declare the live value when
adopting an application that requires it, or every plan proposes to clear
it.

### spec.skipNonVerifiableCallbackUriConfirmationPrompt

`bool` · optional (explicit presence)

skip_non_verifiable_callback_uri_confirmation_prompt skips the prompt
Auth0 shows before redirecting to a callback it cannot verify belongs to
the application -- a custom scheme such as "myapp://" or a localhost
address, common for a native MCP client on a person's machine. Auth0
recommends leaving the prompt on (false): it is what stops a malicious app
posing as this one from receiving the code silently. Declare the live
value when adopting an existing application; leaving it unset on an
adopted application resets it.

### spec.redirectionPolicy

`string` · optional (explicit presence)

redirection_policy decides whether Auth0 sends people back to the
application's callback when sign-in fails, and in email flows.
- "open_redirect_protection": Auth0 shows its own error page instead, and
  email templates do not see the callback's domain
  (application.callback_domain) -- the default for third-party
  applications, because a redirect nobody clicked is a phishing vector
  when the callback belongs to someone else
- "allow_always": the standard redirect; only for an application whose
  callback URIs you trust
Unset, Auth0 keeps open_redirect_protection.

- rule: {"string":{"in":["allow_always","open_redirect_protection"]}}

### spec.organizationDiscoveryMethods

`[]string`

organization_discovery_methods are how a person finds their Auth0
Organization on the prompt before sign-in:
- "email": by the domain of the email address they enter
- "organization_name": by typing the organization's name
Both can be listed. They apply only while the application's organization
behavior is the pre-login prompt, which this resource does not set (the
Auth0 dashboard or Management API sets it). Organizations' availability
varies by Auth0 plan. Empty is not managed. Declare the live value when
adopting an existing application; leaving it unset on an adopted
application resets it.

- rule: {"repeated":{"items":{"string":{"in":["email","organization_name"]}}}}

### spec.defaultOrganization

`Auth0ClientFromMetadataDocumentDefaultOrganization`

default_organization is the Auth0 Organization Auth0 uses for this
application's flows when a request names none. Declare the live value
when adopting an existing application; leaving it unset on an adopted
application removes it.

### spec.defaultOrganization.organizationId

`string` · required

organization_id is the Auth0 Organization's identifier ("org_..."), from
the Auth0 dashboard (Organizations, the organization's ID) or the
Management API.

- rule: {"required":true,"string":{"prefix":"org_"}}

### spec.defaultOrganization.flows

`[]string` · required

flows are the flows that use the default organization.
"client_credentials" is the only flow Auth0 defines -- a grant Auth0 does
not give applications registered from a metadata document.

- rule: {"repeated":{"minItems":"1","items":{"string":{"in":["client_credentials"]}}}}

### spec.clientMetadata

`map<string, string>`

client_metadata is up to ten key-value pairs stored with the application
-- an owner, a ticket, an environment -- readable by Actions and through
the Management API. Keys and values are at most 255 characters; keys are
letters, digits and : , - + = _ * ? " / \ ( ) < > @ tabs and spaces.
Empty is not managed. Declare the live value when adopting an existing
application; leaving it unset on an adopted application clears it.

- rule: {"map":{"maxPairs":"10","keys":{"string":{"maxLen":"255"}},"values":{"string":{"maxLen":"255"}}}}

### spec.jwtConfiguration

`Auth0ClientFromMetadataDocumentJwtConfiguration`

jwt_configuration sets how the JSON Web Tokens Auth0 issues to this
application are signed and how long they last. Unset, Auth0 keeps its own.

### spec.jwtConfiguration.alg

`string` · optional (explicit presence)

alg is the algorithm Auth0 signs the tokens with; these applications
accept only asymmetric ones, so anyone can verify a token with the
tenant's public keys.
- "RS256": RSA with SHA-256, the one every OpenID Connect library supports
- "RS512": RSA with SHA-512
- "PS256": RSA-PSS with SHA-256; Auth0's Management API reference lists it
  as available through an add-on

- rule: {"string":{"in":["RS256","RS512","PS256"]}}

### spec.jwtConfiguration.lifetimeInSeconds

`int32` · optional (explicit presence)

lifetime_in_seconds is how long a token stays valid after it is issued
(its exp claim), in seconds -- for example 3600 for an hour.

### spec.refreshToken

`Auth0ClientFromMetadataDocumentRefreshToken`

refresh_token sets the lifetime and rotation of the refresh tokens Auth0
issues to this application (it needs the refresh_token grant). Auth0
requires them to expire. Unset, Auth0 keeps its own settings.

- rule: refresh_token needs both rotation_type (rotating or non-rotating) and expiration_type (expiring) -- Auth0 refuses a refresh-token configuration without them
- rule: refresh_token.idle_token_lifetime cannot exceed token_lifetime -- a token cannot stay alive idle longer than it lives at all

### spec.refreshToken.rotationType

`string` · optional (explicit presence)

rotation_type decides whether a refresh token is replaced each time it is
used.
- "rotating": every exchange returns a new refresh token and retires the
  old one; reusing a retired token revokes the whole family, so a stolen
  token is caught the moment either party uses it. Auth0 turns it on by
  default for public (SPA and native) third-party applications, and
  OAuth 2.1 and MCP expect it for public clients.
- "non-rotating": the same refresh token is used until it expires

- rule: {"string":{"in":["rotating","non-rotating"]}}

### spec.refreshToken.expirationType

`string` · optional (explicit presence)

expiration_type decides whether refresh tokens expire. Applications
registered from a metadata document must use "expiring".

- rule: {"string":{"in":["expiring"]}}

### spec.refreshToken.leeway

`int32` · optional (explicit presence)

leeway is how many seconds a rotating refresh token may be used again
after its exchange without being treated as reuse, for a client that
retries a request the network lost (for example 3). 0 turns the overlap
off.

- rule: {"int32":{"gte":0}}

### spec.refreshToken.tokenLifetime

`int32` · optional (explicit presence)

token_lifetime is the absolute lifetime of a refresh token in seconds,
after which the person signs in again however active they are (for
example 2592000 for 30 days); up to 157788000. Rotation does not extend
it.

- rule: {"int32":{"lte":157788000,"gte":1}}

### spec.refreshToken.infiniteTokenLifetime

`bool` · optional (explicit presence)

infinite_token_lifetime removes the absolute lifetime (token_lifetime
then no longer applies). Set it false with a token_lifetime.

### spec.refreshToken.idleTokenLifetime

`int32` · optional (explicit presence)

idle_token_lifetime is how many seconds a refresh token survives unused
(for example 1296000 for 15 days); each use renews it. It cannot exceed
token_lifetime.

- rule: {"int32":{"gte":1}}

### spec.refreshToken.infiniteIdleTokenLifetime

`bool` · optional (explicit presence)

infinite_idle_token_lifetime would let an unused refresh token live
forever; applications registered from a metadata document must keep it
false.

- rule: refresh_token.infinite_idle_token_lifetime can only be false -- Auth0 requires refresh tokens of applications registered from a metadata document to expire when unused

### spec.tokenQuota

`Auth0ClientFromMetadataDocumentTokenQuota`

token_quota caps how many client-credentials tokens Auth0 issues to this
application per hour and per day. Auth0 does not give applications
registered from a metadata document the client_credentials grant, so the
quota has nothing to count until it does. Fine-grained token quotas are
in Early Access: a feature Auth0 enables on request. Declare the live value
when adopting an existing application; leaving it unset on an adopted
application removes it.

### spec.tokenQuota.clientCredentials

`Auth0ClientFromMetadataDocumentTokenQuotaClientCredentials` · required

client_credentials is the quota on client-credentials tokens.

- rule: {"required":true}

### spec.tokenQuota.clientCredentials.enforce

`bool` · optional (explicit presence)

enforce decides what happens past a limit: true refuses the request
(HTTP 429), false only logs warnings at 60, 80 and 100 percent -- a way to
watch consumption before enforcing. Unset, the provider sends true.

### spec.tokenQuota.clientCredentials.perDay

`int32` · optional (explicit presence)

per_day is the most tokens Auth0 issues in a UTC day.

- rule: {"int32":{"gte":1}}

### spec.tokenQuota.clientCredentials.perHour

`int32` · optional (explicit presence)

per_hour is the most tokens Auth0 issues in a UTC hour.

- rule: {"int32":{"gte":1}}

## Outputs

Reference an output from another manifest as `valueFrom: {kind: Auth0ClientFromMetadataDocument, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.client_id` | `string` | client_id is the application's id in the Management API ("tpc_...", the prefix of every third-party application). A client grant authorizing the application for an API, and an Auth0Connection's enabled_clients, name the application by it. In sign-in flows the application presents external_client_id instead. |
| `status.outputs.external_client_id` | `string` | external_client_id is the metadata document's URL: the client_id the application sends to /authorize and /oauth/token, and the value its access tokens carry as client_id. |
| `status.outputs.name` | `string` | name is the application's name, from the document's client_name: what Auth0 shows people on the consent screen. |
| `status.outputs.app_type` | `string` | app_type is the application type Auth0 holds ("native", "regular_web" or "spa"): the spec's app_type when set, otherwise the document's. |
| `status.outputs.grant_types` | `[]string` | grant_types are the grants Auth0 holds for the application: the spec's when set, otherwise the document's supported ones. |
| `status.outputs.callbacks` | `[]string` | callbacks are the redirect URIs Auth0 sends people back to, from the document's redirect_uris. |
| `status.outputs.logo_uri` | `string` | logo_uri is the logo Auth0 shows on the consent screen, from the document. |
| `status.outputs.jwks_uri` | `string` | jwks_uri is where the application publishes the public keys it signs private_key_jwt assertions with, from the document; empty for a public client. |
| `status.outputs.third_party_security_mode` | `string` | third_party_security_mode is the security mode Auth0 registered the application in: always "strict" for a registration from a metadata document. |
| `status.outputs.external_metadata_created_by` | `string` | external_metadata_created_by is who registered the application: "admin" (through the Management API, as this resource does) or "client" (the application registered itself). |
| `status.outputs.validation_valid` | `bool` | validation_valid is whether the document, as Auth0 fetched it at the last read, passes Auth0's validation. |
| `status.outputs.validation_warnings` | `[]string` | validation_warnings are what Auth0 ignored in the document at the last read -- a property it does not support, a grant type it filtered out. |
| `status.outputs.validation_violations` | `[]string` | validation_violations are what prevents Auth0 from processing the document fully as it stands at the last read -- for the document's owner to fix before the next fetch (external_client_id_version). |

## See Also

- [Overview](../README.md)
