# Cloudflare Zero Trust Organization

## Overview

`CloudflareZeroTrustOrganization` configures the Zero Trust organization: the account-wide (or zone-wide) Access login experience -- the team domain, login page design, session defaults, MFA policy, and the Access service-key rotation cadence.

Two hard provider facts shape this resource: it is a SINGLETON UPSERT (Cloudflare has no create call -- applying mutates whatever organization the account already carries, born when Zero Trust was first enabled), and DESTROY IS A NO-OP (deleting the resource abandons the live configuration exactly as last applied).

## Key Features

- **The login experience in code** -- team domain, page design, session durations, WARP authentication, unmatched-request denial
- **Organization-wide MFA policy** -- allowed authenticators, MFA sessions, hardware PIV key requirements
- **Folded key rotation** -- the Access service-key rotation cadence (21-365 days) rides the same resource
- **Dashboard lock** -- `is_ui_read_only` makes IaC the only write path

## Use Cases

**Ideal for:**

- Managing the Access login page and session policy as code alongside applications and policies
- Enforcing an organization-wide MFA posture (e.g. security keys only)
- Locking the Zero Trust dashboard read-only so IaC stays the source of truth

**Not ideal for:**

- Creating Zero Trust itself -- the organization is born at Zero Trust onboarding (the team-name step), never by this resource
- Per-application session policy -- that lives on `CloudflareZeroTrustAccessPolicy`

## API Specification

### Scope (exactly one)

| Field | Type | Description |
|-------|------|-------------|
| `account_id` | string | The account-wide organization (the common case). |
| `zone_id` | string/ref | A zone-scoped organization. NOTE: cannot be adopted by import. |

### Key Fields

| Field | Type | Description |
|-------|------|-------------|
| `auth_domain` | string | The FULL team domain, including `.cloudflareaccess.com` (required on every write). |
| `name` | string | The display name on the login page. |
| `session_duration` | string | Access session lifetime (Go duration, e.g. `24h`). |
| `login_design` | object | Colors, logo, header/footer of the login page. |
| `mfa_config` | object | Allowed authenticators + MFA session durations. |
| `mfa_ssh_piv_key_requirements` | object | Hardware PIV key constraints (pin/touch policy, key types/sizes). |
| `key_rotation_interval_days` | int | The folded Access service-key rotation cadence (21-365; account scope only). |

### Outputs

| Field | Description |
|-------|-------------|
| `auth_domain` | The team domain -- what Access applications and WARP enrollment reference |
| `account_id` | The account the organization was applied to (the singleton's identity) |

## Example Manifest

```yaml
apiVersion: cloudflare.planton.dev/v1alpha1
kind: CloudflareZeroTrustOrganization
metadata:
  name: acme-zero-trust-org
spec:
  account_id: "0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d"
  auth_domain: acme.cloudflareaccess.com
  name: Acme Zero Trust
  session_duration: 24h
  login_design:
    background_color: "#0f172a"
    header_text: Sign in to Acme
  key_rotation_interval_days: 90
```

## Destroy Semantics

Destroy is a NO-OP: nothing is reverted at Cloudflare, and the organization keeps the last-applied configuration. The same applies to the folded key-rotation cadence.

## Related Resources

- **CloudflareZeroTrustAccessIdentityProvider** -- how users sign in to this organization
- **CloudflareZeroTrustAccessApplication** -- the applications behind the login page
- **CloudflareZeroTrustAccessPolicy** -- per-application session and access policy

## Further Reading

For operational judgment -- the auth_domain blast radius, the no-op destroy discipline, the MFA/PIV pairing rules -- see GUIDE.md.

## References

- [Cloudflare Zero Trust organizations](https://developers.cloudflare.com/cloudflare-one/setup/)

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
