# Cloudflare Zero Trust Access Service Token

## Overview

`CloudflareZeroTrustAccessServiceToken` is a machine credential -- a client-ID / client-secret pair that non-human clients present in the `CF-Access-Client-ID` / `CF-Access-Client-Secret` request headers to pass through Access-protected applications without an identity-provider login.

The secret is returned only at creation and at rotation. Cloudflare never returns it on later reads, and an imported token cannot recover it. Capture the `client_secret` output into a secret store at deploy time; a lost secret means rotating the token.

## Key Features

- **Account- or zone-scoped** -- set exactly one of `account_id` or `zone_id`
- **First-class rotation** -- increment `client_secret_version` to mint a new secret; `previous_client_secret_expires_at` controls how long the old secret keeps working
- **Both-or-neither rotation pair** -- setting a version without an expiry (or the reverse) is rejected; leaving both unset is the normal non-rotating state (Cloudflare treats the initial secret as version 1)
- **Duration** -- Go-style duration (`8760h`) or `forever`; empty means Cloudflare's one-year default

## Use Cases

**Ideal for:**

- CI jobs, deploy bots, and service-to-service calls into Access-protected apps
- Rotating a compromised or aging secret without deleting the token (so policy `service_token` rules keep matching the same ID)
- A zone-scoped token that only unlocks applications on one zone

**Not ideal for:**

- Human sign-in -- that is `CloudflareZeroTrustAccessIdentityProvider`
- Deciding which applications accept the token -- that is an Access policy `service_token` rule

## API Specification

### Required Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `account_id` XOR `zone_id` | string / StringValueOrRef | Yes | Exactly one. `zone_id` accepts a literal or a reference to a `CloudflareDnsZone`. |
| `name` | string | Yes | Display name. |

### Optional Fields

| Field | Type | Description |
|-------|------|-------------|
| `duration` | string | Go-style duration or `forever`. Empty = 8760h (one year). |
| `client_secret_version` | optional int32 | Increment to rotate. Must be set together with `previous_client_secret_expires_at`. |
| `previous_client_secret_expires_at` | string | RFC3339. When the previous secret stops being accepted after a rotation. |

### Outputs

| Field | Description |
|-------|-------------|
| `service_token_id` | UUID used for import and policy `service_token` rules |
| `client_id` | Presented in `CF-Access-Client-ID` |
| `client_secret` | Sensitive. Returned only at create and rotation |
| `expires_at` | RFC3339 expiry |

## Example Manifests

Minimal (one-year default):

```yaml
apiVersion: cloudflare.planton.dev/v1alpha1
kind: CloudflareZeroTrustAccessServiceToken
metadata:
  name: ci-deployer
spec:
  account_id: "0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d"
  name: ci-deployer
```

Rotation (mint version 2, keep the old secret for 30 days):

```yaml
apiVersion: cloudflare.planton.dev/v1alpha1
kind: CloudflareZeroTrustAccessServiceToken
metadata:
  name: ci-deployer
spec:
  account_id: "0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d"
  name: ci-deployer
  client_secret_version: 2
  previous_client_secret_expires_at: "2026-09-15T00:00:00Z"
```

## Destroy Semantics

Destroy is a real delete. Clients holding the secret fail immediately. Prefer rotation when you need a new secret without breaking policy references.

## Related Resources

- **CloudflareZeroTrustAccessApplication** / **CloudflareZeroTrustAccessPolicy** -- attach the token via a `service_token` rule
- **CloudflareZeroTrustAccessIdentityProvider** -- human sign-in, not machine credentials
- **CloudflareDnsZone** -- `zone_id` foreign key for a zone-scoped token

## Further Reading

For operational judgment -- secret capture, rotation semantics, and the API-token-on-rotation caution -- see GUIDE.md.

## References

- [Cloudflare Access service tokens](https://developers.cloudflare.com/cloudflare-one/identity/service-tokens/)

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
