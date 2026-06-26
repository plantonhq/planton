# Preset: SaaS Fallback Origin

The single per-zone fallback origin every Cloudflare for SaaS zone needs before
custom hostnames can serve traffic. Point it at your application's backend.

## When to use

- Once per SaaS zone, as the prerequisite for `CloudflareCustomHostname`.

## Placeholders

| Placeholder | Description |
|---|---|
| `<saas-zone-id>` | The SaaS zone's ID |
| `<origin-hostname>` | The backend origin hostname (a record within the SaaS zone) |
