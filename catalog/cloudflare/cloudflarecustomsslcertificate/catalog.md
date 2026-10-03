# Cloudflare Custom SSL Certificate

Uploads a bring-your-own TLS certificate to a zone — Cloudflare presents it to visitors instead of a Universal SSL or Advanced certificate. The certificate must be issued by a CA on Cloudflare's trust list (self-signed is rejected), and custom certificates are a Business/Enterprise zone feature enforced at create. This is the manual-renewal path: unlike Universal SSL and Certificate Packs, a custom certificate expires on your calendar, and rotation is replacement — the certificate ID changes, with Cloudflare serving the old certificate until the new one deploys.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Custom Certificate Upload** — one `cloudflare_custom_ssl` on the zone, carrying the chosen SNI class (`type`), bundle method, optional private-key geo restrictions, and optional staging-network deploy. Certificate `priority` is deliberately not managed: at provider v5.23.0 it is read-only.

Destroy is a real delete: the zone falls back to Universal SSL / Advanced certificates — verify one is active first, or visitors get TLS handshake failures.

## Before You Deploy

### Planton Setup

- **Cloudflare Provider Connection** — an active connection in the Connect module whose API token carries **Zone → SSL and Certificates → Edit**. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** — required when using Runner-based credential delivery. Not needed for inline credential authentication.

### Cloudflare Account

- **A zone on a Business plan or above** — lower plans are rejected at the API (402/403).
- **A certificate issued by a publicly trusted CA** covering the zone's hostnames, plus its private key. Store the private key as a managed secret and reference it from `privateKey` — the API never returns the key. If you only need Cloudflare-to-origin authentication, use the Cloudflare Authenticated Origin Pulls Certificate component instead — self-signed is normal there, rejected here.

## Deploy

### Console

Open the deployment store, find **Cloudflare Custom SSL Certificate**, and click **Deploy**. The creation wizard walks you through environment and connection configuration, the target zone, the certificate and private-key material, and the SNI class, bundle method, and geo-restriction choices. Start from the **SNI custom certificate** preset in the [Presets](#presets) tab to pre-populate a working configuration.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: cloudflare.planton.dev/v1alpha1
kind: CloudflareCustomSslCertificate
metadata:
  name: www-custom-certificate
  org: acme-corp
  env: prod
spec:
  zoneId:
    value: "0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d"
  certificate: |
    -----BEGIN CERTIFICATE-----
    ...
    -----END CERTIFICATE-----
  privateKey:
    value: ${secrets-group/prod-tls/www-key}
  type: sni_custom
```

```shell
planton apply -f custom-certificate.yaml
```

This uploads an SNI-class custom certificate to the zone with Cloudflare's default ubiquitous bundle method — Cloudflare starts presenting it to SNI-capable visitors once the asynchronous deployment settles. An Infra Job tracks the provisioning in real time.

### InfraChart

When the zone is deployed in the same InfraPipeline, wire the reference with ValueFromRef:

```yaml
spec:
  zoneId:
    valueFrom:
      kind: CloudflareDnsZone
      name: acme-zone
      fieldPath: status.outputs.zone_id
  certificate: |
    -----BEGIN CERTIFICATE-----
    ...
    -----END CERTIFICATE-----
  privateKey:
    value: ${secrets-group/prod-tls/compliance-key}
  type: sni_custom
  bundleMethod: optimal
  geoRestrictions:
    label: us
```

The InfraPipeline resolves the dependency graph, deploys the zone first, then uploads the certificate to the resolved zone ID.

## Key Configuration

These are the most important decisions when configuring a custom SSL certificate. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**You own the renewal now** — Universal SSL and Certificate Packs renew themselves; a custom certificate expires on your calendar. Watch the `expires_on` output and rotate before it. Rotation is replacement: the upload is destroyed and re-created, the `certificate_id` changes, and anything referencing the old ID must follow. Cloudflare keeps serving the previous certificate until the replacement deploys, so a timely rotation is not an outage.

**legacy_custom or sni_custom** — `legacy_custom` (the default) works on every TLS client but occupies the zone's single legacy slot; a second legacy upload is rejected. `sni_custom` requires SNI-capable clients — every modern browser — and allows multiple uploads. Changing `type` on an existing upload replaces it. Prefer `sni_custom` unless you have measured legacy-client traffic.

**Bundle method** — `ubiquitous` (the default) maximizes older-client compatibility, `optimal` prefers the shortest modern chain, and `force` uses the chain exactly as uploaded — pick `force` only when your PKI team ships a deliberate chain.

**Private-key geography** — `geoRestrictions.label` (`us`, `eu`, `highest_security`) is the coarse control over which data centers hold the private key; `policy` is the fine-grained expression form (e.g. `(country: US) or (region: EU)`). The API returns the parsed policy in a separate read-only field, so a non-empty re-plan on `policy` after apply is provider normalization, not manifest drift.

**Keep the PEM byte-stable** — the provider replaces the upload on any change to the `certificate` value, including formatting-only churn. Store the PEM exactly as issued, trailing newline included.

**Priority is not manageable** — the v4 reprioritization surface was dropped; at v5.23.0 `priority` is read-only. If a specific certificate must be served preferentially, control it by what you upload, not by an ordering knob.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **CloudflareDnsZone** | `zoneId` | `status.outputs.zone_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `certificate_id` | The uploaded certificate's ID (changes on every rotation) | Referencing the upload in TLS tooling and audits |
| `zone_id` | The zone the certificate belongs to | Confirming the upload landed in the intended zone |
| `expires_on` | Expiry timestamp (RFC3339) | Rotation scheduling and expiry alerting — the renewal is yours |

Deployment status is deliberately not an output: deployment is asynchronous (pending before active), so a point-in-time phase would flip on the first refresh and re-plan forever. Read deployment status from the Cloudflare API or dashboard.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Compliance-issued certificate** — an EV/OV certificate from a mandated issuer, uploaded as `sni_custom` with the compatibility-maximizing bundle. The shape for zones where an external PKI team owns the certificate lifecycle. Start from the **SNI custom certificate** preset.

**Data-residency key handling** — `geoRestrictions.label: us` or `eu` (or a fine-grained `policy` expression) keeps the private key inside the required region while Cloudflare serves globally.

**Staged validation** — `deploy: staging` places the certificate on Cloudflare's staging network first (Business/Enterprise), so the chain and hostname coverage are validated before production traffic sees it.

## Works With

- [**Cloudflare DNS Zone**](/infra-catalog/cloudflare-dns-zone) — the zone the certificate is uploaded to; wire `zoneId` via ValueFromRef
- [**Cloudflare Certificate Pack**](/infra-catalog/cloudflare-certificate-pack) — the managed-renewal alternative; pick it unless an external issuer is mandated
- [**Cloudflare Zone TLS Settings**](/infra-catalog/cloudflare-zone-tls-settings) — minimum TLS version and Universal SSL posture on the same zone
- [**Cloudflare Authenticated Origin Pulls Certificate**](/infra-catalog/cloudflare-authenticated-origin-pulls-certificate) — the right kind when the goal is Cloudflare-to-origin authentication rather than visitor-facing TLS
