# CloudflareList — Research & Design Notes

## What it is

A Cloudflare List is an account-scoped, named collection that rule expressions
reference by name. Cloudflare supports four list types, each accepting a single
item shape:

- `ip` — IPv4/IPv6 addresses and CIDRs. The most common type; used in WAF and
  custom rules (`ip.src in $name`).
- `asn` — Autonomous System Numbers.
- `hostname` — hostnames, optionally wildcarded.
- `redirect` — Bulk Redirect entries (source URL → target URL), resolved by a
  redirect ruleset via `from_list`.

The list itself carries only identity (`kind`, `name`, optional `description`).
Its entries live in a separate API and are modeled here as `CloudflareListItem`.

## Why the list and its items are separate kinds

Cloudflare exposes two ways to manage items: an inline `items` block on the list
resource (which overwrites the entire list on every change) and a standalone
list-item resource. The provider explicitly warns against using both for the same
list. We deliberately expose only the standalone item model (`CloudflareListItem`),
mirroring `CloudflareKvNamespace` + `CloudflareWorkersKvPair`:

- A list is a referenceable node (rules point at it by name); an item is an
  independently-owned entry with its own lifecycle.
- Large lists (threat feeds, bulk redirects) are managed incrementally without
  re-writing the whole set on every change.
- It avoids the "two sources of truth for one list" hazard the provider warns
  about.

## Immutability

`kind` and `name` are `RequiresReplace` in the provider: changing either destroys
and recreates the list (and orphans rule references until they are repointed). The
spec documents this and constrains `name` to an expression-safe identifier.

## Composition

- Outputs `list_id` (referenced by `CloudflareListItem.list_id`) and `name`
  (referenced from rule expressions / `CloudflareRuleset` `from_list`).
- A typical infra chart: `CloudflareList` (layer 0) → N `CloudflareListItem`
  (layer 1) → `CloudflareRuleset` consuming the list by name (layer 2).

## Validation

- `account_id`: 32 hex characters.
- `kind`: required, one of the four enum values.
- `name`: required, ≤50 chars, `^[a-zA-Z][a-zA-Z0-9_]*$` (expression-safe).
- `description`: ≤500 chars.

## Engine parity

Terraform `cloudflare_list` and Pulumi `cloudflare.List` are configured from the
same fields with no inline items on either side; outputs map identically to
`list_id` / `name` / `kind`.
