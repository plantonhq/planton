# Cloudflare DNS Zone

Deploys a Cloudflare DNS zone — the root of every zone-scoped Cloudflare resource — with optional inline DNS records at the full record surface, zone-wide DNS settings, DNSSEC signing, a zone hold, and a plan subscription. Its `zone_id` output anchors ValueFromRef wiring for records, rulesets, load balancers, custom hostnames, and every other zone-scoped Infra Component in InfraPipelines.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Cloudflare Zone** -- a DNS zone for the specified domain under the target Cloudflare account, as a full, partial (CNAME setup), secondary, or internal zone
- **DNS Records** -- created only when `records` entries are provided; all 21 Cloudflare record types are supported, with simple records set through `content` and structured records (SRV, CAA, CERT, DNSKEY, DS, HTTPS, LOC, NAPTR, SMIMEA, SSHFP, SVCB, TLSA, URI) through typed data blocks, plus per-record TTL, proxy status, priority, comment, tags, serving settings, and private routing
- **Zone DNS Settings** -- created only when `dnsSettings` is set; CNAME flattening, zone mode, SOA tuning, nameserver-set selection, and internal-DNS fallback
- **Zone DNSSEC** -- created only when `dnssec.enabled` is true; Cloudflare signs the zone and the DS material surfaces as outputs for your registrar
- **Zone Hold** -- created only when `hold.enabled` is true; blocks this hostname (and optionally all subdomains) from being added as a zone in any other Cloudflare account
- **Zone Subscription** -- created only when `subscription` is set; subscribes the zone to a Cloudflare rate plan

## Before You Deploy

### Planton Setup

- **Cloudflare Provider Connection** -- an active connection in the Connect module with a Cloudflare API token that has **Zone -> Zone -> Edit** and **Zone -> DNS -> Edit** (add **Billing -> Edit** when setting a subscription). Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline API token authentication.

### Cloudflare Account

- **A Cloudflare account** able to create zones; the `accountId` field identifies which account owns the zone.
- **Domain registration** -- the domain in `zoneName` must be registered with a registrar. Cloudflare assigns nameservers after zone creation; the zone stays `pending` until the registrar delegates to them.

## Deploy

### Console

Open the deployment store, find **Cloudflare DNS Zone**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Basic Zone** preset in the [Presets](#presets) tab to pre-populate a working configuration.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: cloudflare.planton.dev/v1alpha1
kind: CloudflareDnsZone
metadata:
  name: example-zone
  org: acme-corp
  env: prod
spec:
  zoneName: example.com
  accountId: "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4"
```

```shell
planton apply -f cloudflare-dns-zone.yaml
```

This creates a full Cloudflare-hosted DNS zone for example.com. Cloudflare assigns nameservers that must be configured at your domain registrar. An Infra Job tracks the provisioning in real time.

## Key Configuration

These are the most important decisions when configuring a DNS zone. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Inline records vs. separate resources** -- The `records` field manages DNS records as part of the zone, at the same depth as the standalone CloudflareDnsRecord kind (all 21 types, typed structured data, tags, settings). Use inline records when their lifecycle tracks the zone (bootstrap records, mail exchangers, verification TXTs); use separate CloudflareDnsRecord resources for records that change independently.

**Zone type** -- Defaults to `full` (Cloudflare hosts the zone's DNS entirely). Choose `partial` for a partner-hosted CNAME setup, `secondary` to mirror an external primary, or `internal` for internal-only resolution (pair with `dnsSettings.internalDns`).

**Paused mode** -- Set `paused: true` to create the zone in DNS-only mode with no Cloudflare proxy, CDN, or security features active. Useful during initial migration when you want to verify DNS resolution before enabling Cloudflare's proxy layer.

**DNSSEC** -- Set `dnssec.enabled: true` to have Cloudflare sign the zone. The DS record material (digest, key tag, algorithm) surfaces as outputs to enter at your registrar; the chain of trust completes once the registrar accepts them.

**Zone hold** -- Set `hold.enabled: true` to block this hostname from being created as a zone in any other Cloudflare account -- the standard takeover guard during account migrations. `holdAfter` schedules a temporary release window.

**Plan subscription** -- Set `subscription.ratePlan` to subscribe the zone to a Cloudflare plan (`free`, `pro`, `business`, `enterprise`, or partner variants). Paid plans bill immediately and the deploying token needs Billing Write scope.

## Outputs and Dependencies

### What This Kind Consumes

This kind has no required foreign keys. An internal zone's `dnsSettings.internalDns.referenceZoneId` may reference another CloudflareDnsZone for fallback resolution.

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `zone_id` | The Cloudflare Zone ID of the created DNS zone | CloudflareDnsRecord, CloudflareRuleset, CloudflareLoadBalancer, CloudflareCustomHostname, CloudflareCertificatePack, CloudflareEmailRouting*, CloudflareR2Bucket custom domains, CloudflareWorker routes/domains |
| `nameservers` | The nameserver addresses assigned to this zone | Domain registrar NS configuration |
| `status` | The zone's activation status (`pending` until delegated) | Deployment gating |
| `dnssec_ds`, `dnssec_digest`, `dnssec_digest_type`, `dnssec_algorithm`, `dnssec_key_tag`, `dnssec_public_key` | DS/DNSKEY material, populated only when DNSSEC is enabled | Registrar DS record entry to complete the chain of trust |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Basic zone** -- A DNS zone with no inline records; add DNS records separately as CloudflareDnsRecord resources for independent lifecycle management. Start from the **Basic Zone** preset.

**DNSSEC-signed zone** -- A zone with Cloudflare signing enabled and the DS material exported for the registrar. Start from the **DNSSEC-Signed Zone** preset.

**Zone with a record portfolio** -- A zone bootstrapped with its full record set inline: web A/AAAA records, mail MX and SPF/DKIM TXTs, service SRVs, and issuance-controlling CAAs. Start from the **Typed Records** preset.

## Works With

The zone is the anchor of the Cloudflare resource graph — every zone-scoped kind references its `zone_id` output via ValueFromRef:

- [**Cloudflare DNS Record**](/infra-catalog/cloudflare-dns-record) -- records with lifecycles independent of the zone
- [**Cloudflare Ruleset**](/infra-catalog/cloudflare-ruleset) -- zone-phase rules (WAF, redirects, transforms)
- [**Cloudflare Load Balancer**](/infra-catalog/cloudflare-load-balancer) -- traffic steering on a hostname in the zone
- [**Cloudflare Certificate Pack**](/infra-catalog/cloudflare-certificate-pack) -- advanced edge certificates ordered for the zone
- [**Cloudflare Custom Hostname**](/infra-catalog/cloudflare-custom-hostname) and [**Cloudflare Custom Hostname Fallback Origin**](/infra-catalog/cloudflare-custom-hostname-fallback-origin) -- the Cloudflare-for-SaaS surface on the zone
- [**Cloudflare Email Routing Zone**](/infra-catalog/cloudflare-email-routing-zone) and [**Cloudflare Email Routing Rule**](/infra-catalog/cloudflare-email-routing-rule) -- email routing enabled on the zone
- [**Cloudflare Worker**](/infra-catalog/cloudflare-worker) -- Worker routes and custom domains bound to the zone
