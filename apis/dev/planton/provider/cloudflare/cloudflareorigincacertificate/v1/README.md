# CloudflareOriginCaCertificate

Issue a Cloudflare **Origin CA certificate** — a free TLS certificate that
Cloudflare's edge trusts, installed on your origin server so the
Cloudflare-to-origin hop runs encrypted ("Full (Strict)" SSL). It is not a
public/browser-trusted certificate; it is valid only between Cloudflare and the
origin.

This component is a one-click certificate+key node: by default it generates the
private key and CSR for you and exports the signed **certificate** plus the
(sensitive) **private key**, so a downstream origin can mount both with no
out-of-band key handling.

## When to use

- Securing the Cloudflare-to-origin connection for any proxied hostname (Full
  (Strict) mode) without buying a public cert for the origin.
- Pairing with a Kubernetes TLS secret, a load balancer listener, or a VM that
  serves HTTPS behind Cloudflare.

## Quick start

```yaml
apiVersion: cloudflare.planton.dev/v1
kind: CloudflareOriginCaCertificate
metadata:
  name: origin-cert
spec:
  hostnames:
    - example.com
    - "*.example.com"
```

## Configuration reference

| Field | Required | Description |
|---|---|---|
| `hostnames` | yes | SANs the certificate is valid for (≥1) |
| `requestType` | no | `origin-rsa` (default), `origin-ecc`, or `keyless-certificate` |
| `requestedValidity` | no | Days: 7, 30, 90, 365, 730, 1095, or 5475 (default) |
| `csr` | no | A user-supplied CSR (PEM). When set, no key is generated |

## Outputs

| Output | Description |
|---|---|
| `certificate_id` | The certificate identifier |
| `certificate` | The issued certificate (PEM); public material |
| `private_key` | The generated private key (PEM); empty if a CSR was supplied (sensitive) |
| `expires_on` | Expiry timestamp |

## A note on the private key

When you let this component generate the key (the default), the `private_key`
output is exported as a sensitive value — resolve it downstream as a managed-secret
reference rather than embedding it in plaintext. When you supply your own `csr`,
the key never leaves your control and `private_key` is empty.

## Related components

- `CloudflareDnsRecord` / `CloudflareDnsZone` — the proxied hostnames this cert
  secures on the Cloudflare-to-origin hop.
