---
title: "DNS Zone"
description: "DNS Zone deployment documentation"
icon: "package"
order: 100
componentName: "digitaloceandnszone"
---

# DigitalOcean DNS Zone

Deploys a DNS zone (domain) on DigitalOcean with optional inline DNS records. The component creates a `digitalocean_domain` resource for the zone itself and one `digitalocean_record` resource per value entry in the `records` list. Supported record types include A, AAAA, CNAME, MX, TXT, SRV, CAA, NS, SOA, and PTR, with type-specific fields for priority, weight, port, flags, and tag.

## What Gets Created

When you deploy a DigitalOceanDnsZone resource, Planton provisions:

- **DNS Zone (Domain)** -- a `digitalocean_domain` resource registered under the specified `domainName`
- **DNS Records** -- one `digitalocean_record` resource for each value in each entry of the `records` list; records with multiple values (e.g., round-robin A records) expand into separate DigitalOcean record resources
- **Type-Specific Attributes** -- `priority` is set for MX and SRV records; `weight` and `port` are set for SRV records; `flags` and `tag` are set for CAA records; all other record types omit these fields

## Prerequisites

- **DigitalOcean credentials** configured via environment variables or Planton provider config
- **A registered domain** at a third-party registrar (Namecheap, Google Domains, Cloudflare Registrar, etc.) with nameservers pointed to `ns1.digitalocean.com`, `ns2.digitalocean.com`, and `ns3.digitalocean.com`
- **DNSSEC disabled** at the registrar before delegating to DigitalOcean nameservers (DigitalOcean does not support DNSSEC; leaving it enabled will cause resolution failures)

## Quick Start

Create a file `dns-zone.yaml`:

```yaml
apiVersion: digital-ocean.planton.dev/v1
kind: DigitalOceanDnsZone
metadata:
  name: my-zone
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.DigitalOceanDnsZone.my-zone
spec:
  domainName: example.com
```

Deploy:

```shell
planton apply -f dns-zone.yaml
```

This creates a DNS zone for `example.com` on DigitalOcean with no inline records. DigitalOcean automatically provisions default NS and SOA records for the zone.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `domainName` | `string` | The fully qualified domain name for the DNS zone (e.g., `example.com`). | Required; must match pattern `^(?:[A-Za-z0-9-]+\.)+[A-Za-z]{2,}$` |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `records` | `DigitalOceanDnsZoneRecord[]` | `[]` | List of DNS records to create within the zone. |
| `records[].name` | `string` | -- | Hostname relative to the zone. Use `@` for the zone apex. Required per record. |
| `records[].type` | `DnsRecordType` | -- | DNS record type: `A`, `AAAA`, `CNAME`, `MX`, `TXT`, `SRV`, `CAA`, `NS`, `SOA`, `PTR`. Required per record. |
| `records[].values` | `StringValueOrRef[]` | -- | One or more record values. Each entry is either a literal `value` string or a `valueFrom` reference to another resource's output. Required per record (min 1). |
| `records[].ttlSeconds` | `uint32` | `3600` | Time-to-live in seconds. Controls how long resolvers cache this record. |
| `records[].priority` | `uint32` | `0` | Priority for MX and SRV records. Lower values indicate higher priority. Ignored for other types. |
| `records[].weight` | `uint32` | `0` | Relative weight for SRV records with the same priority. Ignored for non-SRV types. |
| `records[].port` | `uint32` | `0` | TCP/UDP port for SRV records. Ignored for non-SRV types. |
| `records[].flags` | `uint32` | `0` | Flags for CAA records (`0` = non-critical, `128` = critical). Ignored for non-CAA types. |
| `records[].tag` | `string` | `""` | Tag for CAA records (`issue`, `issuewild`, or `iodef`). Ignored for non-CAA types. |

## Examples

### Minimal Zone (Domain Only)

Register a domain on DigitalOcean DNS with no inline records. DigitalOcean creates default NS and SOA records automatically.

```yaml
apiVersion: digital-ocean.planton.dev/v1
kind: DigitalOceanDnsZone
metadata:
  name: bare-zone
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.DigitalOceanDnsZone.bare-zone
spec:
  domainName: bare-zone.dev
```

### Zone with A Records

A zone with apex and `www` A records pointing to the same IP address, plus a CNAME for the `api` subdomain:

```yaml
apiVersion: digital-ocean.planton.dev/v1
kind: DigitalOceanDnsZone
metadata:
  name: web-app
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: staging.DigitalOceanDnsZone.web-app
spec:
  domainName: web-app.io
  records:
    - name: "@"
      type: A
      values:
        - value: "203.0.113.10"
      ttlSeconds: 3600
    - name: "www"
      type: A
      values:
        - value: "203.0.113.10"
      ttlSeconds: 3600
    - name: "api"
      type: CNAME
      values:
        - value: "lb.web-app.io."
      ttlSeconds: 300
```

### Zone with Mixed Record Types (A, CNAME, MX, TXT, CAA)

A production zone with web records, Google Workspace MX, SPF/DMARC TXT records, and a CAA policy restricting certificate issuance to Let's Encrypt:

```yaml
apiVersion: digital-ocean.planton.dev/v1
kind: DigitalOceanDnsZone
metadata:
  name: prod-zone
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.DigitalOceanDnsZone.prod-zone
spec:
  domainName: prod-app.com
  records:
    # Apex A record pointing to a load balancer
    - name: "@"
      type: A
      values:
        - value: "198.51.100.1"
      ttlSeconds: 300
    # www CNAME to apex
    - name: "www"
      type: CNAME
      values:
        - value: "@"
      ttlSeconds: 3600
    # Google Workspace MX records
    - name: "@"
      type: MX
      values:
        - value: "aspmx.l.google.com."
      ttlSeconds: 3600
      priority: 1
    - name: "@"
      type: MX
      values:
        - value: "alt1.aspmx.l.google.com."
      ttlSeconds: 3600
      priority: 5
    - name: "@"
      type: MX
      values:
        - value: "alt2.aspmx.l.google.com."
      ttlSeconds: 3600
      priority: 5
    # SPF record
    - name: "@"
      type: TXT
      values:
        - value: "v=spf1 include:_spf.google.com ~all"
      ttlSeconds: 3600
    # DMARC policy
    - name: "_dmarc"
      type: TXT
      values:
        - value: "v=DMARC1; p=reject; rua=mailto:dmarc@prod-app.com"
      ttlSeconds: 3600
    # CAA -- only Let's Encrypt may issue certificates
    - name: "@"
      type: CAA
      values:
        - value: "letsencrypt.org"
      ttlSeconds: 3600
      flags: 0
      tag: issue
    # CAA -- deny wildcard issuance
    - name: "@"
      type: CAA
      values:
        - value: ";"
      ttlSeconds: 3600
      flags: 0
      tag: issuewild
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `zone_name` | `string` | The domain name of the created DNS zone (e.g., `example.com`) |
| `zone_id` | `string` | The unique identifier of the created DNS zone assigned by DigitalOcean |
| `name_servers` | `string[]` | The list of DigitalOcean nameservers for this zone (`ns1.digitalocean.com`, `ns2.digitalocean.com`, `ns3.digitalocean.com`) |

## Related Components

- [DigitalOceanDroplet](/docs/catalog/digitalocean/droplet) -- provides compute instances whose IPs can be referenced in A records
- [DigitalOceanLoadBalancer](/docs/catalog/digitalocean/load-balancer) -- provisions load balancers whose IPs can be used as record targets
- [DigitalOceanVpc](/docs/catalog/digitalocean/vpc) -- defines the network for infrastructure that DNS records resolve to
- [DigitalOceanKubernetesCluster](/docs/catalog/digitalocean/kubernetes-cluster) -- deploys Kubernetes clusters that can use ExternalDNS to manage records in this zone
