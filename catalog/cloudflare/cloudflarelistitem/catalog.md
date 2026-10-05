# Cloudflare List Item

Adds a single entry to a Cloudflare List. Items have independent lifecycles, so a list can grow or shrink one entry at a time without rewriting the whole set -- ideal when entries are owned by different teams or pipelines. The entry's shape (an IP/CIDR, an ASN, a hostname, or a redirect rule) must match the parent list's kind, and item values are immutable in the provider: changing a value replaces the item with a new identifier.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **List Item** -- one entry written into the referenced list, of the shape that matches the list's kind

## Before You Deploy

### Planton Setup

- **Cloudflare Provider Connection** -- an active connection in the Connect module with a Cloudflare API token that has Account Filter Lists edit access. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **A parent List** -- a CloudflareList to write into. Reference it (recommended) so Planton orders creation and renders the dependency, or supply a literal list ID for a list managed elsewhere. The parent must be an empty container: a list that declares inline items and List Item resources are competing writers that overwrite each other.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline API token authentication.

### Cloudflare Account

- **Matching list kind** -- the entry value must match the parent list's kind. Cloudflare rejects a mismatched entry (e.g. an IP value in a hostname list) at apply time, not at YAML parse time.

## Deploy

### Console

Open the deployment store, find **Cloudflare List Item**, and click **Deploy**. The creation wizard captures the owning account and parent list (with a live connection diagram), then the entry value via a type-aware editor. Start from the **IP List Entry** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: cloudflare.planton.dev/v1alpha1
kind: CloudflareListItem
metadata:
  name: office-ip-1
  org: acme-corp
  env: prod
spec:
  accountId: "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4"
  listId:
    valueFrom:
      kind: CloudflareList
      name: office-ips
      fieldPath: status.outputs.list_id
  ip: "203.0.113.0/24"
  comment: Approved office egress range
```

```shell
planton apply -f cloudflare-list-item.yaml
```

This adds the `203.0.113.0/24` CIDR to the `office-ips` list, referencing the list's output so the dependency is wired automatically. An Infra Job tracks the provisioning in real time.

### InfraChart

When the parent list is deployed in the same InfraPipeline, wire each entry to it with ValueFromRef:

```yaml
spec:
  accountId: "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4"
  listId:
    valueFrom:
      kind: CloudflareList
      name: office-ips
      fieldPath: status.outputs.list_id
  ip: "203.0.113.0/24"
```

The InfraPipeline resolves the dependency graph, creates the empty list first, then writes the entries into it.

## Key Configuration

These are the most important decisions when configuring a list item. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Parent list (`listId`)** -- the list this entry is written to. Reference a CloudflareList (resolves to its `status.outputs.list_id`) so Planton orders creation and the graph shows the edge, or paste a literal list ID. One writer per list: never mix these resources with a parent that declares inline items.

**Entry value (the `item` oneof)** -- exactly one of `ip` (an IPv4/IPv6 address or CIDR), `asn`, `hostname` (optionally wildcarded, with `excludeExactHostname` required for wildcards), or `redirect` (source/target URLs, a status code, and behavior flags). The spec lets you write any shape; the parent list's `kind` is the real constraint enforced at the API -- look at the parent before picking.

**The value is the identity** -- item values are immutable in the provider: changing an IP, ASN, hostname, or redirect replaces the item with a new id. A comment-only edit may update in place; a value edit will not. Treat the value as the key you meant to write, not something you will mutate later.

**Deletes can lag** -- a just-deleted item may still answer reads for a few seconds. Automation that checks absence immediately after destroy can false-fail; a short retry is the expected workaround.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **CloudflareList** | `listId` | `status.outputs.list_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `item_id` | The Cloudflare-assigned identifier of the list item | Verification tooling and imports -- paired with `list_id`, since an item's API identity is the (list, item) tuple |
| `list_id` | The list ID the entry was written to | Completes the item's API identity for tooling that composes on it |

## Common Patterns

**IP entry** -- a single IP/CIDR added to an `ip` list, referenced from firewall rules. Use to extend an allowlist or blocklist one address at a time. Start from the **IP List Entry** preset.

**Redirect entry** -- a source-to-target redirect added to a `redirect` list backing a Bulk Redirect ruleset. Use to migrate individual legacy URLs. Start from the **Bulk Redirect Entry** preset.

**Per-team ownership** -- give each team its own List Item resources against a shared list, so entries are added and removed on independent lifecycles without anyone rewriting the whole set.

## Works With

- [**Cloudflare List**](/infra-catalog/cloudflare-list) -- the parent container this entry is written into (via `listId`)
