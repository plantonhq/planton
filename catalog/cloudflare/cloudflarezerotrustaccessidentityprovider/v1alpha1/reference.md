# CloudflareZeroTrustAccessIdentityProvider

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `cloudflare.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

CloudflareZeroTrustAccessIdentityProviderSpec defines how users sign in to
Access-protected applications: the connection between Cloudflare Zero Trust
and an identity provider (Google, Okta, Azure AD, GitHub, a generic OIDC or
SAML provider, or Cloudflare's own one-time PIN).

The provider TYPE is immutable at Cloudflare (changing it replaces the
provider, invalidating every policy rule that references the old provider
ID). The config surface is a union in practice: which fields apply depends on
the type, and the validation rules below mirror Cloudflare's own per-type
gating so a mismatch fails at validation instead of at the API.

## Example

```yaml
apiVersion: cloudflare.planton.dev/v1alpha1
kind: CloudflareZeroTrustAccessIdentityProvider
metadata:
  name: test-identity-provider
spec:
  account_id: "0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d"
  name: github-login
  type: github
  config:
    client_id: example-client-id
    client_secret:
      value: example-client-secret
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.accountId` | `string` |  |  |  |
| `spec.zoneId` | `string \| valueFrom` |  |  | CloudflareDnsZone (`status.outputs.zone_id`) |
| `spec.name` | `string` | yes |  |  |
| `spec.type` | `string` | yes |  |  |
| `spec.config` | `CloudflareZeroTrustAccessIdentityProviderConfig` |  |  |  |
| `spec.config.claims` | `[]string` |  |  |  |
| `spec.config.clientId` | `string` |  |  |  |
| `spec.config.clientSecret` | `string \| valueFrom` (sensitive) |  |  |  |
| `spec.config.emailClaimName` | `string` |  |  |  |
| `spec.config.pkceEnabled` | `bool` |  |  |  |
| `spec.config.conditionalAccessEnabled` | `bool` |  |  |  |
| `spec.config.directoryId` | `string` |  |  |  |
| `spec.config.prompt` | `string` |  |  |  |
| `spec.config.supportGroups` | `bool` |  |  |  |
| `spec.config.centrifyAccount` | `string` |  |  |  |
| `spec.config.centrifyAppId` | `string` |  |  |  |
| `spec.config.appsDomain` | `string` |  |  |  |
| `spec.config.authUrl` | `string` |  |  |  |
| `spec.config.certsUrl` | `string` |  |  |  |
| `spec.config.scopes` | `[]string` |  |  |  |
| `spec.config.tokenUrl` | `string` |  |  |  |
| `spec.config.authorizationServerId` | `string` |  |  |  |
| `spec.config.oktaAccount` | `string` |  |  |  |
| `spec.config.oneloginAccount` | `string` |  |  |  |
| `spec.config.pingEnvId` | `string` |  |  |  |
| `spec.config.attributes` | `[]string` |  |  |  |
| `spec.config.emailAttributeName` | `string` |  |  |  |
| `spec.config.enableEncryption` | `bool` |  |  |  |
| `spec.config.headerAttributes` | `[]CloudflareZeroTrustAccessIdentityProviderHeaderAttribute` |  |  |  |
| `spec.config.headerAttributes[].attributeName` | `string` |  |  |  |
| `spec.config.headerAttributes[].headerName` | `string` |  |  |  |
| `spec.config.idpPublicCerts` | `[]string` |  |  |  |
| `spec.config.issuerUrl` | `string` |  |  |  |
| `spec.config.signRequest` | `bool` |  |  |  |
| `spec.config.ssoTargetUrl` | `string` |  |  |  |
| `spec.config.restrictToAccountMembers` | `bool` |  |  |  |
| `spec.samlCertificateSetId` | `string` |  |  |  |
| `spec.scimConfig` | `CloudflareZeroTrustAccessIdentityProviderScimConfig` |  |  |  |
| `spec.scimConfig.enabled` | `bool` |  |  |  |
| `spec.scimConfig.identityUpdateBehavior` | `string` |  |  |  |
| `spec.scimConfig.seatDeprovision` | `bool` |  |  |  |
| `spec.scimConfig.userDeprovision` | `bool` |  |  |  |
| `spec.readOnly` | `bool` |  |  |  |

## Field Details

### spec.accountId

`string`

The Cloudflare account ID that owns this identity provider. Set this for an
account-scoped provider (the common case). Mutually exclusive with zone_id.

- rule: account_id must be a 32-character hex string

### spec.zoneId

`string | valueFrom`

The Cloudflare zone this provider is scoped to, as a literal zone ID or a
reference to a CloudflareDnsZone. Set this for a zone-scoped provider.
Mutually exclusive with account_id.

- references: CloudflareDnsZone (`status.outputs.zone_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: CloudflareDnsZone, name: <that resource's name>, fieldPath: status.outputs.zone_id}} -- a bare string does not parse

### spec.name

`string` · required

The name of the identity provider, shown to users on the Access login page.

- rule: {"string":{"minLen":"1"}}

### spec.type

`string` · required

The identity provider type. IMMUTABLE at Cloudflare: changing it replaces
the provider (new provider ID), invalidating policy rules that reference
the old one. The value set and casing are Cloudflare's own (azureAD and
google-apps are spelled exactly so). "onetimepin" needs no config at all
and is an ACCOUNT SINGLETON: Cloudflare refuses a second onetimepin
provider with 409 "a onetimepin connection already exists" (measured live
2026-08-26) -- adopt the existing one by import instead of creating
another. OAuth types (github, google, facebook, linkedin, yandex) need
client_id + client_secret; oidc/saml need their endpoint fields.

- rule: {"required":true,"string":{"in":["onetimepin","azureAD","saml","centrify","facebook","github","google-apps","google","linkedin","oidc","okta","onelogin","pingone","yandex","cloudflare"]}}

### spec.config

`CloudflareZeroTrustAccessIdentityProviderConfig`

The provider connection parameters. Which fields apply depends on `type`
(the message-level rules above enforce the pairing). Omit entirely for
"onetimepin" -- the module sends Cloudflare the empty config object it
expects.

### spec.config.claims

`[]string`

Custom claims to request from the identity provider and add to the signed
Access JWT (usable in policy rules).

### spec.config.clientId

`string`

The OAuth client ID issued by the identity provider for this connection.

### spec.config.clientSecret

`string | valueFrom` · sensitive

Sensitive. The OAuth client secret issued by the identity provider.
Provide a managed-secret reference; the platform resolves it just-in-time
at deploy and never stores it in plaintext.

- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.config.emailClaimName

`string`

The claim name carrying the user's email in the id_token response, when the
provider does not use the standard "email" claim.

### spec.config.pkceEnabled

`bool` · optional (explicit presence)

Enable Proof Key for Code Exchange (PKCE) on the OAuth flow.

### spec.config.conditionalAccessEnabled

`bool` · optional (explicit presence)

azureAD only: load authentication contexts (conditional access) from the
Azure AD tenant so policies can require them.

### spec.config.directoryId

`string`

azureAD only: the Azure AD directory (tenant) UUID.

### spec.config.prompt

`string`

azureAD only: the Microsoft login interaction mode. "login" forces
credential entry (negating single sign-on), "none" forbids any interactive
prompt (errors if SSO cannot complete silently), "select_account" shows the
account picker.

- rule: prompt must be login, select_account, or none

### spec.config.supportGroups

`bool` · optional (explicit presence)

azureAD only: load the user's groups from the tenant so policies can match
on them.

### spec.config.centrifyAccount

`string`

centrify only: the Centrify account URL (e.g. https://abc123.my.centrify.com).

### spec.config.centrifyAppId

`string`

centrify only: the Centrify app ID (e.g. exampleapp).

### spec.config.appsDomain

`string`

google-apps only: the Google Workspace domain (your company's TLD).

### spec.config.authUrl

`string`

oidc only: the authorization_endpoint URL of the identity provider.

### spec.config.certsUrl

`string`

oidc only: the jwks_uri endpoint serving the provider's token-signing keys.

### spec.config.scopes

`[]string`

oidc only: the OAuth scopes to request.

### spec.config.tokenUrl

`string`

oidc only: the token_endpoint URL of the identity provider.

### spec.config.authorizationServerId

`string`

okta only: the Okta authorization server ID (when using a custom
authorization server).

### spec.config.oktaAccount

`string`

okta only: the Okta account URL (e.g. https://example.okta.com).

### spec.config.oneloginAccount

`string`

onelogin only: the OneLogin account URL (e.g. https://example.onelogin.com).

### spec.config.pingEnvId

`string`

pingone only: the PingOne environment identifier.

### spec.config.attributes

`[]string`

saml only: SAML attribute names added to the signed Access JWT and usable
in policy rules.

### spec.config.emailAttributeName

`string`

saml only: the attribute name carrying the user's email in the SAML
response, when not the standard one.

### spec.config.enableEncryption

`bool` · optional (explicit presence)

saml only: encrypt SAML assertions with the certificate set assigned via
saml_certificate_set_id (which becomes required). Configure the public
certificate at the external identity provider.

### spec.config.headerAttributes

`[]CloudflareZeroTrustAccessIdentityProviderHeaderAttribute`

saml only: IdP attributes returned as headers on the request to the origin
after the Access callback.

### spec.config.headerAttributes[].attributeName

`string`

The attribute name as sent by the identity provider.

### spec.config.headerAttributes[].headerName

`string`

The header name added to the request to the origin.

### spec.config.idpPublicCerts

`[]string`

saml only: X.509 certificates used to verify the signature of the SAML
authentication response (PEM).

### spec.config.issuerUrl

`string`

saml only: the IdP Entity ID or Issuer URL.

### spec.config.signRequest

`bool` · optional (explicit presence)

saml only: sign the SAML authentication request with Access credentials
(verify with the public key from the Access certs endpoints).

### spec.config.ssoTargetUrl

`string`

saml only: the URL Access sends SAML authentication requests to.

### spec.config.restrictToAccountMembers

`bool` · optional (explicit presence)

Restrict authentication through this provider to members of the Cloudflare
account. Leave unset for Cloudflare's default behavior.

### spec.samlCertificateSetId

`string`

The UID of the SAML encryption certificate set assigned to this provider
(created out-of-band via the Access API's saml_certificate endpoint).
Required when config.enable_encryption is true.

### spec.scimConfig

`CloudflareZeroTrustAccessIdentityProviderScimConfig`

SCIM provisioning: let the identity provider push user create/update/
deprovision events to Cloudflare so Zero Trust identities stay in sync
without waiting for re-authentication. Not available for onetimepin.
Enabling SCIM mints a bearer secret exposed once in the scim_secret stack
output.

- rule: seat_deprovision requires user_deprovision

### spec.scimConfig.enabled

`bool`

Turn SCIM on. Enabling it for the first time mints the SCIM bearer secret
(exposed once in the scim_secret output; refresh it later via the
Access API's refresh_scim_secret endpoint if lost).

### spec.scimConfig.identityUpdateBehavior

`string`

How a SCIM update event affects the user's Zero Trust identity used in
policy evaluation: "automatic" updates the identity in place (and augments
it with SCIM fields), "reauth" forces re-authentication on group-membership
changes (identities do not carry SCIM fields), "no_action" leaves
identities untouched. Leave empty for Cloudflare's default (no_action).

- rule: identity_update_behavior must be automatic, reauth, or no_action

### spec.scimConfig.seatDeprovision

`bool`

Free the user's Zero Trust seat when the identity provider deprovisions
them. Requires user_deprovision.

### spec.scimConfig.userDeprovision

`bool`

Revoke the user's Access and Gateway sessions when the identity provider
deprovisions them.

### spec.readOnly

`bool`

Declares the provider immutable: Cloudflare refuses API updates and deletes
while this is set. A safety latch for foundational providers -- clearing it
requires an explicit apply.

## Validation Rules

- `spec.account_xor_zone`: set exactly one of account_id or zone_id
- `spec.config_gate_azure_ad`: conditional_access_enabled, directory_id, prompt, and support_groups apply only when type is azureAD
- `spec.config_gate_centrify`: centrify_account and centrify_app_id apply only when type is centrify
- `spec.config_gate_google_apps`: apps_domain applies only when type is google-apps
- `spec.config_gate_oidc`: auth_url, certs_url, scopes, and token_url apply only when type is oidc
- `spec.config_gate_okta`: authorization_server_id and okta_account apply only when type is okta
- `spec.config_gate_onelogin_pingone`: onelogin_account applies only when type is onelogin; ping_env_id only when type is pingone
- `spec.config_gate_saml`: SAML fields (attributes, email_attribute_name, enable_encryption, header_attributes, idp_public_certs, issuer_url, sign_request, sso_target_url) apply only when type is saml
- `spec.encryption_requires_certificate_set`: enable_encryption requires saml_certificate_set_id
- `spec.scim_not_for_onetimepin`: scim_config cannot be set when type is onetimepin

## Outputs

Reference an output from another manifest as `valueFrom: {kind: CloudflareZeroTrustAccessIdentityProvider, name: <resource-name>, fieldPath: status.outputs.<output>}`. A sensitive output is a secret the resource generates: on Planton it is kept in the organization's secret store and the output holds a `$secret/` reference, so feed it only to a sensitive field.

| Output | Type | Description |
|---|---|---|
| `status.outputs.identity_provider_id` | `string` | The UUID of the identity provider -- what Access policy rules (azure_ad, github_organization, gsuite, okta, saml, oidc, login_method, auth_context) reference. |
| `status.outputs.scim_base_url` | `string` | The base URL of Cloudflare's SCIM v2.0 endpoint for this provider (present when SCIM is enabled). Configure it at the identity provider together with scim_secret. |
| `status.outputs.scim_secret` | `string` (sensitive) | Sensitive. The SCIM bearer token minted when SCIM is first enabled. Cloudflare returns it ONLY once -- it is redacted on subsequent reads and does not survive an import (refresh it via the Access API's refresh_scim_secret endpoint if lost). Capture it into a secret store at deploy time. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.zoneId` | CloudflareDnsZone | `status.outputs.zone_id` |

## See Also

- [Overview](../README.md)
