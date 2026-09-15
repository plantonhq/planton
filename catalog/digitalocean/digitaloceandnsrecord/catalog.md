# DigitalOcean DNS Record

Creates a single DNS record within an existing DigitalOcean DNS zone. Supports every record type the DigitalOcean API accepts -- A, AAAA, CNAME, MX, TXT, SRV, NS, CAA, and SOA -- with type-specific fields for priority, weight, port, flags, and tag applied conditionally and enforced at validation time. This is the kind for records whose owner or lifecycle differs from their zone's -- an application adding its own hostname to a shared company zone -- while records that ship with the zone belong in the zone's inline list; pick one home per record so ownership is never split.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **DNS Record** -- a `digitalocean_record` resource in the specified domain with the configured type, name, value, and TTL
- **Type-Specific Attributes** -- `priority` is set for MX and SRV records; `weight` and `port` are set for SRV records; `flags` and `tag` are set for CAA records; all are omitted for inapplicable record types

## Before You Deploy

### Planton Setup

- **DigitalOcean Provider Connection** -- an active connection in the Connect module with a DigitalOcean API token. Map it as the default for your environment, or specify it explicitly when creating the Cloud Resource.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline API token authentication.

### DigitalOcean Account

- **An existing DigitalOcean DNS zone (domain)** managed by DigitalOcean's DNS service. Provide the domain name directly or reference a DigitalOceanDnsZone Cloud Resource via ValueFromRef.
- **A valid record value** matching the record type: an IPv4 address for A records, a hostname for CNAME records, a mail server for MX records, etc.

## Deploy

### Console

Open the deployment store, find **DigitalOcean DNS Record**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Apex A Record** preset in the [Presets](#presets) tab to point a zone's apex at an IP address.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: digital-ocean.planton.dev/v1alpha1
kind: DigitalOceanDnsRecord
metadata:
  name: www-a-record
  org: acme-corp
  env: prod
spec:
  domain:
    value: "example.com"
  name: www
  type: A
  value:
    value: "192.0.2.1"
  ttlSeconds: 3600
```

```shell
planton apply -f dns-record.yaml
```

This creates an A record pointing `www.example.com` to `192.0.2.1` with a one-hour TTL; no MX, SRV, or CAA-specific fields are configured. A Stack Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the record to a DNS zone deployed in the same InfraPipeline:

```yaml
spec:
  domain:
    valueFrom:
      kind: DigitalOceanDnsZone
      name: example-zone
      fieldPath: status.outputs.zone_name
```

The InfraPipeline resolves the dependency graph, deploys the DNS zone first, then provisions the DNS record within it.

## Key Configuration

These are the most important decisions when configuring a DNS record. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Record type** -- Changing `type` recreates the record, so a wrong pick costs a delete-and-create, not an edit. The apex constraint matters most: CNAME cannot exist at `@` (DNS forbids it) -- use A/AAAA there and CNAME only on subdomains. NS and SOA are writable but almost never yours to write: apex NS records belong to the zone's delegation and the SOA is DigitalOcean's operational record; the one legitimate NS use is delegating a subdomain to other nameservers.

**TTL** -- The `ttlSeconds` field controls how long DNS resolvers cache this record, defaulting to 1800 seconds (30 minutes). Use lower values (60-300) during migrations or when records change frequently, and higher values (3600-86400) for stable production records.

**Type-specific fields** -- MX records require `priority` (lower values = higher priority). SRV records require `priority`, `weight`, and `port`. CAA records require `flags` and `tag` (`issue`, `issuewild`, or `iodef`). The protobuf schema enforces these cross-field constraints at validation time. One provider quirk: an explicit 0 in `priority`, `weight`, or `port` is dropped from the create request and the API default applies -- use positive values when exactness matters (CAA `flags: 0` is safe; the API default is 0).

**Hostname values carry a trailing dot on read-back** -- CNAME, MX, NS, SRV, and CAA targets are stored fully qualified (`mail.example.com.`, `letsencrypt.org.`); author the trailing dot, or a zone-relative name (`mail`). A bare `letsencrypt.org` is re-applied on every run, forever.

**Value references** -- The `value` field supports ValueFromRef, allowing you to reference outputs from other Cloud Resources (e.g., a Droplet's IP address or a Load Balancer's hostname) instead of hardcoding values.

## Outputs and Dependencies

### What This Component Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **DigitalOceanDnsZone** | `domain` | `status.outputs.zone_name` |

### What This Component Provides

After provisioning, `status.outputs` carries `record_id`, `hostname`, `record_type`, `domain`, and `ttl_seconds` -- but no other catalog component consumes them via ValueFromRef: a DNS record is a leaf of the dependency graph. `record_type`, `domain`, and `ttl_seconds` echo the manifest back for audit (`ttl_seconds` carries the API's applied default when the spec left it unset). The genuinely new values are `record_id` -- the numeric id that, together with the domain, addresses the record in the DigitalOcean API and in imports (`{domain},{record_id}`) -- and `hostname`, the provider-computed fully qualified name to verify resolution against.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Apex A record** -- Points the zone apex (`@`) at an IPv4 address with a one-hour TTL -- the record CNAME cannot provide. Targets Droplets, Load Balancers, or external IPs; pair it with ValueFromRef to track the target resource's IP instead of hardcoding it. Start from the **Apex A Record** preset.

**WWW CNAME record** -- Aliases `www` to the apex (note the fully qualified target with its trailing dot). The same shape serves subdomains pointing to CDN origins or third-party services. Start from the **WWW CNAME Record** preset.

## Works With

- [**DigitalOcean DNS Zone**](/cloud-catalog/digital-ocean-dns-zone) -- provides the domain (DNS zone) in which records are created