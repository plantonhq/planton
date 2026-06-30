---
title: "Preset: Bring Your Own CSR"
description: "For teams that already manage their own key material. You supply a PEM-encoded CSR; the module requests the certificate for that exact CSR and generates no key, so your private key never leaves your..."
type: "preset"
rank: "02"
presetSlug: "02-byo-csr"
componentSlug: "origin-ca-certificate"
componentTitle: "Origin CA Certificate"
provider: "cloudflare"
icon: "package"
order: 2
---

# Preset: Bring Your Own CSR

For teams that already manage their own key material. You supply a PEM-encoded CSR;
the module requests the certificate for that exact CSR and generates no key, so your
private key never leaves your control and the `private_key` output is empty.

## When to use

- You have an existing private key / key-management workflow, or a Keyless SSL
  setup (`requestType: keyless-certificate`).

## Key choices

- `csr` — the PEM-encoded certificate signing request you generated.
- `requestType` — must match the key algorithm of your CSR.

## Placeholders

| Placeholder | Description |
|---|---|
| `<domain>` | A hostname covered by your CSR |
| `<pem-encoded-csr>` | Your PEM-encoded certificate signing request |
