# Cloudflare Certificate Pack

Orders an advanced edge certificate for a zone: a publicly-trusted (browser-trusted) TLS certificate, provisioned and auto-renewed by Cloudflare, that covers the hostnames you list beyond the free Universal SSL certificate. Use it when you need a specific certificate authority, multiple or longer-lived certificates, or coverage for hostnames Universal SSL does not include. A pack is an order, not a document you edit: changing hosts, CA, validation method, or validity replaces the pack and starts a new validation cycle.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Certificate Pack** -- an advanced, auto-renewed edge certificate ordered against the zone. The pack sits in `pending_validation` until domain control validation completes, then serves as `active` at Cloudflare's edge for every host it covers.

## Before You Deploy

### Planton Setup

- **Cloudflare Provider Connection** -- an active connection in the Connect module with a Cloudflare API token that has SSL and Certificates edit access. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline API token authentication.

### Cloudflare Account

- **A zone** -- the certificate pack is ordered against an existing zone (`zoneId`). For zones on Cloudflare's nameservers, TXT validation completes automatically; zones elsewhere need a manual DCV step for each covered host.

## Deploy

### Console

Open the deployment store, find **Cloudflare Certificate Pack**, and click **Deploy**. The creation wizard captures the zone, the certificate authority, validation method, validity, and the covered hosts. Start from the **Advanced Certificate (TXT validation)** or **Let's Encrypt, Apex-Only, Annual** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: cloudflare.planton.dev/v1alpha1
kind: CloudflareCertificatePack
metadata:
  name: edge-cert-pack
  org: acme-corp
  env: prod
spec:
  zoneId:
    valueFrom:
      kind: CloudflareDnsZone
      name: acme-zone
      fieldPath: status.outputs.zone_id
  certificateAuthority: google
  validationMethod: txt
  validityDays: 90
  hosts:
    - example.com
    - "*.example.com"
```

```shell
planton apply -f cloudflare-certificate-pack.yaml
```

This orders a 90-day advanced certificate from Google Trust Services covering the apex and its subdomains, validated automatically over TXT. An Infra Job tracks the provisioning in real time.

### InfraChart

When the zone is deployed in the same InfraPipeline, wire the pack to it with ValueFromRef:

```yaml
spec:
  zoneId:
    valueFrom:
      kind: CloudflareDnsZone
      name: acme-zone
      fieldPath: status.outputs.zone_id
  certificateAuthority: lets_encrypt
  validationMethod: txt
  validityDays: 90
  hosts:
    - acme.com
    - "*.acme.com"
```

The InfraPipeline resolves the dependency graph, creates the zone first, then orders the pack with the resolved zone ID.

## Key Configuration

These are the most important decisions when configuring a certificate pack. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Almost everything is immutable** -- changing `hosts`, `certificateAuthority`, `validationMethod`, or `validityDays` re-orders (replaces) the pack. There is no in-place renew-with-new-hosts path: a later edit is a delete-and-recreate with a fresh validation cycle. Plan the hostname set before the first apply.

**Hosts (`hosts`)** -- must include the zone apex and may not exceed 50 entries. Apex plus wildcard (`example.com`, `*.example.com`) covers most deployments; a wildcard only covers one subdomain level, so `a.b.example.com` needs its own entry.

**Validation Method (`validationMethod`)** -- `txt` completes automatically for zones on Cloudflare's nameservers and is the right default. `http` requires the origin to serve a well-known URL; `email` sends mail to the CA's registered contacts. Either of those turns issuance into a human-visible wait in `pending_validation`.

**Certificate Authority (`certificateAuthority`)** -- `lets_encrypt` is the sensible default for most zones. `google` and `ssl_com` change the CA's validation behavior and branding rules, not the manifest shape. CA-specific restrictions apply -- Let's Encrypt caps validity at 90 days.

**Validity (`validityDays`)** -- 14, 30, 90, or 365. Cloudflare auto-renews the pack before expiry regardless of length, so shorter validity costs nothing operationally; 365 mainly suits teams that audit certificate rotation annually.

**Cloudflare Branding (`cloudflareBranding`)** -- adds a `sni.cloudflaressl.com` subdomain as the Common Name. Leave it unset unless you specifically need the branded SAN; it only applies where the CA supports it.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **CloudflareDnsZone** | `zoneId` | `status.outputs.zone_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `certificate_pack_id` | The Cloudflare-assigned pack identifier | Verification tooling and imports -- always paired with `zone_id`, since a pack's API identity is the (zone, pack) tuple |
| `zone_id` | The zone the pack was ordered in | Completes the pack's API identity for tooling that composes on it |

Issuance status and the primary-certificate id are deliberately not outputs — both transition asynchronously server-side as issuance progresses; read them from the Cloudflare API or dashboard.

## Common Patterns

**Apex + wildcard, automatic validation** -- one pack covering `example.com` and `*.example.com`, validated over TXT with no manual step for zones on Cloudflare's nameservers. Start from the **Advanced Certificate (TXT validation)** preset.

**Apex-only annual rotation** -- a single-hostname certificate from Let's Encrypt at 365-day validity, for teams that only need the bare domain and prefer a yearly rotation cadence. Start from the **Let's Encrypt, Apex-Only, Annual** preset.

**Fresh order over import** -- when a pack already exists in the zone, prefer ordering a replacement through Planton rather than adopting the existing one. Most order fields do not round-trip through provider import, so a post-import plan stays noisy even when nothing operationally changed.

## Works With

- [**Cloudflare DNS Zone**](/infra-catalog/cloudflare-dns-zone) -- the zone the certificate is ordered for; `zoneId` references its output
- [**Cloudflare Zone TLS Settings**](/infra-catalog/cloudflare-zone-tls-settings) -- controls the TLS posture (minimum version, ciphers) the certificate is served with
