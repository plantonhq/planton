# CloudflareZeroTrustAccessPolicy

Provision a reusable Cloudflare Zero Trust **Access policy** — a named, account-scoped
decision (`allow` / `deny` / `non_identity` / `bypass`) plus the `include` /
`exclude` / `require` rules that decide who it applies to. Policies are standalone
and attached to one or more Access applications by ID, so the same policy can guard
many applications and its rules evolve in one place.

## Why a standalone policy

Cloudflare v5 models Access policies as reusable, account-scoped objects that
applications reference. Defining a policy as its own resource (rather than inline on
an application) lets you reuse it across applications and manage its rules
independently.

## Requirements

- **API token**: requires **Account → Access: Apps and Policies → Edit**.

## Quick start

```yaml
apiVersion: cloudflare.openmcf.org/v1
kind: CloudflareZeroTrustAccessPolicy
metadata:
  name: allow-staff
spec:
  accountId: 0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d
  name: allow-staff
  decision: allow
  include:
    - emailDomain:
        domain: example.com
  require:
    - geo:
        countryCode: US
```

## Configuration reference

| Field | Required | Description |
|---|---|---|
| `accountId` | yes | 32-char Cloudflare account ID |
| `name` | yes | Policy display name |
| `decision` | yes | `allow`, `deny`, `non_identity`, or `bypass` |
| `include` | yes | Match if ANY rule matches (≥1 required) |
| `exclude` | no | Reject if ANY rule matches |
| `require` | no | Match only if ALL rules match |
| `sessionDuration` | no | Session lifetime (e.g. `24h`, default `24h`) |
| `approvalRequired` / `approvalGroups` | no | Approval workflow |
| `isolationRequired` | no | Require browser isolation |
| `purposeJustificationRequired` / `purposeJustificationPrompt` | no | Justification prompt |
| `connectionRules` | no | RDP clipboard constraints (infrastructure apps) |
| `mfaConfig` | no | Per-policy MFA requirements |

The access-rule variants are the same set documented on `CloudflareZeroTrustAccessGroup`.

## Composition

- `include`/`exclude`/`require` `group` rules reference a
  `CloudflareZeroTrustAccessGroup` by `status.outputs.group_id`.
- A `CloudflareZeroTrustAccessApplication` references this policy via
  `status.outputs.policy_id`.

## Outputs

| Output | Description |
|---|---|
| `policy_id` | The Access policy ID (reference it from an application) |

## Related components

- `CloudflareZeroTrustAccessGroup` — reusable rule bundles referenced here.
- `CloudflareZeroTrustAccessApplication` — binds policies to a protected resource.
