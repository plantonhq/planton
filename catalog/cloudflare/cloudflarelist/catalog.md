# Cloudflare List

Deploys a reusable, named Cloudflare List -- a collection you reference from rule expressions such as a WAF rule's `ip.src in $office_ips` or a Bulk Redirect ruleset's `from_list`. A list fixes a single entry shape at creation (IP/CIDR, redirect, hostname, or ASN); its entries are managed independently as List Item Infra Components, so one list can hold a handful of curated values or a large, separately-owned set. Lists are account-scoped, and an empty list is a valid, referenceable object -- the right shape for a container you fill later.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **List** -- an account-scoped, named collection of the chosen kind (`ip`, `redirect`, `hostname`, or `asn`), created empty and ready for entries

## Before You Deploy

### Planton Setup

- **Cloudflare Provider Connection** -- an active connection in the Connect module with a Cloudflare API token that has Account Filter Lists edit access. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline API token authentication.

### Cloudflare Account

- **Account-level access** -- lists are created at the account level (not per-zone), so the API token must be scoped to the account.

## Deploy

### Console

Open the deployment store, find **Cloudflare List**, and click **Deploy**. The creation wizard captures the owning account, the list type, an expression-friendly name, and an optional description. Start from the **IP Allowlist** preset in the [Presets](#presets) tab for a typical firewall list.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: cloudflare.planton.dev/v1alpha1
kind: CloudflareList
metadata:
  name: office-ips
  org: acme-corp
  env: prod
spec:
  accountId: "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4"
  kind: ip
  name: office_ips
  description: Approved office egress IPs
```

```shell
planton apply -f cloudflare-list.yaml
```

This creates an empty IP list named `office_ips`, ready for entries and referenceable from rules as `$office_ips`. An Infra Job tracks the provisioning in real time.

## Key Configuration

These are the most important decisions when configuring a list. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Kind (`kind`)** -- a create-time contract fixing which entry shape the list accepts: `ip` (addresses/CIDRs), `redirect` (Bulk Redirect rules), `hostname` (optionally wildcarded), or `asn` (Autonomous System Numbers). Immutable -- changing it replaces the list. Pick the kind from the consumer, not the data: a redirect list is what Bulk Redirect rulesets consume via `from_list`; an ip list is what WAF/custom rules consume via `ip.src in $name`.

**Name (`name`)** -- the identifier used in filter and rule expressions, written as `$name`. That is why hyphens are forbidden (letters, digits, underscores only) and why the name is immutable: rules reference it literally, so renaming replaces the list -- and every item has to be rewritten. Pick a short, stable, lowercase identifier before the first apply.

**Entries never live on the list itself** -- create the list empty and add `CloudflareListItem` resources, one per entry with an independent lifecycle. Inline items and List Item resources are competing writers on the same collection; mixing them makes the provider fight itself and lose entries.

## Outputs and Dependencies

### What This Kind Consumes

This kind has no foreign-key dependencies -- a list is defined entirely by its own account, kind, and name.

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `list_id` | The Cloudflare-assigned identifier of the list | Referenced by a CloudflareListItem's `listId` field |
| `name` | The list name -- the identifier used in rule expressions as `$name` | Referenced by a CloudflareRuleset Bulk Redirect rule's `from_list` and by WAF expressions |

## Common Patterns

**IP allowlist** -- an `ip` list referenced from a WAF or firewall rule to allow (or block) a managed set of addresses: office egress IPs, partner ranges, known-bad CIDRs. Start from the **IP Allowlist** preset.

**Bulk redirect list** -- a `redirect` list a Bulk Redirect ruleset points at via `from_list`, to migrate large numbers of legacy URLs in one place. Start from the **Bulk Redirect List** preset.

**Container-first rollout** -- create the empty list before the first rule or item exists, so rules can reference `$name` from day one and entries arrive on their own lifecycle.

## Works With

- [**Cloudflare List Item**](/infra-catalog/cloudflare-list-item) -- adds entries to this list (via `listId`), one independently-managed entry at a time
- [**Cloudflare Ruleset**](/infra-catalog/cloudflare-ruleset) -- a Bulk Redirect rule references a redirect list by name (via `from_list`); WAF custom rules match against `$name`
