# Preset: Let's Encrypt, Apex-Only, Annual

A single-hostname (apex) certificate issued by Let's Encrypt with the longest
validity (365 days). Useful when you only need to cover the bare domain and prefer a
yearly rotation cadence.

## When to use

- Apex-only coverage with Let's Encrypt as the CA.

## Key choices

- `certificateAuthority: lets_encrypt`.
- `validityDays: 365` — the maximum.
- `hosts` lists only the apex.

## Placeholders

| Placeholder | Description |
|---|---|
| `<cloudflare-zone-id>` | The target zone's ID |
| `<domain>` | The apex domain the certificate covers |
