# Preset: Advanced Certificate (TXT validation)

The recommended default: an advanced certificate pack covering the zone apex and a
wildcard, issued by Google Trust Services and validated via TXT. For a zone on
Cloudflare's nameservers, TXT validation completes automatically.

## When to use

- Default choice for a custom edge certificate on a Cloudflare-hosted zone.

## Key choices

- `certificateAuthority: google` — broadly compatible; switch to `lets_encrypt` or
  `ssl_com` if required.
- `validityDays: 90` — short-lived, frequently rotated.

## Placeholders

| Placeholder | Description |
|---|---|
| `<cloudflare-zone-id>` | The target zone's ID |
| `<domain>` | The apex domain the certificate covers |
