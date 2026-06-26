# Preset: IP Allowlist

An `ip`-kind list to collect trusted IPs/CIDRs that WAF or custom rules reference
with `ip.src in $office_allowlist`.

## When to use

- Allowlisting office/VPN egress IPs, partner ranges, or monitoring probes.
- Any rule that should match a maintained set of addresses by name.

## Key choices

- `kind: ip` — accepts IPv4/IPv6 addresses and CIDRs (immutable).
- `name` — referenced in rule expressions; keep it short and lowercase.
- Add entries with `CloudflareListItem` (one per IP/CIDR).

## Placeholders

| Placeholder | Description |
|---|---|
| `<cloudflare-account-id>` | 32-character Cloudflare account ID |
