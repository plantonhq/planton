# Cloudflare Origin CA Certificate

Provisions a Cloudflare Origin CA certificate: a free TLS certificate that Cloudflare's edge trusts, installed on your origin server so the Cloudflare-to-origin hop runs encrypted end-to-end (the "Full (Strict)" SSL mode). It is not browser-trusted -- it is valid only between Cloudflare and your origin. By default the module generates the private key and CSR for you and returns both, so a downstream origin can mount the certificate and key without any out-of-band key handling.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Origin CA Certificate** -- an edge-trusted certificate valid for the hostnames you list
- **Private Key (optional)** -- generated and returned as a sensitive output when no CSR is supplied

## Before You Deploy

### Planton Setup

- **Cloudflare Provider Connection** -- an active connection in the Connect module with a Cloudflare API token that has Origin CA edit access. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline API token authentication.

## Deploy

### Console

Open the deployment store, find **Cloudflare Origin CA Certificate**, and click **Deploy**. The creation wizard captures the hostnames, how the key is obtained (Cloudflare generates it, or bring your own CSR), and the validity. Start from the **Generated Key (recommended)** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: cloudflare.planton.dev/v1alpha1
kind: CloudflareOriginCaCertificate
metadata:
  name: edge-origin-cert
  org: acme-corp
  env: prod
spec:
  hostnames:
    - example.com
    - "*.example.com"
  requestType: origin-rsa
  requestedValidity: 5475
```

```shell
planton apply -f cloudflare-origin-ca-certificate.yaml
```

This issues a 15-year RSA Origin CA certificate covering the apex and its subdomains. An Infra Job tracks the provisioning in real time.

## Key Configuration

These are the most important decisions when configuring an Origin CA certificate. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Hostnames (`hostnames`)** -- The Subject Alternative Names the certificate covers. Wildcards like `*.example.com` are allowed, and hostnames may belong to any zone your account controls. Immutable -- changing them re-issues the certificate.

**Key Algorithm (`requestType`)** -- `origin-rsa` (default, broadest compatibility), `origin-ecc` (smaller/faster), or `keyless-certificate`. Applies when Cloudflare generates the key. Immutable.

**Validity (`requestedValidity`)** -- 7, 30, 90, 365, 730, 1095, or 5475 days. Because these certificates are not browser-trusted, the 15-year default is safe and avoids renewal churn. Immutable.

**CSR (`csr`)** -- Supply a PEM CSR to use your own key material; no private key is generated and your key never leaves your control. Omit it (recommended) to have Cloudflare generate the key + CSR.

## Outputs and Dependencies

### What This Kind Consumes

This kind has no foreign key dependencies -- it is account-scoped (there is no `zoneId`), and the covered hostnames travel as plain strings.

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `certificate` | The issued certificate (PEM) | Install on the origin, e.g. as the `tls.crt` of a Kubernetes TLS secret |
| `private_key` | The generated private key (PEM), when no CSR was supplied. Sensitive. | Mount on the origin alongside the certificate |
| `expires_on` | RFC3339 expiry timestamp | Rotation tracking before the certificate lapses |

`status.outputs` also carries `certificate_id`, the identifier you would use to revoke the certificate.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**One-click cert + key** -- let the module generate everything and mount both outputs into a Kubernetes TLS secret on the origin. This is the right default when your key-handling story is "keep it inside Planton outputs". Start from the **Generated Key (recommended)** preset.

**Bring your own key** -- supply a CSR so the private key never leaves your infrastructure; `status.outputs.private_key` stays empty. Choose this when key custody policy forbids generated keys. Start from the **Bring Your Own CSR** preset.

## Works With

- [**Cloudflare DNS Zone**](/infra-catalog/cloudflare-dns-zone) -- the zone whose origin serves traffic behind this certificate
