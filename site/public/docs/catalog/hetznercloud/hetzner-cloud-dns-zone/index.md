---
title: "Hetzner Cloud DNS Zone"
description: "Hetzner Cloud DNS Zone deployment documentation"
icon: "package"
order: 100
componentName: "hetznerclouddnszone"
---

# Hetzner Cloud DNS Zone

Deploys a DNS zone on Hetzner Cloud's authoritative nameservers with declarative record set management. Supports **primary** mode (records managed via the manifest) and **secondary** mode (records synchronized from an external primary nameserver via zone transfer). Record values accept cross-component references via `valueFrom` for DNS records that automatically track infrastructure IP addresses.

## What Gets Created

- **DNS Zone** — an `hcloud_zone` resource that establishes the domain on Hetzner Cloud nameservers with the specified mode, default TTL, and delete protection settings.
- **Record Sets** — one `hcloud_zone_rrset` resource per entry in `recordSets`. Each record set groups all DNS records sharing the same (name, type) pair. Created only in primary mode; in secondary mode, records are pulled from the external primary nameserver.

## Prerequisites

- **Hetzner Cloud API token** configured via environment variable (`HCLOUD_TOKEN`) or OpenMCF provider config
- **Domain registration** — you must own the domain and have access to its registrar to configure NS delegation after zone creation

For **secondary** zones:
- An external primary nameserver accessible from Hetzner Cloud's infrastructure
- TSIG credentials if the primary requires authenticated zone transfers

## Quick Start

Create a file `dns-zone.yaml`:

```yaml
apiVersion: hetznercloud.openmcf.org/v1
kind: HetznerCloudDnsZone
metadata:
  name: my-zone
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.HetznerCloudDnsZone.my-zone
spec:
  domainName: example.com
  mode: primary
```

Deploy:

```shell
openmcf apply -f dns-zone.yaml
```

This creates an empty primary DNS zone for `example.com`. Check the `nameservers` stack output and configure these NS records at your domain registrar to activate the zone. Add record sets to the manifest to populate the zone with DNS records.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `domainName` | `string` | The DNS domain name for the zone (e.g., `example.com`). This becomes the zone's Hetzner Cloud name. Changing this value forces replacement. | min length: 1 |
| `mode` | `enum` | Zone operating mode. Valid values: `primary` (Hetzner Cloud is authoritative, records managed via `recordSets`) or `secondary` (records synchronized from external primary via zone transfer). Changing this value forces replacement. | required, defined values only |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `ttl` | `int32` | `3600` | Default TTL in seconds for records in the zone. Individual record sets can override this. |
| `deleteProtection` | `bool` | `false` | Prevents accidental deletion of the zone via the Hetzner Cloud API. |
| `primaryNameservers` | `PrimaryNameserver[]` | — | External primary nameservers for zone transfer. Required when mode is `secondary`; forbidden when mode is `primary`. |
| `primaryNameservers[].address` | `string` | — | Public IPv4 or IPv6 address of the primary nameserver. Required within each entry. |
| `primaryNameservers[].port` | `int32` | `53` | Port of the primary nameserver. |
| `primaryNameservers[].tsigAlgorithm` | `string` | — | TSIG algorithm for authenticating zone transfers (e.g., `hmac-sha256`, `hmac-sha512`). |
| `primaryNameservers[].tsigKey` | `string` | — | TSIG shared secret key for authenticating zone transfers. |
| `recordSets` | `RecordSet[]` | — | DNS record sets. Each entry manages all records for a unique (name, type) pair. Only valid when mode is `primary`; forbidden when mode is `secondary`. |
| `recordSets[].name` | `string` | — | Record name relative to the zone. Use `@` for the apex, a subdomain label (e.g., `www`), or `*` for wildcard. Required within each entry. |
| `recordSets[].type` | `string` | — | DNS record type: `A`, `AAAA`, `CNAME`, `MX`, `TXT`, `NS`, `SRV`, `CAA`, `PTR`, `TLSA`, `DS`. Required within each entry. |
| `recordSets[].ttl` | `int32` | zone TTL | Per-record-set TTL override in seconds. |
| `recordSets[].records` | `RecordValue[]` | — | Record values. At least one required per record set. |
| `recordSets[].records[].value` | `StringValueOrRef` | — | The record value. Accepts a literal string or a `valueFrom` reference to another component's output. Format depends on record type (see examples). Required within each entry. |
| `recordSets[].records[].comment` | `string` | — | Optional comment for this record. |

## Examples

### Primary Zone with Common Records

A production zone with A records, a CNAME alias, MX records for email, and TXT records for SPF and DMARC.

```yaml
apiVersion: hetznercloud.openmcf.org/v1
kind: HetznerCloudDnsZone
metadata:
  name: acme-zone
  org: acme-corp
  env: production
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: acme-corp
    pulumi.openmcf.org/project: dns
    pulumi.openmcf.org/stack.name: production.HetznerCloudDnsZone.acme-zone
spec:
  domainName: acme-corp.com
  mode: primary
  ttl: 3600
  deleteProtection: true
  recordSets:
    - name: "@"
      type: A
      ttl: 300
      records:
        - value: "93.184.216.34"
          comment: "web server 1"
        - value: "93.184.216.35"
          comment: "web server 2"
    - name: www
      type: CNAME
      records:
        - value: "acme-corp.com."
    - name: "@"
      type: MX
      records:
        - value: "10 mail.acme-corp.com."
        - value: "20 backup.acme-corp.com."
    - name: "@"
      type: TXT
      records:
        - value: "\"v=spf1 include:_spf.google.com ~all\""
    - name: _dmarc
      type: TXT
      records:
        - value: "\"v=DMARC1; p=reject; rua=mailto:dmarc@acme-corp.com\""
```

### Primary Zone with Cross-Component References

DNS records that reference IP addresses from other Hetzner Cloud resources using `valueFrom`. When the referenced resource's IP changes, the DNS record updates automatically.

```yaml
apiVersion: hetznercloud.openmcf.org/v1
kind: HetznerCloudDnsZone
metadata:
  name: webapp-dns
  org: acme-corp
  env: production
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: acme-corp
    pulumi.openmcf.org/project: webapp
    pulumi.openmcf.org/stack.name: production.HetznerCloudDnsZone.webapp-dns
spec:
  domainName: webapp.acme-corp.com
  mode: primary
  ttl: 3600
  recordSets:
    - name: "@"
      type: A
      ttl: 60
      records:
        - value:
            valueFrom:
              kind: HetznerCloudLoadBalancer
              name: web-lb
              fieldPath: status.outputs.ipv4_address
    - name: api
      type: A
      ttl: 60
      records:
        - value:
            valueFrom:
              kind: HetznerCloudServer
              name: api-server
              fieldPath: status.outputs.ipv4_address
    - name: www
      type: CNAME
      records:
        - value: "webapp.acme-corp.com."
    - name: "@"
      type: CAA
      records:
        - value: "0 issue \"letsencrypt.org\""
```

### Secondary Zone with TSIG Authentication

A secondary zone that synchronizes records from an external primary nameserver, authenticated with TSIG.

```yaml
apiVersion: hetznercloud.openmcf.org/v1
kind: HetznerCloudDnsZone
metadata:
  name: internal-dns
  org: acme-corp
  env: production
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: acme-corp
    pulumi.openmcf.org/project: internal
    pulumi.openmcf.org/stack.name: production.HetznerCloudDnsZone.internal-dns
spec:
  domainName: internal.acme-corp.com
  mode: secondary
  primaryNameservers:
    - address: "10.0.0.1"
      port: 53
      tsigAlgorithm: hmac-sha256
      tsigKey: "dGhpcyBpcyBhIHNlY3JldCBrZXk="
    - address: "10.0.0.2"
      port: 53
      tsigAlgorithm: hmac-sha256
      tsigKey: "dGhpcyBpcyBhIHNlY3JldCBrZXk="
```

## Stack Outputs

| Output | Type | Description |
|--------|------|-------------|
| `zone_id` | `string` | The Hetzner Cloud numeric ID of the created zone. Can be referenced by other components via `StringValueOrRef`. |
| `nameservers` | `string[]` | The authoritative Hetzner nameservers assigned to the zone (e.g., `["helium.ns.hetzner.de", "hydrogen.ns.hetzner.com", "oxygen.ns.hetzner.com"]`). Configure these NS records at your domain registrar to activate the zone. |

## Related Components

- [HetznerCloudServer](/docs/catalog/hetznercloud/hetzner-cloud-server) — A records can reference `status.outputs.ipv4_address` via `valueFrom` to point DNS at server IPs.
- [HetznerCloudFloatingIp](/docs/catalog/hetznercloud/hetzner-cloud-floating-ip) — A records can reference `status.outputs.ip_address` for failover-capable DNS.
- [HetznerCloudLoadBalancer](/docs/catalog/hetznercloud/hetzner-cloud-load-balancer) — A records can reference `status.outputs.ipv4_address` to point DNS at load balancer IPs.
- [HetznerCloudCertificate](/docs/catalog/hetznercloud/hetzner-cloud-certificate) — Managed certificates require DNS records pointing to a load balancer for the ACME HTTP-01 challenge.
