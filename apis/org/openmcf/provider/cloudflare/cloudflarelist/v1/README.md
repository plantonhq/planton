# CloudflareList

Provision an account-scoped Cloudflare List: a reusable, named collection of IPs,
ASNs, hostnames, or redirects that rule expressions reference by name. The list is
the container; its entries are managed as `CloudflareListItem` resources.

## When to use

- An allow/deny set of IPs or ASNs referenced from WAF or custom rules
  (`ip.src in $office_allowlist`).
- A hostname set shared across rules.
- A Bulk Redirect list (`kind: redirect`) whose entries a redirect ruleset
  resolves with `from_list`.

Use `CloudflareListItem` to add entries. A single list can hold a handful of
curated entries or a large, independently-managed set — each item has its own
lifecycle, so adding or removing one entry never rewrites the whole list.

## Quick start

```yaml
apiVersion: cloudflare.openmcf.org/v1
kind: CloudflareList
metadata:
  name: office-allowlist
spec:
  accountId: 0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d
  kind: ip
  name: office_allowlist
  description: Corporate office egress IPs
```

## Configuration reference

| Field | Required | Description |
|---|---|---|
| `accountId` | yes | 32-char Cloudflare account ID |
| `kind` | yes | `ip`, `redirect`, `hostname`, or `asn` (immutable) |
| `name` | yes | Name used in rule expressions (immutable; letters/digits/underscore, must start with a letter) |
| `description` | no | Human-readable summary |

`kind` and `name` are fixed at creation; changing either replaces the list.

## Outputs

| Output | Description |
|---|---|
| `list_id` | The Cloudflare-assigned list ID (referenced by `CloudflareListItem`) |
| `name` | The list name (used in rule expressions) |
| `kind` | The list kind |

## Related components

- `CloudflareListItem` — an entry written into this list.
- `CloudflareRuleset` — references lists from rule expressions and `from_list`.
