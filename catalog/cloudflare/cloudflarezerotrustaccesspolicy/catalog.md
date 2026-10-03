# Cloudflare Zero Trust Access Policy

Deploys a reusable Cloudflare Zero Trust Access policy: a decision (allow / deny / non-identity / bypass) plus the include / exclude / require rules that decide which requests it applies to, with optional approval, browser-isolation, purpose-justification, RDP, and MFA controls. Policies are standalone and account-scoped: one policy can guard many Access applications, which attach it by referencing its `policy_id` output, so its rules evolve in one place.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Access Policy** -- a reusable policy with a decision, three rule lists (`include`/`exclude`/`require`) of the 26 Cloudflare rule criteria, an optional session duration, approval workflow, browser-isolation and purpose-justification flags, RDP connection rules, and per-policy MFA

## Before You Deploy

### Planton Setup

- **Cloudflare Provider Connection** -- an active connection in the Connect module with a Cloudflare API token that has Access (Zero Trust) edit permission. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline API token authentication.

### Cloudflare Account

- **Zero Trust enabled** -- the account must have Cloudflare Zero Trust (Access) set up.
- **Access groups (optional)** -- the `group` rule references a CloudflareZeroTrustAccessGroup; create one first (or reference an existing group ID) to reuse membership.

## Deploy

### Console

Open the deployment store, find **Cloudflare Zero Trust Access Policy**, and click **Deploy**. The creation wizard walks you through identity and decision, the include / exclude / require rule builders, the governance controls, and the connection/MFA options. Start from the **Allow staff** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: cloudflare.planton.dev/v1alpha1
kind: CloudflareZeroTrustAccessPolicy
metadata:
  name: engineering-allow
  org: acme-corp
  env: prod
spec:
  accountId: "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4"
  name: engineering-allow
  decision: allow
  include:
    - group:
        id:
          valueFrom:
            kind: CloudflareZeroTrustAccessGroup
            name: engineering
            fieldPath: status.outputs.group_id
  sessionDuration: 24h
```

```shell
planton apply -f cloudflare-zero-trust-access-policy.yaml
```

This creates an allow policy that grants the referenced engineering group. An Infra Job tracks the provisioning in real time.

### InfraChart

Deploy the group, the policy, and the application together; the application references the policy's `policy_id`:

```yaml
spec:
  policies:
    - policy:
        valueFrom:
          kind: CloudflareZeroTrustAccessPolicy
          name: engineering-allow
          fieldPath: status.outputs.policy_id
```

The InfraPipeline resolves the dependency graph, deploys the group and policy first, then the application with the resolved policy ID.

## Key Configuration

These are the most important decisions when configuring a policy. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Decision** -- What happens when the rules match: `allow` grants, `deny` blocks, `non_identity` permits service tokens / mTLS without a user, and `bypass` skips Access entirely (e.g. health checks). Bypass removes enforcement -- use it only for unauthenticated paths.

**Include / Exclude / Require rules** -- The same 26-criterion rule model as Access groups: include is an OR (at least one required), exclude is a NOT (wins over include), require is an AND.

**Session duration** -- How long a session is valid before re-authentication (e.g. `30m`, `24h`); `0s` forces auth on every request. Defaults to 24h.

**Governance** -- Optional approval workflow (one or more approval groups, each needing a number of approvals), browser-isolation enforcement, and a purpose-justification prompt.

**Connection & MFA** -- Optional RDP clipboard constraints (text-only or blocked) for infrastructure targets, and per-policy MFA (allowed authenticators, disable, re-prompt interval).

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **CloudflareZeroTrustAccessGroup** | `include[].group.id` | `status.outputs.group_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `policy_id` | The Cloudflare-assigned identifier of the policy | Referenced by a CloudflareZeroTrustAccessApplication's policies list |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Allow a team** -- decision `allow`, include a `group` rule referencing an Access group. Start from the **Allow staff** preset.

**Break-glass with approval** -- decision `allow`, require approval with a security approval group and per-policy MFA. Start from the **Admins with approval and MFA** preset.

**Health-check bypass** -- decision `bypass`, include an IP rule for your monitoring range.

## Works With

- [**Cloudflare Zero Trust Access Group**](/infra-catalog/cloudflare-zero-trust-access-group) -- referenced by this policy's `group` rule
- [**Cloudflare Zero Trust Access Application**](/infra-catalog/cloudflare-zero-trust-access-application) -- attaches this policy via its policies list
