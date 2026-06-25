# CloudflareZeroTrustAccessGroup

Provision a reusable Cloudflare Zero Trust **Access group** — a named bundle of
access rules (the same `include` / `exclude` / `require` building blocks an Access
policy uses). Factor shared membership criteria (an engineering team, a set of
corporate email domains, a country allow-list, an IdP group) into a group once,
then reference it from many policies — or from other groups — by ID.

## Why a standalone group

In Cloudflare's model, an Access group is a first-class, reusable object. Defining
it as its own resource (instead of repeating the same rules inside every policy)
means the criteria live in one place and every policy that references the group
updates automatically when the group changes.

## Scope

An Access group is **account-scoped** (the common case — reusable across every
application in the account) or **zone-scoped**. Set exactly one of `accountId` or
`zoneId`.

## Requirements

- **API token**: requires **Account → Access: Organizations, Identity Providers,
  and Groups → Edit** (account-scoped groups), and the equivalent zone permission
  for zone-scoped groups.

## Quick start

```yaml
apiVersion: cloudflare.openmcf.org/v1
kind: CloudflareZeroTrustAccessGroup
metadata:
  name: engineering-team
spec:
  accountId: 0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d
  name: engineering-team
  include:
    - emailDomain:
        domain: example.com
  require:
    - geo:
        countryCode: US
  exclude:
    - email:
        email: contractor@example.com
```

## Access rules

`include`, `exclude`, and `require` are lists of rules. Each rule sets exactly one
variant (its "type"):

- Identity: `email`, `emailDomain`, `emailList`, `everyone`, `group`,
  `azureAd`, `githubOrganization`, `gsuite`, `okta`, `saml`, `oidc`,
  `authContext`, `loginMethod`, `cloudflareAccountMember`.
- Network / device: `ip`, `ipList`, `geo`, `devicePosture`, `certificate`,
  `commonName`.
- Service / tokens: `serviceToken`, `anyValidServiceToken`, `linkedAppToken`.
- Risk / external: `userRiskScore`, `externalEvaluation`, `authMethod`.

Logic: a user matches if they satisfy **any** `include` rule, are not caught by
**any** `exclude` rule, and satisfy **all** `require` rules.

## Composition

- `group` rules reference another `CloudflareZeroTrustAccessGroup` by ID
  (`status.outputs.group_id`) — compose groups out of other groups.
- A `CloudflareZeroTrustAccessPolicy` references this group via a `group` rule.

## Outputs

| Output | Description |
|---|---|
| `group_id` | The Access group ID (reference it from a policy or another group) |

## Related components

- `CloudflareZeroTrustAccessPolicy` — references groups in its rules.
- `CloudflareZeroTrustAccessApplication` — binds policies to a protected resource.
