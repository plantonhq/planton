# CloudflareListItem

Write a single entry into a Cloudflare List as a first-class, independently-owned
resource. Each item has its own lifecycle, so a list can be grown or trimmed one
entry at a time without rewriting the whole set.

## When to use

- Adding IPs/CIDRs, ASNs, hostnames, or redirects to a `CloudflareList`.
- Managing large or dynamically-changing lists where rewriting the full set on
  every change is undesirable.

The entry's shape must match the parent list's `kind`:

| List kind | Item field |
|---|---|
| `ip` | `ip` |
| `asn` | `asn` |
| `hostname` | `hostname` |
| `redirect` | `redirect` |

## Quick start

```yaml
apiVersion: cloudflare.planton.dev/v1
kind: CloudflareListItem
metadata:
  name: office-ip
spec:
  accountId: 0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d
  listId:
    valueFrom:
      kind: CloudflareList
      name: office-allowlist
      fieldPath: status.outputs.list_id
  ip: 203.0.113.10/32
```

## Configuration reference

| Field | Required | Description |
|---|---|---|
| `accountId` | yes | 32-char Cloudflare account ID |
| `listId` | yes | List ID, or a reference to a `CloudflareList` |
| `ip` | one-of | IPv4/IPv6 address or CIDR |
| `asn` | one-of | Non-negative 32-bit ASN |
| `hostname` | one-of | `{ urlHostname, excludeExactHostname? }` |
| `redirect` | one-of | `{ sourceUrl, targetUrl, statusCode?, includeSubdomains?, preservePathSuffix?, preserveQueryString?, subpathMatching? }` |
| `comment` | no | Informative summary |

Exactly one of `ip`/`asn`/`hostname`/`redirect` must be set. For wildcard
hostnames, `excludeExactHostname` is required; for non-wildcard hostnames it must
be omitted. Item values are immutable — changing one replaces the entry.

> Single-IP entries: write the bare address (e.g. `203.0.113.7`), not `/32`.
> Cloudflare normalizes a `/32` (and IPv6 `/128`) to a bare address, after which
> the provider's post-create read-back cannot match the submitted value and the
> apply errors with "list item pagination did not return a matching list item"
> even though the entry was created. Use the bare address or a wider CIDR.

## Outputs

| Output | Description |
|---|---|
| `item_id` | The list item's identifier |
| `list_id` | The list the entry was written to |

## Related components

- `CloudflareList` — the container this entry belongs to.
- `CloudflareRuleset` — references the list (by name) from rule expressions.
