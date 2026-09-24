# TLS Server From Config

## Use Case

A 30-day server certificate whose contents -- the host name, key usage, and public key -- are reviewed in the manifest, issued through the shared TLS server template.

## When to Use

- A load balancer's or internal server's certificate you would otherwise issue by hand
- When the certificate's names should go through code review

## What This Creates

- A certificate for `api.internal.example.com` from the `internal-servers` pool, shaped by the `tls-server` template, for the public key in the manifest

## Customize

| Field | Default | Why Change |
|-------|---------|------------|
| `config.publicKey.key` | an example P-256 key | Base64 of your key's PEM public key (`filebase64("key.pub.pem")`); the private key stays with you. |
| `config.subjectConfig` | `api.internal.example.com` | Your host names. |
| `lifetime` | `2592000s` (30 days) | Capped by the pool and template anyway. |
