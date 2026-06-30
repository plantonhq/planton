# Preset: Bulk Redirect List

A `redirect`-kind list holding source→target URL rules. A redirect ruleset
(`CloudflareRuleset`, http_request_redirect phase) resolves these with `from_list`,
enabling large-scale URL redirects managed as data.

## When to use

- Migrating URLs at scale (site relaunch, domain consolidation).
- Marketing vanity URLs that map to canonical destinations.

## Key choices

- `kind: redirect` — entries are redirect definitions (immutable).
- Add entries with `CloudflareListItem` using the `redirect` item shape
  (`sourceUrl`, `targetUrl`, optional status code and matching flags).
- Wire a `CloudflareRuleset` redirect rule's `from_list` to this list's name.

## Placeholders

| Placeholder | Description |
|---|---|
| `<cloudflare-account-id>` | 32-character Cloudflare account ID |
