---
title: "Uploaded Certificate"
description: "This preset uploads a user-provided TLS certificate and private key to Hetzner Cloud. You supply PEM-encoded files, and Hetzner Cloud stores them for use by load balancer HTTPS services. Unlike the..."
type: "preset"
rank: "02"
presetSlug: "02-uploaded-certificate"
componentSlug: "certificate"
componentTitle: "Certificate"
provider: "hetznercloud"
icon: "package"
order: 2
---

# Uploaded Certificate

This preset uploads a user-provided TLS certificate and private key to Hetzner Cloud. You supply PEM-encoded files, and Hetzner Cloud stores them for use by load balancer HTTPS services. Unlike the managed variant, you are fully responsible for certificate renewal -- Hetzner Cloud does not monitor expiry or auto-renew uploaded certificates.

Both fields are immutable: changing either the certificate chain or private key forces replacement of the entire certificate resource in Hetzner Cloud.

## When to Use

- You have an existing certificate from a commercial CA (DigiCert, Sectigo, etc.) or an internal enterprise CA
- You need a wildcard certificate (e.g., `*.example.com`) -- Let's Encrypt HTTP-01 validation cannot issue wildcards
- You need an Extended Validation (EV) or Organization Validation (OV) certificate that Let's Encrypt does not offer
- Compliance requirements mandate a specific CA or certificate lifecycle process

## Key Configuration Choices

- **Uploaded variant** (`uploaded`) -- full control over the certificate material; no dependency on Let's Encrypt or ACME challenges
- **Certificate chain** (`certificate`) -- must include the server certificate first, followed by intermediate CA certificates in order, with the root CA last; incomplete chains cause TLS errors in clients
- **Private key as secret** (`privateKey`) -- contains sensitive cryptographic material; IaC modules treat this as a Pulumi secret / Terraform sensitive variable

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| ----------- | ----------- | ------------- |
| `<pem-certificate-chain>` | PEM-encoded certificate chain (server cert + intermediates + root, in order) | Your CA's issuance portal, or `fullchain.pem` from tools like certbot |
| `<pem-private-key>` | PEM-encoded private key corresponding to the certificate | Generated during CSR creation, or `privkey.pem` from tools like certbot |

## Related Presets

- **01-managed-lets-encrypt** -- use instead for automated TLS with zero renewal overhead, when you control DNS and do not need a specific CA
