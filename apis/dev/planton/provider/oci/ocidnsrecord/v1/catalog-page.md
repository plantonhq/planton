# OCI DNS Record

Deploys an OCI DNS Record Set (RRSet) — a set of DNS resource records sharing the same domain and record type within an OCI DNS zone. Updates replace the entire record set atomically, supporting A, AAAA, CNAME, MX, TXT, SRV, CAA, NS, PTR, and other standard DNS record types.

## What Gets Created

When you deploy an OciDnsRecord resource, Planton provisions:

- **DNS Record Set** — a `dns.Rrset` resource within the target zone. Each record item carries its own rdata and TTL. The set is managed atomically — updates replace all records for the (domain, rtype) tuple.

## Prerequisites

- **OCI credentials** configured via environment variables or Planton provider config (API Key, Instance Principal, Security Token, Resource Principal, or OKE Workload Identity)
- **An OCI DNS zone** — either a zone OCID or zone name, either a literal value or a reference to an OciDnsZone resource
- **A DNS view OCID** (for private zones only) — required when referencing a private zone by name

## Quick Start

Create a file `dns-record.yaml`:

```yaml
apiVersion: oci.planton.dev/v1
kind: OciDnsRecord
metadata:
  name: app-a-record
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.OciDnsRecord.app-a-record
spec:
  zoneNameOrId:
    value: "example.com"
  domain: "app.example.com"
  rtype: "A"
  items:
    - rdata: "192.0.2.1"
      ttl: 300
```

Deploy:

```shell
planton apply -f dns-record.yaml
```

This creates an A record for `app.example.com` pointing to `192.0.2.1` with a 5-minute TTL.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `zoneNameOrId` | `StringValueOrRef` | OCID or name of the target DNS zone. ForceNew. Can reference an OciDnsZone resource via `valueFrom`. | Required |
| `domain` | `string` | Fully qualified domain name for the record set (e.g., `app.example.com`). ForceNew. | Min length 1 |
| `rtype` | `string` | DNS record type (e.g., `A`, `AAAA`, `CNAME`, `MX`, `TXT`, `SRV`, `CAA`, `NS`, `PTR`). ForceNew. | Min length 1 |
| `items` | `RecordItem[]` | DNS records in this record set. | Min 1 item |

### RecordItem

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `rdata` | `string` | Record data in type-specific presentation format. Examples: `"192.0.2.1"` (A), `"10 mail.example.com."` (MX), `"\"v=spf1 include:example.com ~all\""` (TXT). | Min length 1 |
| `ttl` | `int32` | Time to live in seconds. Controls how long resolvers cache this record. Values below 30 are not recommended by OCI. | >= 1 |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `viewId` | `StringValueOrRef` | — | OCID of the private DNS view. Required when accessing a private zone by name. Not needed when `zoneNameOrId` is an OCID. ForceNew. |

## Examples

### Single A Record

An A record pointing a subdomain to a single IP address:

```yaml
apiVersion: oci.planton.dev/v1
kind: OciDnsRecord
metadata:
  name: app-a-record
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.OciDnsRecord.app-a-record
spec:
  zoneNameOrId:
    value: "example.com"
  domain: "app.example.com"
  rtype: "A"
  items:
    - rdata: "192.0.2.1"
      ttl: 300
```

### Multiple A Records with Zone Reference

Round-robin A records using `valueFrom` to reference an OciDnsZone:

```yaml
apiVersion: oci.planton.dev/v1
kind: OciDnsRecord
metadata:
  name: web-a-records
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.OciDnsRecord.web-a-records
spec:
  zoneNameOrId:
    valueFrom:
      kind: OciDnsZone
      name: prod-zone
      fieldPath: status.outputs.zoneId
  domain: "web.example.com"
  rtype: "A"
  items:
    - rdata: "192.0.2.1"
      ttl: 300
    - rdata: "192.0.2.2"
      ttl: 300
    - rdata: "192.0.2.3"
      ttl: 300
```

### MX Records for Email

Mail exchange records with priority values embedded in rdata:

```yaml
apiVersion: oci.planton.dev/v1
kind: OciDnsRecord
metadata:
  name: mail-mx-records
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.OciDnsRecord.mail-mx-records
spec:
  zoneNameOrId:
    value: "example.com"
  domain: "example.com"
  rtype: "MX"
  items:
    - rdata: "10 mail1.example.com."
      ttl: 3600
    - rdata: "20 mail2.example.com."
      ttl: 3600
```

### CNAME Record

A CNAME alias pointing a subdomain to another hostname:

```yaml
apiVersion: oci.planton.dev/v1
kind: OciDnsRecord
metadata:
  name: api-cname
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.OciDnsRecord.api-cname
spec:
  zoneNameOrId:
    value: "example.com"
  domain: "api.example.com"
  rtype: "CNAME"
  items:
    - rdata: "lb.example.com."
      ttl: 300
```

### TXT Record for SPF

A TXT record for email sender policy:

```yaml
apiVersion: oci.planton.dev/v1
kind: OciDnsRecord
metadata:
  name: spf-txt
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.OciDnsRecord.spf-txt
spec:
  zoneNameOrId:
    value: "example.com"
  domain: "example.com"
  rtype: "TXT"
  items:
    - rdata: "\"v=spf1 include:_spf.google.com ~all\""
      ttl: 3600
```

## Stack Outputs

This component does not produce stack outputs. DNS record sets are identified by their (zone, domain, rtype) tuple, all of which are inputs.

## Related Components

- [OciDnsZone](/docs/catalog/oci/ocidnszone) — provides the zone referenced by `zoneNameOrId` via `valueFrom`
- [OciApplicationLoadBalancer](/docs/catalog/oci/ociapplicationloadbalancer) — load balancer IPs are common targets for A/CNAME records
- [OciNetworkLoadBalancer](/docs/catalog/oci/ocinetworkloadbalancer) — NLB IPs are common targets for A records
- [OciPublicIp](/docs/catalog/oci/ocipublicip) — reserved public IPs used as record targets
