# Cloudflare Zero Trust List

Deploys a reusable Cloudflare Zero Trust list: a named set of values — domains, IPs, URLs, emails, device serial numbers, and kin — that Gateway policies and device-posture rules reference by ID instead of repeating the values inline. Centralizing values in a list lets many policies share one source of truth that evolves in one place. The list type is immutable at Cloudflare: changing it replaces the list with a new ID and breaks every policy that referenced the old one.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Zero Trust List** — a `cloudflare_zero_trust_list` on the account, with its items as a set (order is not significant and is not preserved)

## Before You Deploy

### Planton Setup

- **Cloudflare Provider Connection** — an active connection in the Connect module with a Cloudflare API token carrying Account → Zero Trust → Edit. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** — required when using Runner-based credential delivery. Not needed for inline API token authentication.

### Cloudflare Account

- **Zero Trust enabled** — the account must have completed Cloudflare Zero Trust onboarding (the team-name step).
- **A consumer in mind** — the list does nothing on its own; a Gateway policy `traffic`/`identity` expression or a posture rule's `input.id` gives it effect.

## Deploy

### Console

Open the deployment store, find **Cloudflare Zero Trust List**, and click **Deploy**. The creation wizard walks you through the owning account, the list type, and the item set. Start from the **Domain list** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: cloudflare.planton.dev/v1alpha1
kind: CloudflareZeroTrustList
metadata:
  name: blocked-domains
  org: acme-corp
  env: prod
spec:
  accountId: "0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d"
  name: blocked-domains
  type: DOMAIN
  description: Domains Gateway policies block for all users
  items:
    - value: gambling.example.com
    - value: casino.example.net
```

```shell
planton apply -f zt-list.yaml
```

This creates a DOMAIN list that any number of Gateway DNS policies can reference by its `list_id`. An Infra Job tracks the provisioning in real time.

## Key Configuration

These are the most important decisions when configuring a Zero Trust list. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Type is a one-way door** — a DOMAIN list cannot become an IP list in place; Cloudflare mints a new list with a new ID, and every policy or posture rule referencing the old `list_id` now points at a deleted object. The discipline: create the new list, retarget the consumers, then destroy the old one. Use the canonical uppercase form (`DOMAIN`, `IP`, `EMAIL`, …) — the API stores type uppercase, and lowercase input round-trips as permanent plan drift.

**Items are a set** — order is not significant and is not preserved; two applies that only shuffle item order are the same list. Every item requires a `value` here even though the API tolerates value-less entries — an entry with only a description matches nothing and can only be a mistake.

**URL-type lists drift at provider v5.23.0** — the API normalizes URL values (trailing slashes, scheme casing) and the provider does not, so a URL list shows a perpetual plan diff. Prefer DOMAIN or IP lists for managed configuration; if you need URL, write already-normalized values and expect the diff until the provider catches up.

**This is not CloudflareList** — `CloudflareList` is the older account-level list family consumed by Rulesets. This kind is the Zero Trust/Gateway list. They do not share IDs, APIs, or import formats: a Gateway policy cannot reference a Ruleset list, and a Ruleset cannot reference a Zero Trust list.

**Destroy breaks referencers** — destroy is a real delete, and policies referencing `list_id` start failing their list lookup. Update or delete those policies first. Emptying `items` while keeping the list is the reversible alternative when the ID must stay stable.

## Outputs and Dependencies

### What This Kind Consumes

This kind has no foreign key dependencies. It is a leaf resource other Zero Trust kinds reference by ID.

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `list_id` | The UUID of the created list | Gateway policy `traffic`/`identity` expressions; posture rule `input.id` for serial-number and client-ID checks |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Domain list** — a shared blocklist or allowlist of hostnames that two or more Gateway DNS policies match on; the usual first Zero Trust list on an account. Start from the **Domain list** preset.

**IP list** — corporate egress CIDRs to allowlist, or known-bad addresses to block, shared across HTTP and L4 policies. Start from the **IP list** preset.

**Device inventory list** — a SERIAL or DEVICE list of managed hardware identifiers, referenced from a posture rule's `input.id` so only inventoried devices pass the check.

## Works With

- [**Cloudflare Zero Trust Gateway Policy**](/infra-catalog/cloudflare-zero-trust-gateway-policy) — matches this list from `traffic` or `identity` expressions.
- [**Cloudflare Zero Trust Device Posture Rule**](/infra-catalog/cloudflare-zero-trust-device-posture-rule) — serial-number and unique-client-ID checks read a list of device identifiers via `input.id`.
- [**Cloudflare List**](/infra-catalog/cloudflare-list) — the older Ruleset-family list; a different object with different consumers. Do not mix the two.
