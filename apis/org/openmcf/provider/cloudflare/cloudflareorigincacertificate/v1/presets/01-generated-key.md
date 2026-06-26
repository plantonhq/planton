# Preset: Generated Key (recommended)

The recommended default: the module generates an RSA key + CSR for your hostnames
and returns the signed certificate together with the (sensitive) private key. A
downstream origin can mount both directly.

## When to use

- Default choice for securing the Cloudflare-to-origin hop (Full (Strict) SSL).

## Key choices

- `requestType: origin-rsa` for broadest compatibility (use `origin-ecc` for a
  smaller/faster ECDSA key).
- `requestedValidity: 5475` (15 years) — Origin CA certs can be long-lived.

## Placeholders

| Placeholder | Description |
|---|---|
| `<domain>` | The apex domain the certificate covers (a wildcard SAN is added) |
