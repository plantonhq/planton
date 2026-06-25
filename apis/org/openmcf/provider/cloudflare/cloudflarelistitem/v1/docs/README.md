# CloudflareListItem — Research & Design Notes

## What it is

A single entry in a Cloudflare List. The Cloudflare API exposes list items as
their own resource (distinct from the list's inline `items` block), which is the
model used here: one `CloudflareListItem` = one entry, with an independent
lifecycle and a foreign key to its parent list.

## Item shapes

The entry shape must match the parent list's `kind`:

- `ip` — an IPv4/IPv6 address or CIDR.
- `asn` — a non-negative 32-bit ASN.
- `hostname` — `{ url_hostname, exclude_exact_hostname? }`. For wildcard
  hostnames (`*.example.com`), `exclude_exact_hostname` is required (true = only
  subdomains; false = apex too); for non-wildcard hostnames it must be omitted.
- `redirect` — a Bulk Redirect rule (`source_url` → `target_url`) with optional
  `status_code` (301/302/307/308, default 301) and matching flags.

Modeled as a proto `oneof` with a message-level CEL rule requiring exactly one
case to be set.

## Why a separate kind (not inline items)

See `CloudflareList`'s notes: the list is the referenceable container; items are
independently-owned entries. This mirrors `CloudflareKvNamespace` +
`CloudflareWorkersKvPair`. It avoids the provider's documented "inline items vs
list_item" conflict and supports large/incrementally-managed lists.

## Immutability

Item values are `RequiresReplace` in the provider — changing an entry's value (or
its list) replaces the item.

## Engine parity

Both engines build the same single resource. Optional booleans on the redirect
shape are omitted when false (Terraform sends null; Pulumi omits the pointer), so
both rely on the provider's `false` default and produce byte-for-byte identical
plans. `status_code` is omitted when 0 so the provider applies its 301 default.

## Composition

- `list_id` is a `StringValueOrRef` defaulting to `CloudflareList`'s
  `status.outputs.list_id`, so items declared after a list resolve the dependency
  automatically.
- Outputs `item_id` and `list_id`.

## Provider quirk: single-IP entries and the `/32` read-back

Both engines use the same `cloudflare_list_item` provider resource, which has a
post-create read-back that matches the created item by paging the list and
comparing the submitted value. Cloudflare normalizes a `/32` IPv4 (and `/128`
IPv6) to a bare address, so the read-back compares `203.0.113.7/32` against the
stored `203.0.113.7`, fails to match, and the apply errors with "list item
pagination did not return a matching list item" — even though the item was
created. Guidance (documented in the user README): use the bare address for a
single IP, or a wider CIDR. This is an upstream provider behavior, not a module
defect; if a future provider release normalizes before the read-back, this note
can be dropped without any spec change.
