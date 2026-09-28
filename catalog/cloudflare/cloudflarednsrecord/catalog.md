# Cloudflare DNS Record

Deploys a single DNS record within a Cloudflare zone, covering both simple types (A, AAAA, CNAME, MX, TXT, NS, PTR, OPENPGPKEY) whose value is one `content` string, written as a literal or read from another resource's output and structured types (SRV, CAA, HTTPS, SVCB, TLSA, DS, DNSKEY, CERT, LOC, NAPTR, SMIMEA, SSHFP, URI) whose fields travel in a typed `data` block. Proxy status, TTL, MX priority, comments, tags, and per-record serving settings are all configurable, and the manifest is validated so the value representation always matches the record type.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **Cloudflare DNS Record** -- one record in the specified zone with the configured type, value, TTL, and proxy setting. When `proxied` is `true` (A, AAAA, or CNAME only), traffic routes through Cloudflare's CDN and WAF and the origin IP is hidden; otherwise the record is DNS-only.

## Before You Deploy

### Planton Setup

- **Cloudflare Provider Connection** -- an active connection in the Connect module with a Cloudflare API token that has DNS edit access. Map it as the default for your environment, or specify it explicitly when creating the Cloud Resource.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline API token authentication.

### Cloudflare Account

- **A Cloudflare DNS zone** -- the target zone must exist and be active. Provide the zone ID directly or reference a CloudflareDnsZone Cloud Resource via ValueFromRef.
- **The record's target** -- for A records, the origin's IPv4 address; for AAAA, the IPv6 address; for CNAME/MX/NS, the target hostname; for structured types, the fields of the matching `data` block.

## Deploy

### Console

Open the deployment store, find **Cloudflare DNS Record**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Proxied A Record** preset in the [Presets](#presets) tab to pre-populate a proxied web-facing record.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: cloudflare.planton.dev/v1alpha1
kind: CloudflareDnsRecord
metadata:
  name: www-record
  org: acme-corp
  env: prod
spec:
  zoneId:
    value: "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4"
  name: www
  type: A
  content:
    value: "203.0.113.50"
  proxied: true
  ttl: 1
```

```shell
planton apply -f cloudflare-dns-record.yaml
```

This creates a proxied A record pointing `www` to the specified IP address with automatic TTL. Traffic flows through Cloudflare's CDN and WAF. A Stack Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the DNS record to a zone deployed in the same InfraPipeline:

```yaml
spec:
  zoneId:
    valueFrom:
      kind: CloudflareDnsZone
      name: example-zone
      fieldPath: status.outputs.zone_id
  name: www
  type: A
  content:
    value: "203.0.113.50"
  proxied: true
```

The InfraPipeline resolves the dependency graph, deploys the DNS zone first, then provisions the DNS record with the resolved zone ID.

## Key Configuration

These are the most important decisions when configuring a DNS record. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**`content` vs. `data` -- exactly one, matching the type** -- simple types (A, AAAA, CNAME, MX, TXT, NS, PTR, OPENPGPKEY) take a single presentation-format string in `content`, either `value:` or a `valueFrom:` reference; structured types (SRV, CAA, HTTPS, and the rest) take their matching `data` block (`data.srv`, `data.caa`, ...). Setting both, neither, or a mismatched block fails validation before anything reaches Cloudflare. An SRV record's priority, weight, port, and target all live inside `data.srv` -- the top-level `priority` field is for MX only.

**Proxy vs. DNS-only (`proxied`)** -- `true` (orange cloud) routes traffic through Cloudflare's CDN, WAF, and DDoS protection while hiding the origin IP; `false` (grey cloud) resolves directly and exposes the origin. Only A, AAAA, and CNAME records can be proxied -- the spec rejects `proxied: true` on anything else.

**TTL (`ttl`)** -- `1` (or 0) means automatic, which is what proxied records should use. For DNS-only records, 30-86400 seconds -- the 30s floor applies to Enterprise zones, most zones start at 60. Lower TTL means faster failover at the cost of more DNS query volume.

**MX priority (`priority`)** -- required for MX records, lower is preferred: `10` for the primary mail server, `20`/`30` for backups. Ignored for every other type.

**Record settings (`settings`)** -- proxied-only toggles: `ipv4Only`/`ipv6Only` suppress the other address family, and `flattenCname` resolves a CNAME externally and returns address records (only meaningful DNS-only, since proxied CNAMEs are always flattened). Leave the block out unless you know you need one of these.

**Private routing (`privateRouting`)** -- restricts the record to Cloudflare's internal routing (internal DNS / Magic WAN) instead of the public internet. Defaults to public; turning it on is a topology decision, not a tuning knob.

## Outputs and Dependencies

### What This Component Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **CloudflareDnsZone** | `zoneId` | `status.outputs.zone_id` |

### What This Component Provides

After provisioning, `status.outputs` contains values that downstream Cloud Resources can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `record_id` | The unique identifier of the created DNS record | Verification tooling and imports -- paired with `zone_id`, since a record's API identity is the (zone, record) tuple |
| `record_name` | The record's name within the zone (`www`, or `@` for the apex) | Application configuration and health-check endpoints built on the resolved hostname |
| `record_type` | The DNS record type that was created | Distinguishing records when a chart manages several against one zone |
| `proxied` | Whether the record is proxied through Cloudflare | Verifying the orange-cloud posture after apply |
| `zone_id` | The zone the record lives in | Completes the record's API identity for tooling that composes on it |

## Common Patterns

**Proxied A record** -- an A record with the Cloudflare proxy enabled for web servers and APIs: CDN, WAF, and DDoS protection in front, origin IP hidden behind. Start from the **Proxied A Record** preset.

**MX record for email** -- an MX record with priority routing for Google Workspace, Microsoft 365, or custom mail servers. MX records cannot be proxied and are always DNS-only. Start from the **MX Record for Email** preset.

**SRV record for a service** -- service discovery via `data.srv` (priority, weight, port, target) under a `_service._proto` name like `_sip._tcp`. Start from the **SRV Record for a Service** preset.

**CAA record to restrict issuance** -- pin which certificate authorities may issue for the domain via `data.caa` (`issue`, `issuewild`, or `iodef`). Start from the **CAA Record to Restrict Certificate Issuance** preset.

## Works With

- [**Cloudflare DNS Zone**](/cloud-catalog/cloudflare-dns-zone) -- provides the zone ID where this DNS record is created
- [**Cloudflare Custom Hostname Fallback Origin**](/cloud-catalog/cloudflare-custom-hostname-fallback-origin) -- SaaS zones need an in-zone record backing the fallback origin hostname; this component creates it
