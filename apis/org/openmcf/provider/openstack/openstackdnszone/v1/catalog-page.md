# OpenStack DNS Zone

Deploys an OpenStack Designate DNS zone with configurable zone type (PRIMARY or SECONDARY), SOA email, TTL, optional master nameservers for zone transfers, and optional inline DNS record sets provisioned alongside the zone.

## What Gets Created

When you deploy an OpenStackDnsZone resource, OpenMCF provisions:

- **DNS Zone** — an `openstack_dns_zone_v2` resource representing an authoritative domain in OpenStack Designate. The zone can be PRIMARY (Designate is the authoritative source) or SECONDARY (replicated from upstream master nameservers).
- **Inline DNS Record Sets** (optional) — one `openstack_dns_recordset_v2` resource per entry in the `records` list. Each record set is keyed by `recordType` + `recordName` for stable IaC state management. Supported types include A, AAAA, CNAME, MX, TXT, SRV, NS, PTR, CAA, SOA, SPF, SSHFP, and NAPTR.

## Prerequisites

- **OpenStack credentials** configured via environment variables or OpenMCF provider config
- **Designate DNS service** enabled and available in the target OpenStack project
- **A valid domain name** for the zone (e.g., `example.com`)
- **Master nameserver addresses** if creating a SECONDARY zone for zone transfers

## Quick Start

Create a file `dns-zone.yaml`:

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackDnsZone
metadata:
  name: my-zone
  labels:
    openmcf.org/provisioner: pulumi
    openmcf.org/stack.jobId: dev.OpenstackDnsZone.my-zone
    openmcf.org/stack.module.source: github.com/plantonhq/openmcf//apis/org/openmcf/provider/openstack/openstackdnszone/v1/iac/pulumi/module
spec:
  domainName: example.com
  email: admin@example.com
```

Deploy:

```shell
openmcf apply -f dns-zone.yaml
```

This creates a PRIMARY DNS zone for `example.com` in OpenStack Designate with the specified administrator email in the SOA record.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `domainName` | `string` | The DNS domain name for the zone (e.g., `example.com`). This is the authoritative domain managed by Designate. Designate may auto-append a trailing dot internally. ForceNew: changing this requires recreating the zone. | Must be a valid DNS domain (lowercase labels separated by dots, ending with a TLD) |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `email` | `string` | — | Email address of the zone administrator. Used in the SOA record for the zone. |
| `description` | `string` | — | Human-readable description of the DNS zone. |
| `ttl` | `int32` | Designate default | Default Time To Live (in seconds) for records in this zone. Determines how long resolvers cache records from this zone. |
| `type` | `string` | `PRIMARY` | The zone type. `PRIMARY` for zones where Designate is the authoritative source. `SECONDARY` for zones replicated from upstream master nameservers. ForceNew: changing this requires recreating the zone. Must be `PRIMARY` or `SECONDARY` when specified. |
| `masters` | `string[]` | `[]` | List of master nameserver addresses for SECONDARY zones. Required when `type` is `SECONDARY`. Ignored for PRIMARY zones. |
| `records` | `OpenStackDnsRecord[]` | `[]` | Inline DNS records to create alongside the zone. Each entry provisions a separate record set resource. For independently managed records, use the standalone [OpenStackDnsRecord](/docs/catalog/openstack/openstackdnsrecord) component instead. |
| `region` | `string` | provider default | Override the region from the provider config for this zone. ForceNew: changing this requires recreating the zone. |

#### Inline Record Sub-Fields (`records[]`)

Each entry in the `records` list defines a DNS record set within the zone:

| Field | Type | Default | Description | Validation |
|-------|------|---------|-------------|------------|
| `recordType` | `RecordType` | — | **(Required)** The DNS record type. Supported values: `A`, `AAAA`, `CNAME`, `MX`, `TXT`, `SRV`, `NS`, `PTR`, `CAA`, `SOA`, `SPF`, `SSHFP`, `NAPTR`. | Must be a defined enum value; cannot be `record_type_unspecified` |
| `recordName` | `string` | — | **(Required)** The fully qualified domain name for this record. Must end with a trailing dot (e.g., `www.example.com.`). Wildcard records are supported (e.g., `*.example.com.`). | Must be a valid FQDN ending with a trailing dot |
| `values` | `string[]` | — | **(Required)** The DNS record values. For A records: IPv4 addresses. For AAAA records: IPv6 addresses. For CNAME records: target hostname with trailing dot. For MX records: priority and mail server (e.g., `10 mail.example.com.`). For TXT records: text values. Multiple values create a round-robin record set. | Minimum 1 item required |
| `ttl` | `int32` | `60` | Time To Live (in seconds) for this specific record. Overrides the zone-level TTL. | — |

## Examples

### Basic Primary Zone

A minimal PRIMARY zone with an administrator email and a default TTL:

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackDnsZone
metadata:
  name: example-zone
  labels:
    openmcf.org/provisioner: pulumi
    openmcf.org/stack.jobId: dev.OpenstackDnsZone.example-zone
    openmcf.org/stack.module.source: github.com/plantonhq/openmcf//apis/org/openmcf/provider/openstack/openstackdnszone/v1/iac/pulumi/module
spec:
  domainName: example.com
  email: dns-admin@example.com
  description: Primary zone for example.com
  ttl: 3600
```

### Zone with Inline A and CNAME Records

A zone with inline records for a web application, including an A record for the apex domain and a CNAME for the `www` subdomain:

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackDnsZone
metadata:
  name: app-zone
  labels:
    openmcf.org/provisioner: pulumi
    openmcf.org/stack.jobId: dev.OpenstackDnsZone.app-zone
    openmcf.org/stack.module.source: github.com/plantonhq/openmcf//apis/org/openmcf/provider/openstack/openstackdnszone/v1/iac/pulumi/module
spec:
  domainName: app.example.com
  email: ops@example.com
  ttl: 300
  records:
    - recordType: A
      recordName: app.example.com.
      values:
        - "192.0.2.10"
        - "192.0.2.11"
      ttl: 60
    - recordType: CNAME
      recordName: www.app.example.com.
      values:
        - "app.example.com."
      ttl: 300
```

### Zone with MX and TXT Records for Email

A zone configured for email delivery with MX records pointing to mail servers and TXT records for SPF and DKIM verification:

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackDnsZone
metadata:
  name: mail-zone
  labels:
    openmcf.org/provisioner: pulumi
    openmcf.org/stack.jobId: prod.OpenstackDnsZone.mail-zone
    openmcf.org/stack.module.source: github.com/plantonhq/openmcf//apis/org/openmcf/provider/openstack/openstackdnszone/v1/iac/pulumi/module
spec:
  domainName: corp.example.com
  email: postmaster@corp.example.com
  ttl: 3600
  records:
    - recordType: MX
      recordName: corp.example.com.
      values:
        - "10 mail1.corp.example.com."
        - "20 mail2.corp.example.com."
      ttl: 3600
    - recordType: TXT
      recordName: corp.example.com.
      values:
        - "v=spf1 mx ip4:192.0.2.0/24 ~all"
      ttl: 3600
    - recordType: A
      recordName: mail1.corp.example.com.
      values:
        - "192.0.2.25"
      ttl: 300
    - recordType: A
      recordName: mail2.corp.example.com.
      values:
        - "192.0.2.26"
      ttl: 300
```

### Secondary Zone with Master Nameservers

A SECONDARY zone that replicates DNS data from upstream master nameservers via zone transfers (AXFR/IXFR):

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackDnsZone
metadata:
  name: replica-zone
  labels:
    openmcf.org/provisioner: pulumi
    openmcf.org/stack.jobId: prod.OpenstackDnsZone.replica-zone
    openmcf.org/stack.module.source: github.com/plantonhq/openmcf//apis/org/openmcf/provider/openstack/openstackdnszone/v1/iac/pulumi/module
spec:
  domainName: replicated.example.com
  description: Secondary zone replicated from upstream nameservers
  type: SECONDARY
  masters:
    - "ns1.upstream.example.com"
    - "ns2.upstream.example.com"
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `zone_id` | `string` | UUID of the created DNS zone. This is the primary output used as a foreign key by DNS record components. |
| `zone_name` | `string` | DNS zone name (derived from `domainName` in spec). This is the authoritative domain managed by this zone. |
| `region` | `string` | OpenStack region where the DNS zone was created |

## Related Components

- [OpenStackDnsRecord](/docs/catalog/openstack/openstackdnsrecord) — standalone DNS recordset component for independently managed records; use instead of inline `records` when DAG-visible dependencies are needed
- [OpenStackFloatingIp](/docs/catalog/openstack/openstackfloatingip) — floating IPs commonly referenced by A records within a DNS zone
- [OpenStackLoadBalancer](/docs/catalog/openstack/openstackloadbalancer) — load balancer VIPs commonly referenced by A or CNAME records within a DNS zone
- [OpenStackInstance](/docs/catalog/openstack/openstackinstance) — compute instances whose IPs can be published as DNS records
