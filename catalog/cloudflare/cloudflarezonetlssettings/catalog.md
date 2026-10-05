# Cloudflare Zone TLS Settings

Manages a Cloudflare zone's edge TLS posture as one resource: Universal SSL issuance, Total TLS per-hostname certificates, automatic origin TLS key exchange, origin TLS compliance modes, per-hostname TLS overrides, and certificate-authority hostname associations. Any field left unset is not managed — the zone keeps its current value. The one hazard to know up front: `universalSslEnabled: false` stops certificate issuance and can make proxied hostnames unreachable over HTTPS unless other certificates cover them.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Universal SSL Setting** — created only when `universalSslEnabled` is set; controls Universal SSL certificate issuance for the zone (no delete at Cloudflare — destroy abandons the last-applied value)
- **Total TLS Configuration** — created only when `totalTls` is set; issues individual certificates for every proxied hostname, including deep subdomains Universal SSL's wildcard does not cover (no delete at Cloudflare)
- **Automatic Origin TLS Key Exchange Setting** — created only when `autoOriginTlsKex` is set; lets Cloudflare negotiate the strongest key exchange the origin supports (no delete at Cloudflare)
- **Origin TLS Compliance Modes** — created only when `originTlsComplianceModes` is non-empty; requires the listed modes on Cloudflare-to-origin connections (real delete — destroy clears the requirement)
- **Per-Hostname TLS Settings** — one API object per set attribute per hostname row: a `minTlsVersion` override, an `http2` override, and a `ciphers` override each become their own object keyed by hostname (real delete — destroy removes the overrides and hostnames fall back to zone-wide settings)
- **CA Hostname Associations** — one object per `caHostnameAssociations` row: a row without `mtlsCertificateId` manages the zone's managed-CA list, and each row with one manages that mTLS certificate's hostname list (no delete at Cloudflare)

## Before You Deploy

### Planton Setup

- **Cloudflare Provider Connection** — an active connection in the Connect module with an API token holding Zone → SSL and Certificates → Edit on the target zone. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** — required when using Runner-based credential delivery. Not needed for inline credential authentication.

### Cloudflare Account

- **A zone on the account** — `zoneId` names the zone; reference a CloudflareDnsZone Infra Component or pass the zone ID from the dashboard.
- **Advanced Certificate Manager** (required for Total TLS and per-hostname overrides) — without the zone's ACM subscription the API refuses these surfaces with 401 code 1450 (measured live).
- **An ACTIVE zone for `autoOriginTlsKex`** — the automatic key-exchange setting is refused on zones still pending activation.
- **An mTLS certificate** (only for per-certificate CA associations) — `caHostnameAssociations[].mtlsCertificateId` references a CloudflareMtlsCertificate Infra Component or a literal certificate ID.

## Deploy

### Console

Open the deployment store, find **Cloudflare Zone TLS Settings**, and click **Deploy**. The creation wizard walks you through environment and connection configuration, the zone reference, and the six TLS surfaces. Start from the **Total TLS with Google Trust Services** preset in the [Presets](#presets) tab to pre-populate a working configuration.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: cloudflare.planton.dev/v1alpha1
kind: CloudflareZoneTlsSettings
metadata:
  name: zone-tls
  org: acme-corp
  env: prod
spec:
  zoneId:
    value: "0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d"
  totalTls:
    enabled: true
```

```shell
planton apply -f zone-tls.yaml
```

This enables Total TLS — every proxied hostname in the zone gets its own certificate — while Universal SSL, origin settings, and per-hostname overrides remain untouched. At least one TLS surface must be configured; a resource that manages nothing is rejected. An Infra Job tracks the provisioning in real time.

### InfraChart

When the zone and an mTLS certificate are deployed in the same InfraPipeline, wire the references with ValueFromRef:

```yaml
spec:
  zoneId:
    valueFrom:
      kind: CloudflareDnsZone
      name: acme-com
      fieldPath: status.outputs.zone_id
  totalTls:
    enabled: true
    certificateAuthority: google
  caHostnameAssociations:
    - hostnames:
        - mtls.acme.com
      mtlsCertificateId:
        valueFrom:
          kind: CloudflareMtlsCertificate
          name: partner-mtls
          fieldPath: status.outputs.certificate_id
```

The InfraPipeline resolves the dependency graph, deploys the zone and the certificate first, then applies the TLS settings with the resolved IDs.

## Key Configuration

These are the most important decisions when configuring a zone's TLS settings. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Universal SSL is the one switch that can cause an outage** — `universalSslEnabled: false` stops issuance, and any proxied hostname not covered by another certificate (dedicated, custom, or Total TLS) becomes unreachable over HTTPS: browsers show certificate errors, with no partial degradation. Disable it only as a step in a planned migration, after confirming every proxied hostname has other coverage — typically a CloudflareCertificatePack ordered first. And because the setting has no delete, disabling it and then destroying the resource leaves the zone disabled until something re-enables it.

**Know your destroy class before you destroy** — the six surfaces split four-and-two. `universalSslEnabled`, `totalTls`, `autoOriginTlsKex`, and `caHostnameAssociations` have no delete: destroy drops them from state and the zone keeps the last-applied values, so write the value you want to leave behind before destroying. `hostnameSettings` and `originTlsComplianceModes` have real deletes: destroy is a real revert — do not destroy a resource carrying a compliance requirement you still need.

**One hostname row fans out into per-setting objects** — a `hostnameSettings` row setting `minTlsVersion` and `http2` becomes two API objects, each keyed by (setting, hostname). Editing one row never churns another row's resources, but renaming a hostname replaces its override objects rather than updating them. Each row must set at least one of the three overrides.

**Total TLS validity is not yours to set** — certificates live for 90 days, fixed by Cloudflare, and auto-renew. The only decision is `certificateAuthority` (`google`, `lets_encrypt`, or `ssl_com`) when compliance cares; unset lets Cloudflare choose.

**Compliance modes are an open vocabulary** — `originTlsComplianceModes` passes any string through to the API deliberately, so a new Cloudflare mode works the day it ships. Cloudflare documents `fips` and `pqh` (post-quantum hybrid) today. The cost: a typo also passes through, and Cloudflare's API — not manifest validation — is what rejects it. Check the apply output.

**The zone-wide layer lives elsewhere** — zone-wide TLS knobs (minimum TLS version for the whole zone, Always Use HTTPS, the zone cipher list) belong to CloudflareZoneSettings; this kind is the issuance and per-hostname layer. Per-hostname overrides here apply only to hostnames in this zone — SaaS vanity hostnames are CloudflareCustomHostname territory.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **CloudflareDnsZone** | `zoneId` | `status.outputs.zone_id` |
| **CloudflareMtlsCertificate** (optional, per row) | `caHostnameAssociations[].mtlsCertificateId` | `status.outputs.certificate_id` |

### What This Kind Provides

This kind has no consumable outputs of its own: TLS settings are a zone-scoped singleton with no resource ID, so `status.outputs` only echoes the input `zone_id` back for reference. Downstream resources that need the zone should reference the CloudflareDnsZone directly.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Total TLS everywhere** — per-hostname certificates for the whole zone, issued by Google Trust Services, with automatic origin key exchange; the shape for zones with deep subdomains beyond Universal SSL's wildcard. Start from the **Total TLS with Google Trust Services** preset.

**Surgical hostname hardening** — raise one hostname (an API endpoint) to TLS 1.3 with HTTP/2 while the rest of the zone stays on zone-wide settings; a legacy hostname can simultaneously stay on TLS 1.0 with a pinned cipher. Start from the **Per-Hostname TLS Overrides** preset.

**Compliance-bound origins** — require `fips` (or `pqh`) on every Cloudflare-to-origin connection for regulated workloads, keeping the requirement declared in the graph rather than in a dashboard.

## Works With

- [**Cloudflare DNS Zone**](/infra-catalog/cloudflare-dns-zone) — the zone whose TLS posture this resource manages; `zoneId` references its `zone_id` output
- [**Cloudflare Certificate Pack**](/infra-catalog/cloudflare-certificate-pack) — advanced edge certificates, often the coverage that makes disabling Universal SSL safe
- [**Cloudflare mTLS Certificate**](/infra-catalog/cloudflare-mtls-certificate) — the certificate whose CA hostname associations a row can manage
- [**Cloudflare Zone Settings**](/infra-catalog/cloudflare-zone-settings) — zone-wide TLS knobs (minimum TLS version, Always Use HTTPS) live there
- [**Cloudflare Custom Hostname**](/infra-catalog/cloudflare-custom-hostname) — TLS for SaaS vanity hostnames outside this zone
