---
display_name: IP Entry
---

# IP List Entry

Add a single IP/CIDR to an `ip`-kind `CloudflareList`, wired to the list by
reference so it composes in an Infra Chart.

## When to use

- Adding a trusted (or blocked) address/range to a maintained IP list.

## Key choices

- `listId`: reference a `CloudflareList` by name so the item is created after the
  list and the dependency is explicit. A literal list ID also works.
- `ip`: an IPv4/IPv6 address or CIDR (e.g. `203.0.113.10/32`).

## Placeholders

| Placeholder | Description |
|---|---|
| `0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d` | 32-character Cloudflare account ID |
| `<list-name>` | Name of the CloudflareList to add to |
| `<ip-or-cidr>` | IPv4/IPv6 address or CIDR |
| `<comment>` | Optional informative summary |
