# Scaleway DNS Zone

Deploys a Scaleway DNS zone with optional inline DNS records. The zone represents a delegated portion of the DNS namespace for a domain you own, managed through Scaleway Domains and DNS. OpenMCF provisions the zone and any inline records as a composite resource, exporting the zone name and nameservers for downstream resource references and domain registrar delegation.

## What Gets Created

When you deploy a ScalewayDnsZone resource, OpenMCF provisions:

- **DNS Zone** — a `domain.Zone` resource for the specified domain and optional subdomain prefix (e.g., `example.com` or `staging.example.com`)
- **DNS Records** (0..N) — one `domain.Record` resource per entry in the `records` list, each linked to the created zone. Records default to a 3600-second TTL if not specified.

## Prerequisites

- **Scaleway credentials** configured via environment variables or OpenMCF provider config
- **A registered domain** — Scaleway does not perform domain registration; the domain must already exist at a registrar (Namecheap, Google Domains, etc.)
- **Registrar access** — after zone creation, you must configure the nameservers returned in `status.outputs.nameServers` at your domain registrar to delegate DNS resolution to Scaleway

## Quick Start

Create a file `dns-zone.yaml`:

```yaml
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayDnsZone
metadata:
  name: my-dns-zone
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.ScalewayDnsZone.my-dns-zone
spec:
  domain: example.com
```

Deploy:

```shell
openmcf apply -f dns-zone.yaml
```

This creates a root DNS zone for `example.com` with no inline records. The zone name and nameservers are exported as stack outputs. Configure the nameservers at your domain registrar to activate DNS resolution through Scaleway.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `domain` | `string` | The registered parent domain name (e.g., `"example.com"`). Must be a domain you own or have been delegated control of. Cannot be changed after creation (forces zone recreation). | Required |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `subdomain` | `string` | `""` (root zone) | Subdomain prefix for this zone. Leave empty for the root zone. Set to a value like `"staging"` to create a zone for `staging.example.com`, enabling subdomain delegation with a separate set of nameservers. Can be updated after creation without recreating the zone. |
| `records` | `list` | `[]` (empty) | Inline DNS records to create within this zone. Each entry creates one `domain.Record` resource. Suitable for static records known at zone creation time (MX, TXT, CAA, NS). For records whose values depend on other infrastructure outputs, prefer the standalone ScalewayDnsRecord kind. |

**Record entry fields** (each item in `records`):

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `name` | `string` | `""` (zone apex) | Record name relative to the zone. Use empty string or `"@"` for the zone apex. Examples: `"www"`, `"api"`, `"_dmarc"`. |
| `type` | `enum` | — | DNS record type. Required. Supported values: `A`, `AAAA`, `ALIAS`, `CAA`, `CNAME`, `DNAME`, `MX`, `NS`, `PTR`, `SOA`, `SRV`, `TXT`, `TLSA`. |
| `data` | `StringValueOrRef` | — | Record data/value. Required. Can be a literal string or a reference to another resource's output. |
| `ttl` | `uint32` | `3600` | Time to live in seconds. Valid range: 60-2592000 (1 minute to 30 days). |
| `priority` | `uint32` | `0` | Priority for MX and SRV records. Lower values indicate higher priority. Ignored for other record types. |

## Examples

### Root Zone with No Records

A bare DNS zone for a domain, with all records managed as standalone ScalewayDnsRecord resources or by external systems:

```yaml
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayDnsZone
metadata:
  name: example-root
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.ScalewayDnsZone.example-root
spec:
  domain: example.com
```

### Subdomain Zone with MX and SPF Records

A subdomain zone for `staging.example.com` with email routing (MX) and an SPF policy (TXT):

```yaml
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayDnsZone
metadata:
  name: staging-zone
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: staging.ScalewayDnsZone.staging-zone
  env: staging
  org: acme
spec:
  domain: example.com
  subdomain: staging
  records:
    - name: ""
      type: MX
      data:
        value: "mail.example.com."
      ttl: 3600
      priority: 10
    - name: ""
      type: MX
      data:
        value: "mail2.example.com."
      ttl: 3600
      priority: 20
    - name: ""
      type: TXT
      data:
        value: "v=spf1 include:_spf.google.com ~all"
      ttl: 3600
```

### Production Zone with Multiple Record Types

A root zone for a production domain with A records, CNAME, CAA, DMARC, and mail routing:

```yaml
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayDnsZone
metadata:
  name: prod-zone
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.ScalewayDnsZone.prod-zone
  env: prod
  org: acme
spec:
  domain: acme-corp.com
  records:
    - name: ""
      type: A
      data:
        value: "203.0.113.10"
      ttl: 3600
    - name: www
      type: CNAME
      data:
        value: "acme-corp.com."
      ttl: 3600
    - name: ""
      type: MX
      data:
        value: "mail.acme-corp.com."
      ttl: 3600
      priority: 1
    - name: ""
      type: MX
      data:
        value: "mail-backup.acme-corp.com."
      ttl: 3600
      priority: 10
    - name: ""
      type: TXT
      data:
        value: "v=spf1 include:_spf.google.com ~all"
      ttl: 86400
    - name: _dmarc
      type: TXT
      data:
        value: "v=DMARC1; p=reject; rua=mailto:dmarc@acme-corp.com"
      ttl: 86400
    - name: ""
      type: CAA
      data:
        value: '0 issue "letsencrypt.org"'
      ttl: 86400
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `zoneName` | `string` | The computed zone name (`"{subdomain}.{domain}"` for subdomain zones, or `"{domain}"` for root zones). This is the primary output referenced by downstream ScalewayDnsRecord resources via `StringValueOrRef`. |
| `nameServers` | `list(string)` | Nameservers assigned by Scaleway for this zone. These must be configured at the domain registrar for DNS delegation. |
| `nameServersDefault` | `list(string)` | Scaleway's default nameservers for this zone. Usually identical to `nameServers` unless custom nameservers have been configured. |
| `nameServersMaster` | `list(string)` | Master nameservers for this zone. For standard zones, typically the same as the default nameservers. |
| `status` | `string` | Zone status in Scaleway's infrastructure (e.g., `"active"`, `"pending"`, `"error"`). |

## Related Components

- [ScalewayLoadBalancer](/docs/catalog/scaleway/scalewayloadbalancer) — provisions a Scaleway Load Balancer whose IP address can be referenced by DNS A records in this zone or via standalone ScalewayDnsRecord resources
- [ScalewayKapsuleCluster](/docs/catalog/scaleway/scalewaykapsulecluster) — deploys a managed Kubernetes cluster whose wildcard DNS endpoint can be pointed to via CNAME records
- [ScalewayInstance](/docs/catalog/scaleway/scalewayinstance) — creates compute instances whose public IPs can be mapped to A records in this zone
