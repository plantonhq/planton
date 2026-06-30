---
title: "Origin CA Certificate"
description: "Origin CA Certificate deployment documentation"
icon: "package"
order: 100
componentName: "cloudflareorigincacertificate"
---

# Cloudflare Origin CA Certificate

Issue a free Cloudflare Origin CA certificate to encrypt the connection between
Cloudflare's edge and your origin server (the "Full (Strict)" SSL mode).

## What Gets Created

- A `cloudflare_origin_ca_certificate` valid for the requested hostnames.
- When no CSR is supplied: a generated private key + CSR (via the `tls` provider),
  with the key exported as a sensitive output.

## Prerequisites

- A Cloudflare API token with `SSL and Certificates` permission (the deprecated
  Origin CA Key is not required).

## Configuration Reference

**Required**

- `hostnames` — the SANs the certificate covers (e.g. the zone apex and a wildcard).

**Optional**

- `requestType` — `origin-rsa` (default), `origin-ecc`, or `keyless-certificate`.
- `requestedValidity` — 7, 30, 90, 365, 730, 1095, or 5475 days (default 5475).
- `csr` — supply your own CSR to keep your key private (no key is generated).

## Stack Outputs

| Output | Description |
|---|---|
| `certificate_id` | The certificate identifier |
| `certificate` | The issued certificate (PEM) |
| `private_key` | The generated private key (PEM, sensitive); empty if a CSR was supplied |
| `expires_on` | Expiry timestamp |

## Related Components

- `CloudflareDnsRecord`, `CloudflareDnsZone`
