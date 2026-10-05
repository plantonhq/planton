# Cloudflare Zero Trust Access Group

Deploys a reusable Cloudflare Zero Trust Access group: a named bundle of membership rules (an engineering team, a set of corporate email domains, a country allow-list) that many Access policies and other groups can reference. Factoring shared membership criteria into a group keeps policies small and lets the criteria evolve in one place. Groups are account- or zone-scoped and have an independent lifecycle, so one group can serve every policy that needs the same audience.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Access Group** -- a reusable group whose membership is decided by three rule lists: `include` (OR), `exclude` (NOT, wins over include), and `require` (AND), each composed of any of the 26 Cloudflare rule criteria

## Before You Deploy

### Planton Setup

- **Cloudflare Provider Connection** -- an active connection in the Connect module with a Cloudflare API token that has Access (Zero Trust) edit permission. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline API token authentication.

### Cloudflare Account

- **Zero Trust enabled** -- the account must have Cloudflare Zero Trust (Access) set up.
- **Identity providers (optional)** -- IdP-backed rules (Okta, Azure AD, GitHub, SAML, OIDC, ...) require the corresponding identity provider configured in Zero Trust; supply its identity-provider ID in the rule.

## Deploy

### Console

Open the deployment store, find **Cloudflare Zero Trust Access Group**, and click **Deploy**. The creation wizard walks you through scope (account or zone) and name, then the include / exclude / require rule builders. Start from the **Engineering team group** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: cloudflare.planton.dev/v1alpha1
kind: CloudflareZeroTrustAccessGroup
metadata:
  name: engineering
  org: acme-corp
  env: prod
spec:
  accountId: "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4"
  name: engineering
  include:
    - emailDomain:
        domain: acme.com
  require:
    - geo:
        countryCode: US
```

```shell
planton apply -f cloudflare-zero-trust-access-group.yaml
```

This creates a group matching anyone with an `acme.com` email, restricted to the US. An Infra Job tracks the provisioning in real time.

### InfraChart

Compose a group and the policies that reference it; the policy's `group` rule points at the group via ValueFromRef:

```yaml
spec:
  include:
    - group:
        id:
          valueFrom:
            kind: CloudflareZeroTrustAccessGroup
            name: engineering
            fieldPath: status.outputs.group_id
```

The InfraPipeline resolves the dependency graph, deploys the group first, then the policy with the resolved group ID.

## Key Configuration

These are the most important decisions when configuring a group. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Scope** -- A group is `account_id`-scoped (reusable across every application in the account -- the common case) or `zone_id`-scoped (limited to one zone). Exactly one is set.

**Include rules** -- The OR list; a user matches the group if they satisfy ANY include rule. At least one is required. Each rule is one of 26 criteria: email, email domain, IP/CIDR, country, another Access group, an IdP group (Okta/Azure AD/GitHub/SAML/OIDC), a service token, a device-posture check, a user-risk level, and more.

**Exclude rules** -- The NOT list; matching ANY exclude rule removes the user, and exclude wins over include.

**Require rules** -- The AND list; a user must satisfy EVERY require rule (e.g. a country AND a device-posture check).

**Default group** -- When enabled, the group is automatically added to the membership of newly created applications. Use sparingly.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **CloudflareDnsZone** (zone-scoped only) | `zoneId` | `status.outputs.zone_id` |
| **CloudflareZeroTrustAccessGroup** (group-of-groups) | `include[].group.id` | `status.outputs.group_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `group_id` | The Cloudflare-assigned identifier of the group | Referenced by a CloudflareZeroTrustAccessPolicy's `group` rule, or another group's `group` rule |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Team by email domain** -- Include an email-domain rule; reuse the group across every internal application. Start from the **Engineering team group** preset.

**IdP group with enforced MFA** -- Include an identity-provider group (Okta, Azure AD, GitHub) and require the `mfa` auth method, so membership stays managed in the IdP while Cloudflare enforces the login strength. Start from the **IdP group with MFA login method** preset.

**Group-of-groups** -- Include several team groups to compose a broader audience without re-listing members.

## Works With

- [**Cloudflare Zero Trust Access Policy**](/infra-catalog/cloudflare-zero-trust-access-policy) -- references this group via its `group` rule
- [**Cloudflare Zero Trust Access Application**](/infra-catalog/cloudflare-zero-trust-access-application) -- attaches policies that reference this group
