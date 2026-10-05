# Cloudflare Zero Trust Access Identity Provider

## Overview

`CloudflareZeroTrustAccessIdentityProvider` connects Cloudflare Access to an identity source -- Google, Okta, Azure AD, GitHub, a generic OIDC or SAML provider, or Cloudflare's own one-time PIN -- so users can sign in to Access-protected applications. The provider type is immutable at Cloudflare: changing it replaces the provider and invalidates every policy rule that referenced the old ID.

Set exactly one of `account_id` or `zone_id`. Account-scoped providers (the common case) are usable by every application in the account. The `config` surface is a union in practice: which fields apply depends on `type`, and validation rejects a mismatch before the API does.

## Key Features

- **Fifteen provider types**, using Cloudflare's own spellings (`azureAD`, `google-apps`, `onetimepin`)
- **Per-type config gating** -- Azure AD, Okta, OIDC, SAML, Centrify, OneLogin, and PingOne fields are rejected on the wrong type
- **SCIM provisioning** -- enabling SCIM mints a bearer secret returned once in the `scim_secret` output (not available for `onetimepin`)
- **Create-only SCIM secret** -- capture it at deploy; it is redacted on later reads and does not survive import
- **Read-only latch** -- `read_only` tells Cloudflare to refuse API updates and deletes until an explicit apply clears it

## Use Cases

**Ideal for:**

- Putting Google, Okta, or Azure AD in front of internal apps protected by Access
- A one-time PIN fallback for contractors who have no corporate IdP
- SAML or generic OIDC connections to an IdP Cloudflare does not first-class
- SCIM so user create/update/deprovision events land in Zero Trust without waiting for re-authentication

**Not ideal for:**

- Authorizing who may reach an app -- that is `CloudflareZeroTrustAccessPolicy`
- Machine clients -- that is `CloudflareZeroTrustAccessServiceToken`
- Creating the Zero Trust organization (team name) itself -- that is account setup, later `CloudflareZeroTrustOrganization`

## API Specification

### Required Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `account_id` XOR `zone_id` | string / StringValueOrRef | Yes | Exactly one. `zone_id` accepts a literal or a reference to a `CloudflareDnsZone`. |
| `name` | string | Yes | Shown to users on the Access login page. |
| `type` | string | Yes | Immutable. One of `onetimepin`, `azureAD`, `saml`, `centrify`, `facebook`, `github`, `google-apps`, `google`, `linkedin`, `oidc`, `okta`, `onelogin`, `pingone`, `yandex`, `cloudflare`. |

### Optional Fields

| Field | Type | Description |
|-------|------|-------------|
| `config` | message | Connection parameters. Omit for `onetimepin` (the module sends the empty object Cloudflare expects). OAuth types need `client_id` + `client_secret` (sensitive StringValueOrRef). OIDC needs `auth_url` / `certs_url` / `token_url` / `scopes`. SAML needs `issuer_url` / `sso_target_url` and kin. |
| `saml_certificate_set_id` | string | Required when `config.enable_encryption` is true. |
| `scim_config` | message | `enabled`, `identity_update_behavior` (`automatic` / `reauth` / `no_action`), `seat_deprovision`, `user_deprovision`. Forbidden when `type` is `onetimepin`. |
| `read_only` | bool | Cloudflare refuses updates and deletes while set. |

### Outputs

| Field | Description |
|-------|-------------|
| `identity_provider_id` | UUID referenced by Access policy rules |
| `scim_base_url` | Cloudflare's SCIM v2.0 endpoint (present when SCIM is enabled) |
| `scim_secret` | Sensitive. Minted once when SCIM is first enabled; never readable later |

## Example Manifests

One-time PIN (no config):

```yaml
apiVersion: cloudflare.planton.dev/v1alpha1
kind: CloudflareZeroTrustAccessIdentityProvider
metadata:
  name: otp-fallback
spec:
  account_id: "0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d"
  name: one-time-pin
  type: onetimepin
```

GitHub OAuth:

```yaml
apiVersion: cloudflare.planton.dev/v1alpha1
kind: CloudflareZeroTrustAccessIdentityProvider
metadata:
  name: github-login
spec:
  account_id: "0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d"
  name: github
  type: github
  config:
    client_id: your-github-oauth-app-id
    client_secret:
      value: your-github-oauth-app-secret
```

## Destroy Semantics

Destroy is a real delete. The type is immutable -- changing it replaces the provider (new ID) and breaks every policy that referenced the old one. The SCIM secret does not come back after destroy or import.

## Related Resources

- **CloudflareZeroTrustAccessApplication** / **CloudflareZeroTrustAccessPolicy** -- the doors and guards that consume this provider
- **CloudflareZeroTrustAccessServiceToken** -- machine credentials, not human sign-in
- **CloudflareDnsZone** -- `zone_id` foreign key for a zone-scoped provider

## Further Reading

For operational judgment -- the OTP API-token update caution, SCIM secret lifecycle, and type-immutability -- see GUIDE.md.

## References

- [Cloudflare Access identity providers](https://developers.cloudflare.com/cloudflare-one/identity/idp-integration/)

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
