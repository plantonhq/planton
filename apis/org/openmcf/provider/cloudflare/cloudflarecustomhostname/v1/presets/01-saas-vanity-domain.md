# Preset: SaaS Vanity Domain (recommended)

The recommended default: onboard a customer's hostname with a Cloudflare-issued DV
certificate validated over TXT. The customer adds a CNAME and the ownership TXT
record (from the stack outputs) and the hostname goes live.

## When to use

- The default for letting a customer use your SaaS product on their own domain.

## Key choices

- `ssl.method: txt` — TXT-based domain control validation.
- `ssl.type: dv` — domain validation (the only supported level).
- `settings.minTlsVersion: "1.2"` — a sensible security floor.

## Placeholders

| Placeholder | Description |
|---|---|
| `<saas-zone-id>` | The SaaS zone's ID |
| `<customer-hostname>` | The customer's hostname, e.g. `support.acme.com` |
