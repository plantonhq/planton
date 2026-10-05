# DigitalOcean Certificate

A DigitalOcean SSL certificate described once in a Planton manifest: either a free Let's Encrypt certificate that DigitalOcean issues and auto-renews (supply the domains), or a custom certificate you upload as PEM material (supply the key and certificate). The choice of branch in the manifest fully determines the certificate type — there is no separate type field to keep consistent.

## What this kind models

The spec maps onto DigitalOcean's `digitalocean_certificate` in full:

| Spec field | What it controls |
|---|---|
| `certificateName` | The certificate's stable identity — the resource ID, and what load balancers reference |
| `letsEncrypt.domains` | The FQDNs (or wildcards) a Let's Encrypt certificate covers; every domain must be DigitalOcean-DNS-managed in the same account |
| `custom.leafCertificate` | PEM-encoded server certificate |
| `custom.privateKey` | PEM-encoded private key (a secret — DigitalOcean never returns it; state stores only a hash) |
| `custom.certificateChain` | Optional PEM-encoded intermediate chain |

Exactly one of `letsEncrypt` / `custom` is set; validation enforces the choice.

## Quick start

A Let's Encrypt certificate:

```yaml
apiVersion: digital-ocean.planton.dev/v1alpha1
kind: DigitalOceanCertificate
metadata:
  name: web-cert
spec:
  certificateName: web-cert
  letsEncrypt:
    domains:
      - example.com
      - "*.example.com"
```

A custom certificate:

```yaml
spec:
  certificateName: uploaded-cert
  custom:
    leafCertificate: |
      -----BEGIN CERTIFICATE-----
      ...
      -----END CERTIFICATE-----
    privateKey: |
      -----BEGIN PRIVATE KEY-----
      ...
      -----END PRIVATE KEY-----
```

Consume it from a load balancer's forwarding rule by NAME:

```yaml
# on the DigitalOceanLoadBalancer spec
forwardingRules:
  - entryProtocol: https
    entryPort: 443
    targetProtocol: http
    targetPort: 80
    certificateName:
      valueFrom:
        kind: DigitalOceanCertificate
        name: web-cert
        fieldPath: status.outputs.certificate_id
```

## Behavior worth knowing

- **The name IS the identity** — a Let's Encrypt certificate's UUID rotates on every auto-renewal, so DigitalOcean addresses certificates by their stable name; the `certificate_id` output carries the name.
- **Let's Encrypt needs DigitalOcean DNS** — issuance validates ownership against DigitalOcean-managed zones in the same account; domains hosted elsewhere fail at create.
- **Everything is create-only** — any change replaces the certificate, created-before-destroyed so consumers referencing the name never observe a gap.
- **PEM material is write-only** — the API never returns it; an imported custom certificate carries empty PEM fields (expected, not drift).
- **Certificates import by name** — the catalog's first name-addressed import.

## Outputs

| Output | Meaning |
|---|---|
| `certificate_id` | The certificate's resource identifier — the NAME, not a UUID (`status.outputs.certificate_id`) |
| `expiry_rfc3339` | Expiration timestamp; moves forward on every Let's Encrypt auto-renewal |

## See also

- `GUIDE.md` — operational judgment calls (Let's Encrypt vs custom, rotation, the UUID trap)
- `presets/` — Let's Encrypt and custom starting points
- `v1alpha1/reference.md` — the generated field-by-field contract

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
