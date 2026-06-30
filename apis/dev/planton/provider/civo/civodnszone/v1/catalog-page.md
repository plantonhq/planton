# Civo DNS Zone

Provisions a DNS zone (domain) on Civo Cloud with declarative DNS record management. The component creates the zone and any associated records in a single manifest, supporting A, AAAA, CNAME, MX, TXT, and other standard record types with configurable TTLs and cross-resource value references.

## What Gets Created

When you deploy a CivoDnsZone resource, Planton provisions:

- **Civo DNS Domain** — a `civo_dns_domain_name` resource representing the DNS zone for the specified domain
- **DNS Records** — one `civo_dns_domain_record` resource per value per record entry in the `records` list, linked to the created zone
- **Nameserver Delegation Info** — the authoritative nameservers (`ns0.civo.com`, `ns1.civo.com`, `ns2.civo.com`) exported as stack outputs so you can configure delegation at your registrar

## Prerequisites

- **Civo credentials** configured via environment variables or Planton provider config
- **A registered domain name** whose nameservers you can point to the Civo nameservers returned in stack outputs

## Quick Start

Create a file `civo-dns-zone.yaml`:

```yaml
apiVersion: civo.planton.dev/v1
kind: CivoDnsZone
metadata:
  name: my-dns-zone
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.CivoDnsZone.my-dns-zone
spec:
  domainName: example.com
  records:
    - name: "@"
      type: A
      values:
        - value: "93.184.216.34"
      ttlSeconds: 3600
```

Deploy:

```shell
planton apply -f civo-dns-zone.yaml
```

This creates a DNS zone for `example.com` on Civo with a single A record pointing the apex domain to an IP address. After deployment, update your domain registrar to use the Civo nameservers returned in the stack outputs.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `domainName` | `string` | The fully-qualified domain name for the DNS zone (e.g., `"example.com"`). | Required. Must match the pattern `^(?:[A-Za-z0-9-]+\.)+[A-Za-z]{2,}$`. |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `records` | `CivoDnsZoneRecord[]` | `[]` | A list of DNS records to create within the zone. Each record specifies a type, name, one or more values, and a TTL. |

#### CivoDnsZoneRecord

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `name` | `string` | — | The host/name for the DNS record, relative to the zone. Use `"@"` for apex (root) records. Required. |
| `type` | `DnsRecordType` | — | The DNS record type. Supported values: `A`, `AAAA`, `ALIAS`, `CNAME`, `MX`, `NS`, `PTR`, `SOA`, `SRV`, `TXT`, `CAA`. Required. |
| `values` | `StringValueOrRef[]` | — | One or more values for the record. Each entry is either a literal `value` string or a `valueFrom` reference to another resource's output. At least one value is required. |
| `ttlSeconds` | `uint32` | `3600` | Time-to-live in seconds. Determines how long resolvers cache the record. |

#### StringValueOrRef

Each entry in `values` is one of:

- **Literal** — `{ value: "93.184.216.34" }` provides the value directly
- **Reference** — `{ valueFrom: { kind: "...", name: "...", fieldPath: "..." } }` resolves the value from another Planton resource's stack outputs at deploy time

## Examples

### Zone with Multiple Record Types

A zone for `example.com` with A, CNAME, and MX records:

```yaml
apiVersion: civo.planton.dev/v1
kind: CivoDnsZone
metadata:
  name: example-zone
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.CivoDnsZone.example-zone
spec:
  domainName: example.com
  records:
    - name: "@"
      type: A
      values:
        - value: "93.184.216.34"
      ttlSeconds: 3600
    - name: www
      type: CNAME
      values:
        - value: "example.com"
      ttlSeconds: 3600
    - name: "@"
      type: MX
      values:
        - value: "10 mail.example.com"
        - value: "20 mail2.example.com"
      ttlSeconds: 3600
```

### Zone with TXT Records for Email Verification

A minimal zone that sets up SPF and a domain verification TXT record:

```yaml
apiVersion: civo.planton.dev/v1
kind: CivoDnsZone
metadata:
  name: verified-zone
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.CivoDnsZone.verified-zone
spec:
  domainName: verified.io
  records:
    - name: "@"
      type: TXT
      values:
        - value: "v=spf1 include:_spf.google.com ~all"
      ttlSeconds: 3600
    - name: "_dmarc"
      type: TXT
      values:
        - value: "v=DMARC1; p=reject; rua=mailto:dmarc@verified.io"
      ttlSeconds: 3600
```

### Using Foreign Key References for Record Values

Point a CNAME record at the IP address output from an Planton-managed CivoComputeInstance instead of hardcoding it:

```yaml
apiVersion: civo.planton.dev/v1
kind: CivoDnsZone
metadata:
  name: ref-zone
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.CivoDnsZone.ref-zone
spec:
  domainName: myapp.dev
  records:
    - name: "@"
      type: A
      values:
        - valueFrom:
            kind: CivoComputeInstance
            name: my-web-server
            fieldPath: status.outputs.public_ip
      ttlSeconds: 300
    - name: www
      type: CNAME
      values:
        - value: "myapp.dev"
      ttlSeconds: 3600
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `zoneName` | `string` | The domain name of the DNS zone managed on Civo |
| `zoneId` | `string` | The unique identifier (UUID) of the DNS zone, assigned by Civo |
| `nameServers` | `string[]` | The authoritative nameserver addresses for the zone (e.g., `ns0.civo.com`, `ns1.civo.com`, `ns2.civo.com`). Set these at your domain registrar to delegate DNS to Civo. |

## Related Components

- [CivoComputeInstance](/docs/catalog/civo/civocomputeinstance) — compute instances whose public IPs can be referenced as DNS record values
- [CivoKubernetesCluster](/docs/catalog/civo/civokubernetescluster) — Kubernetes clusters whose ingress IPs can be mapped to DNS records
- [CivoFirewall](/docs/catalog/civo/civofirewall) — firewalls that protect the instances behind the DNS zone
