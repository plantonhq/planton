# Preset: Bring Your Own Certificate (Enterprise)

For Enterprise accounts that upload their own certificate and key for the custom
hostname instead of using a Cloudflare-issued DV certificate, and route it to a
specific origin.

## When to use

- You have an existing certificate for the customer hostname and an Enterprise plan.

## Key choices

- `ssl.customCertificate` / `ssl.customKey` — your PEM certificate and private key
  (the key is sensitive; resolve it as a managed secret).
- `customOriginServer` — the backend this hostname routes to.

> Note: uploaded certificates are an Enterprise-only feature.

## Placeholders

| Placeholder | Description |
|---|---|
| `<saas-zone-id>` | The SaaS zone's ID |
| `<customer-hostname>` | The customer's hostname |
| `<origin-hostname>` | The backend origin for this hostname |
| `<pem-encoded-certificate>` | Your certificate (PEM) |
| `<pem-encoded-private-key>` | Your private key (PEM) |
